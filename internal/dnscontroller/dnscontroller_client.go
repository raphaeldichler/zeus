// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package dnscontroller

import (
	"github.com/raphaeldichler/zeus/internal/util/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	DNSControllerClient
	conn *grpc.ClientConn
}

func NewClient() *Client {
	conn, err := grpc.NewClient(
		"unix:////zeus/dns/dns.sock",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	// the configuration of the client should be correct
	// currently, we dont see any other reason that this should fail
	assert.ErrNil(err)

	client := NewDNSControllerClient(conn)
	return &Client{
		DNSControllerClient: client,
		conn:                conn,
	}
}

func (self *Client) Close() error {
	return self.conn.Close()
}
