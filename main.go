package main

import (
	"context"
	"fmt"

	"github.com/docker/cli/cli/connhelper"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

func main() {
  conn, err := connhelper.GetConnectionHelper("ssh://azureuser@51.116.100.5")
  if err != nil {
    panic(err)
  }

  cli, err := client.NewClientWithOpts(
    client.WithDialContext(conn.Dialer),
    client.WithAPIVersionNegotiation(),
  )
	if err != nil {
		panic(err)
	}
	defer cli.Close()

  resp, err := cli.ContainerCreate(
    context.Background(),
    &container.Config{
      Image: "essaymentor.azurecr.io/poseidon/test:155",
    },
    &container.HostConfig{
      AutoRemove: true,
    },
    nil, 
    nil, 
    "my-remoted-golang-container-1",
  )
	if err != nil {
		panic(err)
	}

  if err := cli.ContainerStart(
    context.Background(),
    resp.ID,
    container.StartOptions{},
  ); err != nil {
    panic(err)
  }



	containers, err := cli.ContainerList(context.Background(), container.ListOptions{})
	if err != nil {
		panic(err)
	}

	for _, ctr := range containers {
		fmt.Printf("%s %s (status: %s)\n", ctr.ID, ctr.Image, ctr.Status)
	}
}
