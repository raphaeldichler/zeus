// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package socket

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
)

// A specialied enviroment manager which is capable of managing socket files
// and perform specific operation on it.
type FileEnvironmentManager struct {
	base       string
	socketFile string
	mode       os.FileMode
}

func NewFileEnvironmentManager(
	basePath string,
	socketFile string,
	mode os.FileMode,
) FileEnvironmentManager {
	return FileEnvironmentManager{
		base:       basePath,
		socketFile: socketFile,
		mode:       mode,
	}
}

func (f FileEnvironmentManager) Setup() error {
	return os.MkdirAll(f.base, f.mode)
}

func (f FileEnvironmentManager) SocketPath() string {
	return filepath.Join(f.base, f.socketFile)
}

func (f FileEnvironmentManager) UnixSocketURI() string {
	return fmt.Sprintf("unix:///%s", f.SocketPath())
}

func (f FileEnvironmentManager) Listen() (net.Listener, error) {
	if _, err := os.Stat(f.SocketPath()); err == nil {
		if err := os.Remove(f.SocketPath()); err != nil {
			return nil, err
		}
	}

	listen, err := net.Listen("unix", f.SocketPath())
	if err != nil {
		return nil, err
	}

	if err := os.Chmod(f.SocketPath(), f.mode); err != nil {
		return nil, err
	}

	return listen, nil
}
