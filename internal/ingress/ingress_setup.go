// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package ingress

import (
	"os"

	"github.com/raphaeldichler/zeus/internal/nginxcontroller"
)

// Sets up all nessesary files and directories on the host system
func Setup() error {
	// ensure that the mount directory of ingress controller exists
	if err := os.MkdirAll(nginxcontroller.HostSocketDirectory(), 0700); err != nil {
		return err
	}

	return nil
}
