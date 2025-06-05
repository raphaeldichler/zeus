// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package server

import (
	"encoding/json"
	"net/http"
)

type HandlerOption[T any] func(cfg *HandlerConfig[T])

type (
	DecoderFunc[T any]  func(w http.ResponseWriter, r *http.Request, out *T) error
	ValidateFunc[T any] func(w http.ResponseWriter, r *http.Request, req T) (ok bool)
)

type HandlerConfig[T any] struct {
	Decoder    DecoderFunc[T]
	Validation ValidateFunc[T]
}

func DefaultHandlerConfig[T any]() *HandlerConfig[T] {
	return &HandlerConfig[T]{
		Decoder: func(w http.ResponseWriter, r *http.Request, out *T) error {
			return json.NewDecoder(r.Body).Decode(out)
		},
		Validation: func(w http.ResponseWriter, r *http.Request, req T) (ok bool) {
			return true
		},
	}
}

func WithRequestDecoder[T any](f DecoderFunc[T]) HandlerOption[T] {
	return func(cfg *HandlerConfig[T]) {
		cfg.Decoder = f
	}
}

func WithRequestValidation[T any](f ValidateFunc[T]) HandlerOption[T] {
	return func(cfg *HandlerConfig[T]) {
		cfg.Validation = f
	}
}
