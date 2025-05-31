// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package nginxcontroller

import (
	"net"
	"net/http"
	"syscall"
)

const (
  SocketPath string = "/etc/zeus/ingress/nginx.sock"
)

type Controller struct {
  listen net.Listener
  nginxPath string
}

func NewServer(
  nginxPath string,
) (*Controller, error) {
  listen, err := net.Listen("unix", SocketPath)
  if err != nil {
    return nil, err
  }

  return &Controller{
    listen: listen,
    nginxPath: nginxPath,
  }, nil
}


func (self *Controller) Run() error {
  http.HandleFunc("/apply", self.Apply)
  http.HandleFunc("/acme", self.Acme)

  defer self.listen.Close()
  return http.Serve(self.listen, nil)
}

func (self *Controller) ReloadNginxConfig(configPath string) error {

  return syscall.Kill(1, syscall.SIGHUP)
}

func (self *Controller) Acme(w http.ResponseWriter, r *http.Request) {

}




