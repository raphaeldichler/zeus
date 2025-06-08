// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package runtime

import (
	"context"

	"github.com/docker/docker/client"
	"github.com/raphaeldichler/zeus/internal/assert"
)

const NetworkDaemonName = "network"

func networkName(applicaiton string) string {
	assert.StartsNotWith(applicaiton, '/', "applications cannot start with '/'")
	assert.IsAsciiString(applicaiton, "application can only contain ascii chars")
	return "zeus/network/" + applicaiton
}

type Network struct {
	id     string
	client *client.Client
	name   string
}

// Interact with the docker daemon and initialises a new network
//
// The network gets labels with:
//   - zeus.object.type=network
//   - zeus.application.name={application}
func CreateNewNetwork(
	application string,
) (*Network, error) {
	assert.NotNil(c, "init of docker-client failed")

	networkName := networkName(application)
	networkId, err := createBridgedNetwork(application, networkName)
	if err != nil {
		return nil, err
	}

	return newNetwork(networkId, networkName), nil
}

func newNetwork(
	id string,
	name string,
) *Network {
	return &Network{
		id:     id,
		client: c,
		name:   name,
	}
}

func (self *Network) Cleanup() error {
	ctx := context.Background()
	return self.client.NetworkRemove(ctx, self.id)
}
