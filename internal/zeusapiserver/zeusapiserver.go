// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package zeusapiserver

import (
	"net"
	"os"

	"github.com/raphaeldichler/zeus/internal/server"
)

const SocketPath = "/run/zeus/zeusctl.sock"

type ZeusController struct {
	Applications map[string]string
	Server       *server.HttpServer
}

func New() (*ZeusController, error) {
	if _, err := os.Stat(SocketPath); err == nil {
		if err := os.Remove(SocketPath); err != nil {
			return nil, err
		}
	}

	listen, err := net.Listen("unix", SocketPath)
	if err != nil {
		return nil, err
	}

	self := &ZeusController{
		Applications: make(map[string]string),
	}
	self.Server = server.New(
		server.WithListener(listen),
		// Ingress
		server.Get(
			IngressInspectAPIPath,
			self.GetIngressInspect,
			server.WithRequestDecoder(GetIngressInspectRequestDecoder),
			server.WithRequestValidation(self.GetIngressInspectRequestValidation),
		),
		server.Post(
			IngressApplyAPIPath,
			self.PostIngressApply,
			server.WithRequestDecoder(PostIngressApplyRequestDecoder),
			server.WithRequestValidation(self.PostIngressApplyRequestValidator),
		),
	)

	return self, nil
}

func (self *ZeusController) Run() error {
	return self.Server.Run()
}
