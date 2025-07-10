// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package runtime

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/raphaeldichler/zeus/internal/assert"
	"github.com/raphaeldichler/zeus/internal/util/optional"
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

func TrySelectOneContainer(
	application string,
	labels ...Label,
) (optional.Optional[Container], error) {
	selectedContainers, err := SelectContainer(labels...)
	if err != nil {
		return optional.Empty[Container](), err
	}

	switch len(selectedContainers) {
	case 0:
		return optional.Empty[Container](), nil

	case 1:
		c, err := selectedContainers[0].NewContainer(application)
		if err != nil {
		  return optional.Empty[Container](), err
		}
		return optional.Of(c), nil

	default:
		assert.Unreachable(
			"Too many container exists in the current context. Possible external tampering.",
		)
	}

	return optional.Empty[Container](), nil
}

func SelectAllNonApplicationContainers(
	application string,
) ([]*Container, error) {
	ctx := context.Background()
	args := filters.NewArgs(filters.Arg("label", labelApplicationName))
	containers, err := c.ContainerList(ctx, container.ListOptions{
		Filters: args,
	})
	if err != nil {
		return nil, err
	}

	var result []*Container = nil
	for _, cont := range containers {
		applicationLabel := cont.Labels[labelApplicationName]
		if applicationLabel == application {
			continue
		}

		selected := &SelectedContainer{id: cont.ID}
		c, err := selected.NewContainer(applicationLabel)
		if err != nil {
			return nil, err
		}

		result = append(result, c)
	}

	return result, nil
}
