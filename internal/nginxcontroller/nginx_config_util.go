// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package nginxcontroller

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"github.com/raphaeldichler/zeus/internal/assert"
)

type ErrorResponse struct {
	ErrorType    string `json:"error_type"`
	ErrorMessage string `json:"error_message"`
	DebugInfo    string `json:"debug_info,omitempty"`
	Details      any    `json:"details,omitempty"`
}

func reply(
	w http.ResponseWriter,
	status int,
	errorResponse *ErrorResponse,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(errorResponse)
}

func replyInternalServerError(
	w http.ResponseWriter,
	message string,
) {
	reply(
		w,
		http.StatusInternalServerError,
		&ErrorResponse{
			ErrorType:    "internal-error",
			ErrorMessage: message,
		},
	)
}

func replyBadRequest(
	w http.ResponseWriter,
	errorResponse *ErrorResponse,
) {
	reply(
		w,
		http.StatusBadRequest,
		errorResponse,
	)
}

type directory string

func openDirectory() (directory, error) {
	tmp, err := os.MkdirTemp("", "")
	if err != nil {
		return "", err
	}

	return directory(tmp), nil
}

func (d directory) close() error {
	return os.RemoveAll(string(d))
}

func (d directory) store(filename string, content []byte) (string, error) {
	path := filepath.Join(string(d), filename)
	return path, os.WriteFile(path, content, 0600)
}

func (d directory) storeFile(ext string, content []byte) (string, error) {
	assert.StartsNotWith(ext, '.', "the method appends a '.' to the filename")

	b := make([]byte, 16)
	rand.Read(b)
	filename := hex.EncodeToString(b) + "." + ext

	return d.store(filename, content)
}

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

func (self *Set[T]) size() int {
	return len(self.arr)
}
