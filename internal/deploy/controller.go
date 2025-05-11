package deploy

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/labstack/echo/v4"
	"github.com/raphaeldichler/zeus/internal/reply"
	"github.com/sirupsen/logrus"
)


type DeployMetadata struct {
  Name string `json:"name"`
}

type DeployContainer struct {
  Image string `json:"image"`
}

type DeployRequest struct {
  Application string `param:"application"`
  Metadata DeployMetadata `json:"metadata"`
  Container DeployContainer `json:"container"`
  Secret map[string]string `json:"secret"`
}


// /v1.0/applications/:application/services/deploy
func PostDeployController(ctx echo.Context) error {
  cmd := new(DeployRequest)
  if err := ctx.Bind(cmd); err != nil {
    return reply.BadRequest(ctx, reply.Err("invalid_request", "idk"))
  }

  
  return nil
  //deploy(cmd, cli)
}

func isImagePullAllowed(image string) bool {
  return true
  //strings.HasPrefix(image, "essaymentor.azurecr.io") 
}

var (
  ErrInvalidImageRegistry = errors.New("image cannot be pulled: registry not allowed")
)

func TarGzFromString(filename, content string) (io.Reader, error) {
	var buf bytes.Buffer

	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	hdr := &tar.Header{
		Name:    filename,
		Mode:    0600,
		Size:    int64(len(content)),
		ModTime: time.Now(),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return nil, err
	}

	if _, err := tw.Write([]byte(content)); err != nil {
		return nil, err
	}

	if err := tw.Close(); err != nil {
		return nil, err
	}
	if err := gw.Close(); err != nil {
		return nil, err
	}

	return bytes.NewReader(buf.Bytes()), nil
}

func Deploy(command *DeployRequest, client *client.Client) (bool, reply.ErrorResponse) {
  if !isImagePullAllowed("nginx") {
    return false, reply.Err(
      "invalid_registry",
      "image cannot be pulled: registry not allowed",
    )
  }


  ctx := context.Background()

  logrus.Infoln("Ensure image was downloaded")
  _, err := client.ImageInspect(
    ctx,
    "nginx",
  )
  if err != nil {
    // log: could not found image locally pull from repository
    logrus.Infoln(err)

    r, err := client.ImagePull(
      ctx,
      "nginx",
      image.PullOptions{},
    )
    if err != nil {
      panic(err)
    }
    _, err = io.Copy(io.Discard, r)
    if err != nil {
      panic(err)
    }

    r.Close()
  }
  logrus.Infoln("Image exists?")


  /*
  ?) pull image? do we need 
  1) create container
  2) copy secrets into container
  */

  c, err := client.ContainerCreate(
    ctx,
    &container.Config{
      Image: "nginx",
    },
    &container.HostConfig{

    },
    &network.NetworkingConfig{

    },
    nil,
    "",
  )
  if err != nil {
    panic(err)
  }

  tar, err := TarGzFromString("/run/secret/my_secret", "foobra")
  if err != nil {
    panic(err)
  }
  if err := client.CopyToContainer(
    ctx,
    c.ID,
    "/",
    tar,
    container.CopyToContainerOptions{},
  ); err != nil {
    panic(err)
  }

  if err := client.ContainerStart(
    ctx,
    c.ID,
    container.StartOptions{},
  ); err != nil {
    panic(err)
  }




  return true, reply.Err("", "")
}


