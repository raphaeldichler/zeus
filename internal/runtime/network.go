// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package runtime

import (
	"context"

	"github.com/docker/docker/client"
	"github.com/raphaeldichler/zeus/internal/assert"
	"github.com/raphaeldichler/zeus/internal/log"
)

func networkName(applicaiton string) string {
	assert.StartsNotWith(applicaiton, '/', "applications cannot start with '/'")
	assert.IsAsciiString(applicaiton, "application can only contain ascii chars")
	return "zeus/network/" + applicaiton
}

type NetworkIdentifier interface {
	NetworkName() string
}

type Network struct {
	id          string
	client      *client.Client
	networkName string

	log *log.Logger
}

func NewNetwork(
	application string,
	daemon string,
) (*Network, error) {
	assert.NotNil(c, "init of docker-client failed")

	networkName := networkName(application)
	networkId, err := createBridgedNetwork(networkName)
	if err != nil {
		return nil, err
	}

	return &Network{
		id:          networkId,
		client:      c,
		networkName: networkName,
		log:         log.New(application, daemon),
	}, err
}

func (self *Network) NetworkName() string {
	return self.networkName
}

func (self *Network) Cleanup() error {
	ctx := context.Background()
	return self.client.NetworkRemove(ctx, self.id)
}
