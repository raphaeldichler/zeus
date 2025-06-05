// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package server

import (
	"net"
	"net/http"
)

type HttpRequest interface {
	Validate(w http.ResponseWriter, r *http.Request) bool
}

type Controller struct {
	Method string
	Path   string
	Run    func(w http.ResponseWriter, r *http.Request)
}

type ServerConfig struct {
	Listener    net.Listener
	Controllers []Controller
}

func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		Listener:    nil,
		Controllers: make([]Controller, 0),
	}
}

type ServerOption func(cfg *ServerConfig)

type ControllerFunction[T any] func(w http.ResponseWriter, r *http.Request, command *T)

func WithListener(
	listener net.Listener,
) ServerOption {
	return func(cfg *ServerConfig) {
		cfg.Listener = listener
	}
}

func Post[T any](
	path string,
	controller ControllerFunction[T],
	opts ...HandlerOption[T],
) ServerOption {
	return handle("POST", path, controller, opts...)
}

func Get[T any](
	path string,
	controller ControllerFunction[T],
	opts ...HandlerOption[T],
) ServerOption {
	return handle("GET", path, controller, opts...)
}

func Delete[T any](
	path string,
	controller ControllerFunction[T],
	opts ...HandlerOption[T],
) ServerOption {
	return handle("DELETE", path, controller, opts...)
}

func Put[T any](
	path string,
	controller ControllerFunction[T],
	opts ...HandlerOption[T],
) ServerOption {
	return handle("PUT", path, controller, opts...)
}

func handle[T any](
	method string,
	path string,
	controller ControllerFunction[T],
	opts ...HandlerOption[T],
) ServerOption {
	hc := DefaultHandlerConfig[T]()
	for _, opt := range opts {
		opt(hc)
	}

	return func(cfg *ServerConfig) {
		ctr := Controller{
			Method: method,
			Path:   path,
			Run: func(w http.ResponseWriter, r *http.Request) {
				req := new(T)

				if err := hc.Decoder(w, r, req); err != nil {
					return
				}

				if ok := hc.Validation(w, r, *req); !ok {
					return
				}

				controller(w, r, req)
			},
		}

		cfg.Controllers = append(cfg.Controllers, ctr)
	}
}
