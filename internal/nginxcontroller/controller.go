// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package nginxcontroller

import (
	"net"
	"net/http"
	"os"
	"os/exec"

	"github.com/raphaeldichler/zeus/internal/server"
)

const (
	NginxConfigPath = "/etc/nginx/nginx.conf"
	SocketPath      = "/run/zeus/nginx.sock"
)

type Controller struct {
	server *server.HttpServer
	nginx  string
	config *NginxConfig
}

func NewServer() (*Controller, error) {
	nginx, err := exec.LookPath("nginx")
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(SocketPath); err == nil {
		if err := os.Remove(SocketPath); err != nil {
			return nil, err
		}
	}

	listen, err := net.Listen("unix", SocketPath)
	if err != nil {
		return nil, err
	}

	self := &Controller{
		server: nil,
		nginx:  nginx,
		config: NewNginxConfig(),
	}
	self.server = server.New(
		server.Listen(listen),
		server.Post("/apply", self.Apply),
		server.Post("/acme", self.SetAcme),
		server.Delete("/acme", self.DeleteAcme),
	)

	return self, nil
}

func (self *Controller) SetConfig(cfg *NginxConfig) {
	self.config = cfg
}

func (self *Controller) Run() error {
	return self.server.Run()
}

func (self *Controller) ReloadNginxConfig() error {
	cmd := exec.Command(self.nginx, "-s", "reload")
	return cmd.Run()
}

func (self *Controller) StoreAndApplyConfig(
	w http.ResponseWriter,
	cfg *NginxConfig,
	d directory,
) error {
	err := cfg.Store(d)
	if err != nil {
		replyInternalServerError(w, "Failed to store nginx config. "+err.Error())
	}

	if err := self.ReloadNginxConfig(); err != nil {
		replyInternalServerError(w, "Failed to reload nginx config. "+err.Error())
		return err
	}

	return nil
}
