// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package runtime

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/raphaeldichler/zeus/internal/util/assert"
)

type SelectedNetwork struct {
	name string
	id   string
}

func (self *SelectedNetwork) NewNetwork(
	application string,
) *Network {
	networkName := networkName(application)
	assert.True(networkName == self.name, "selecting should use the correct labels")

	return newNetwork(self.id, networkName)
}

// Selects a container by the labels if it exists. No promise about the container is made,
// it can be in any state.
//
// If not container exists nil is returned.
func SelectNetworks(
	labels ...Label,
) ([]SelectedNetwork, error) {
	assert.NotNil(c, "init of docker-client failed")

	args := filters.NewArgs()
	for _, l := range labels {
		args.Add("label", fmt.Sprintf("%s=%s", l.key, l.value))
	}

	ctx := context.Background()
	summary, err := c.NetworkList(
		ctx, network.ListOptions{
			Filters: args,
		},
	)
	if err != nil {
		return nil, err
	}

	var result []SelectedNetwork = nil
	for _, e := range summary {
		result = append(result, SelectedNetwork{name: e.Name, id: e.ID})
	}

	return result, nil
}

func TrySelectApplicationNetwork(
	application string,
) (*Network, error) {

	networks, err := SelectNetworks(
		ObjectTypeLabel(NetworkObject),
		ApplicationNameLabel(application),
	)
	if err != nil {
		return nil, err
	}

	switch len(networks) {
	case 0:
		return nil, nil

	case 1:
		return networks[0].NewNetwork(application), nil

	default:
		assert.Unreachable("Either 0 or 1 networks must be selected")
	}

	return nil, nil
}

func SelectAllNonApplicationNetworks(
	application string,
) ([]*Network, error) {
	ctx := context.Background()
	args := filters.NewArgs(filters.Arg("label", labelApplicationName))
	networks, err := c.NetworkList(ctx, network.ListOptions{
		Filters: args,
	})
	if err != nil {
		return nil, err
	}

	var result []*Network = nil
	for _, nw := range networks {
		applicationLabel := nw.Labels[labelApplicationName]
		if applicationLabel == application {
			continue
		}

		selected := &SelectedNetwork{name: nw.Name, id: nw.ID}
		network := selected.NewNetwork(applicationLabel)

		result = append(result, network)
	}

	return result, nil
}
