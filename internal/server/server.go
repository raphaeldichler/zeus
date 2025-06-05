// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

// YAF -> Yet Another Framework
package server

import (
	"net"
	"net/http"
)

type HttpServer struct {
	Config   *ServerConfig
	Listener net.Listener
}

func New(
	opts ...ServerOption,
) *HttpServer {
	cfg := DefaultServerConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	return &HttpServer{
		Config:   cfg,
		Listener: cfg.Listener,
	}
}

func (self *HttpServer) Run() error {
	mux := http.NewServeMux()
	self.Config.setupControllers(mux)

	defer self.Listener.Close()
	return http.Serve(self.Listener, mux)
}
