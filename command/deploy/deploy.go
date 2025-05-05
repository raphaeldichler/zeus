package deploy

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
	"gopkg.in/yaml.v3"
)

type AuthConfig struct {
	Auths map[string]registry.AuthConfig `json:"auths"`
}

type Config struct {
	Version  string   `yaml:"version"`
	Metadata Metadata `yaml:"metadata"`
	Container Container `yaml:"container"`
}

type Metadata struct {
	Name string `yaml:"name"`
}

type Container struct {
	Image   string   `yaml:"image"`
	Ports   []string `yaml:"ports"`
	Labels  []string `yaml:"labels"`
	Volumes []string `yaml:"volumes"`
}

type DeployContext struct {
  ApplicationName string
}

func (self *DeployContext) NetworkLabel() string {
  return fmt.Sprintf("zeus.%s.nw", self.ApplicationName)
}

func (self *DeployContext) ServiceLabel() string {
  return fmt.Sprintf("zeus.%s.svc.name", self.ApplicationName)
}


func loadConfig(pathDeployFile string) (*Config, error) {
	data, err := os.ReadFile(pathDeployFile)
	if err != nil {
		log.Fatalf("error reading file: %v", err)
	}

  config := new(Config)
	if err = yaml.Unmarshal(data, config); err != nil {
    return nil, err
	}

  return config, nil
}

func deployService(cli *client.Client, cfg *Config) error {
  // what do we need to do?
  // 1) check if service and network exists
  // if network doesnt exist, we create a new one
  deployCtx := &DeployContext{
    ApplicationName: "poseidon",
  }

  if err := pullImage(
    "registry.io/poseidon/foobra:v1.0",
    cli,
  ); err != nil {
    return err
  }

  networkID, err := ensureNetwork(
    deployCtx,
    cli,
  )


  serviceID, err := tryFindServiceByName(
    *deployCtx,
    cli,
    cfg.Metadata.Name,
  )
  if err != nil {
    return err
  }


  fmt.Printf("ServiceID: %s\n", serviceID)
  fmt.Printf("NetworkID: %s\n", networkID)
  return nil
}

func pullImage(
  refStr string,
  cli *client.Client,
) error {
  ctx := context.Background()
  
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("failed to get home directory: %v", err)
	}

	configPath := filepath.Join(home, ".docker", "config.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("error reading file: %v", err)
	}

  config := new(AuthConfig)
	if err = yaml.Unmarshal(data, config); err != nil {
    return err
	}

  auths, ok := config.Auths["some.domain.com"]
  if !ok {
    fmt.Println("failed to find registry")
    return nil
  }
  auth, err := registry.EncodeAuthConfig(auths)


  reader, err := cli.ImagePull(
    ctx,
    refStr,
    image.PullOptions{
      RegistryAuth: auth,

    },
  )
  if err != nil {
    return err
  }
  defer reader.Close()

  return nil
}

func tryFindServiceByName(
  deployCtx DeployContext,
  cli *client.Client,
  serviceName string,
) (string, error) {
  ctx := context.Background()

  list, err := cli.ContainerList(ctx, container.ListOptions{})
  if err != nil {
    return "", err
  }

  for _, container := range list {
    key, ok := container.Labels[deployCtx.ServiceLabel()]
    if ok && key == serviceName {
        return container.ID, nil
    }
  }

  return "", nil
}

func ensureNetwork(
  deployCtx *DeployContext,
  cli *client.Client,
) (string, error) {
  networkID, err := tryFindNetworkByName(
    *deployCtx,
    cli,
  )
  if err != nil {
    return "", err
  }

  if networkID == "" {
    networkID, err = createNetwork(
      deployCtx,
      cli,
    )
    if err != nil {
      return "", err
    }
  }

  return networkID, nil
}

func tryFindNetworkByName(
  deployCtx DeployContext,
  cli *client.Client,
) (string, error) {
  ctx := context.Background()

  list, err := cli.NetworkList(ctx, network.ListOptions{})
  if err != nil {
    return "", err
  }

  for _, network := range list {
    key, ok := network.Labels[deployCtx.NetworkLabel()]
    if ok {
        if key != "" {
          // TODO: log warning, that the label has invalid labeling
        }
        return network.ID, nil
    }
  }

  return "", nil
}

func createNetwork(
  deployCtx *DeployContext,
  cli *client.Client,
) (string, error) {
  ctx := context.Background()

  networkName := fmt.Sprintf("%d", rand.Int())
  network, err := cli.NetworkCreate(
    ctx, 
    networkName,
    network.CreateOptions{
      Driver: "bridge",
      Labels: map[string]string{
        deployCtx.NetworkLabel(): "", 
      },
    },
  )
  if err != nil {
    return "", err
  }

  return network.ID, nil
}
