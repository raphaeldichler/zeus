// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package nginxcontroller

import (
	"net"
	"net/http"
	"os/exec"
)

const (
  SocketPath string = "/etc/zeus/ingress/nginx.sock"
)

type Controller struct {
  listen net.Listener
  nginx string
  config *NginxConfig
}

func NewServer() (*Controller, error) {
  nginx, err := exec.LookPath("nginx")
  if err != nil {
    return nil, err
  }

  listen, err := net.Listen("unix", SocketPath)
  if err != nil {
    return nil, err
  }

  return &Controller{
    listen: listen,
    nginx: nginx,
    config: NewNginxConfig(),
  }, nil
}


func (self *Controller) Run() error {
  http.HandleFunc("/apply", self.Apply)
  http.HandleFunc("/acme", self.Acme)

  defer self.listen.Close()
  return http.Serve(self.listen, nil)
}

func (self *Controller) ReloadNginxConfig(configPath string) error {
  cmd := exec.Command(self.nginx, "-s", "reload", "-c", configPath)
  return cmd.Run()
}
