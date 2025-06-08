// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package runtime

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/raphaeldichler/zeus/internal/assert"
)

type SelectedContainer struct {
	id string
}

func (self *SelectedContainer) NewContainer(
	application string,
) (*Container, error) {
	networks, err := SelectNetworks(
		ObjectTypeLabel(NetworkObject),
		ApplicationNameLabel(application),
	)
	if err != nil {
		return nil, err
	}

	var network *Network = nil
	switch len(networks) {
	case 1:
		network = networks[0].NewNetwork(application)

	default:
		assert.Unreachable("Network must exists, is created on application start")
	}

	return toContainer(application, self.id, network), nil
}

// Selects a container by the labels if it exists. No promise about the container is made,
// it can be in any state.
func SelectContainer(
	labels ...Label,
) ([]SelectedContainer, error) {
	assert.NotNil(c, "init of docker-client failed")

	args := filters.NewArgs()
	for _, l := range labels {
		args.Add("label", fmt.Sprintf("%s=%s", l.key, l.value))
	}

	ctx := context.Background()
	summary, err := c.ContainerList(
		ctx, container.ListOptions{
			Filters: args,
		},
	)
	if err != nil {
		return nil, err
	}

	var result []SelectedContainer = nil
	for _, e := range summary {
		result = append(result, SelectedContainer{id: e.ID})
	}

	return result, nil
}
