// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package server

import (
	"encoding/json"
	"io"
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

type DecoderFunc[T HttpRequest] func(r io.Reader, out *T) error

type HandlerConfig[T HttpRequest] struct {
	Decoder DecoderFunc[T]
}

func DefaultHandlerConfig[T HttpRequest]() *HandlerConfig[T] {
	return &HandlerConfig[T]{
		Decoder: func(r io.Reader, out *T) error {
			return json.NewDecoder(r).Decode(out)
		},
	}
}

func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		Listener:    nil,
		Controllers: make([]Controller, 0),
	}
}

type ServerOption func(cfg *ServerConfig)

type HandlerOption[T HttpRequest] func(cfg *HandlerConfig[T])

type ControllerFunction[T HttpRequest] func(w http.ResponseWriter, r *http.Request, command *T)

func Listen(
	listener net.Listener,
) ServerOption {
	return func(cfg *ServerConfig) {
		cfg.Listener = listener
	}
}

func Post[T HttpRequest](
	path string,
	controller ControllerFunction[T],
	opts ...HandlerOption[T],
) ServerOption {
	return handle("POST", path, controller, opts...)
}

func Get[T HttpRequest](
	path string,
	controller ControllerFunction[T],
	opts ...HandlerOption[T],
) ServerOption {
	return handle("GET", path, controller, opts...)
}

func Delete[T HttpRequest](
	path string,
	controller ControllerFunction[T],
	opts ...HandlerOption[T],
) ServerOption {
	return handle("DELETE", path, controller, opts...)
}

func Put[T HttpRequest](
	path string,
	controller ControllerFunction[T],
	opts ...HandlerOption[T],
) ServerOption {
	return handle("PUT", path, controller, opts...)
}

func handle[T HttpRequest](
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

				if err := hc.Decoder(r.Body, req); err != nil {
					return
				}

				if ok := (*req).Validate(w, r); !ok {
					return
				}

				controller(w, r, req)
			},
		}

		cfg.Controllers = append(cfg.Controllers, ctr)
	}
}

func Decoder[T HttpRequest](f DecoderFunc[T]) HandlerOption[T] {
	return func(cfg *HandlerConfig[T]) {
		cfg.Decoder = f
	}
}
