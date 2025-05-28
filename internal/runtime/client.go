// Copyright 2025 The Zeus Authors.
// Licensed under the MIT License. See the LICENSE file for details.

package runtime

import (
	"github.com/docker/docker/client"
	"github.com/raphaeldichler/zeus/internal/assert"
)

var (
	c *client.Client = nil
)

func init() {
	cli, err := client.NewClientWithOpts(
		client.WithAPIVersionNegotiation(),
	)
	assert.ErrNil(err)
	c = cli
}
