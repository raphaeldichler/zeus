// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package nginxcontroller

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"

	"github.com/raphaeldichler/zeus/internal/assert"
)

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
