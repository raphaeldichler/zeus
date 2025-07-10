// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package nginxcontroller

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/raphaeldichler/zeus/internal/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	DefaultClientTimeout = 5 * time.Second
)

type Client struct {
	NginxControllerClient
	conn *grpc.ClientConn
}

func NewClient(
	application string,
) *Client {
	conn, err := grpc.NewClient(
		fmt.Sprintf("unix://%s", filepath.Join(HostSocketDirectory(), "nginx.sock")),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	// the configuration of the client should be correct
	// currently, we dont see any other reason that this should fail
	assert.ErrNil(err)

	client := NewNginxControllerClient(conn)
	return &Client{
		NginxControllerClient: client,
		conn:                  conn,
	}
}

func (self *Client) Close() error {
	return self.conn.Close()
}
