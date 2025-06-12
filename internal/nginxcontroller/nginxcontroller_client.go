// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package nginxcontroller

import (
	"path/filepath"
	"time"

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
) (*Client, error) {

	socket := filepath.Join(HostSocketDirectory(application), "nginx.sock")
	conn, err := grpc.NewClient(
		"unix://"+socket,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	client := NewNginxControllerClient(conn)

	return &Client{
		NginxControllerClient: client,
		conn:                  conn,
	}, nil
}

func (self *Client) CLose() error {
	return self.conn.Close()
}
