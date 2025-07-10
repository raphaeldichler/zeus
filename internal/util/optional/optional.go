// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package optional

import "github.com/raphaeldichler/zeus/internal/assert"

type Optional[T any] struct {
	t *T
}

func Empty[T any]() Optional[T] {
	return Optional[T]{t: nil}
}

func Of[T any](t *T) Optional[T] {
	return Optional[T]{t: t}
}

func (o Optional[T]) IsPresent() bool {
	return o.t != nil
}

func (o Optional[T]) IsEmpty() bool {
	return o.t == nil
}

func (o Optional[T]) IfPresent(f func (*T) (*T)) Optional[T] {
  if o.IsEmpty() {
    return o
  }

  return Of(f(o.t))
}

func (o Optional[T]) Get() *T {
	assert.NotNil(o.t, "get of optional requires it to be present")
	return o.t
}
