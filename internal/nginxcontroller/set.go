// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package nginxcontroller

type Compare[T any] interface {
	Equal(other *T) bool
}

type Set[T Compare[T]] struct {
	arr []*T
}

func NewSet[T Compare[T]]() Set[T] {
	return Set[T]{
		arr: make([]*T, 0),
	}
}

func (self *Set[T]) remove(other *T) *T {
	for idx, loc := range self.arr {
		if (*loc).Equal(other) {
			loc := self.arr[idx]
			self.arr[idx] = self.arr[len(self.arr)-1]
			self.arr = self.arr[:len(self.arr)-1]

			return loc
		}
	}

	return nil
}

func (self *Set[T]) set(other *T) {
	for idx, e := range self.arr {
		if (*other).Equal(e) {
			self.arr[idx] = other
			return
		}
	}

	self.arr = append(self.arr, other)
}

func (self *Set[T]) entries() []*T {
	return self.arr
}

func (self *Set[T]) get(other *T) *T {
	for _, e := range self.arr {
		if (*other).Equal(e) {
			return e
		}
	}

	return nil
}
