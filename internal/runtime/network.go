// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package runtime

import (
	"context"
	"fmt"

	"github.com/docker/docker/client"
	"github.com/raphaeldichler/zeus/internal/dnscontroller"
	"github.com/raphaeldichler/zeus/internal/util/assert"
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

  dns *Container
  dnsClient dnscontroller.Client
}

// Interact with the docker daemon and initialises a new network. Additionally start a DNS sever inside the network.
//
// The network gets labeled with:
//   - zeus.object.type=network
//   - zeus.application.name={application}
// The DNS gets labeled with:
//   - zeus.object.type=dns
//   - zeus.application.name={applicaiton}
func CreateNewNetwork(
	application string,
) (*Network, error) {
	assert.NotNil(c, "init of docker-client failed")

	networkName := networkName(application)
	networkId, err := createBridgedNetwork(application, networkName)
	if err != nil {
		return nil, err
	}

  network := newNetwork(networkId, networkName)

  dnsContainer, err := CreateNewContainer(
    application,
    WithImage("coredns:v1"),
    WithConnectedToNetwork(network),
    WithLabels(
      ObjectTypeLabel(DNSObject),
      ApplicationNameLabel(application),
    ),
    WithMount("/run/zeus/", "/run/zeus/"),
  )
  fmt.Println(dnsContainer.id)
  if err != nil {
    if err := network.Cleanup(); err != nil {
      // also the network cleanup can fail, for recovery
    }

    return nil, err
  }
  dnsClient := dnscontroller.NewClient()

  network.dns = dnsContainer
  network.dnsClient = *dnsClient

  return network, nil
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

func (self *Network) String() string {
	return self.name
}

func (self *Network) Cleanup() error {
	ctx := context.Background()
	return self.client.NetworkRemove(ctx, self.id)
}
