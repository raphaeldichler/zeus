package command

import (
	"fmt"
  "errors"
	"os"
	"path/filepath"

	"github.com/docker/cli/cli/connhelper"
	"github.com/docker/docker/client"
	"github.com/raphaeldichler/zeus/command/assert"
	"gopkg.in/yaml.v3"
)

const (
  DefaultConfigPath = "~/.zeus/config.yml"
  defaultPath = "$HOME/.zeus/config.yml"
  ZeusConfigEnv = "ZEUSCONFIG"
)

type ZeusConfig struct {
  Application string `yaml:"application"`
  IP string `yaml:"ip"`
  Ssh SSH `yaml:"ssh"`
}

type SSH struct {
  User string `yaml:"user"`
  Cert string `yaml:"cert"`
}

type Command struct {
  Application string
  Cli   *client.Client
}


func New(configPath string) (*Command, error) {
  var zeusconfigPath string = defaultPath

  if zeusconfigPathEnv := os.Getenv(ZeusConfigEnv); zeusconfigPathEnv != "" {
    zeusconfigPath = zeusconfigPathEnv
  }

  if configPath != "" {
    zeusconfigPath = configPath
  }

  zeusconfigPath, err := filepath.Abs(zeusconfigPath)
  if err != nil {
    return nil, err
  }

  cfg, err := loadZeusConfig(zeusconfigPath)
  if err != nil {
    return nil, err
  }
  

  connection, err := connhelper.GetConnectionHelperWithSSHOpts(
    fmt.Sprintf("ssh://%s@%s", cfg.Ssh.User, cfg.IP),
    []string{},
  )
  if err != nil {
    return nil, err
  }

  cli, err := client.NewClientWithOpts(
    client.WithDialContext(connection.Dialer),
    client.WithAPIVersionNegotiation(),
  )

  return &Command {
    Application: cfg.Application,
    Cli: cli,
  }, nil
}

func loadZeusConfig(pathToConfig string) (*ZeusConfig, error) {
  assert.True(filepath.IsAbs(pathToConfig), "cannot read data, if path is not abs")
  if _, err := os.Stat(pathToConfig); errors.Is(err, os.ErrNotExist) {
    assert.Unreachable("cannot read data, if file doesnt exists")
  }

	data, err := os.ReadFile(pathToConfig)
  if err != nil {
    return nil, err
  }

  cfg := new(ZeusConfig)
	if err = yaml.Unmarshal(data, cfg); err != nil {
    return nil, err
	}

  return cfg, nil
}
