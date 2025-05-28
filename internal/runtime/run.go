// Copyright 2025 The Zeus Authors.
// Licensed under the MIT License. See the LICENSE file for details.

package runtime

import (
	"context"
	"io"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/raphaeldichler/zeus/internal/assert"
)

func pull(
	imageRef string,
) error {
	ctx := context.Background()
	_, err := c.ImageInspect(
		ctx,
		imageRef,
	)
	if err == nil {
		return nil
	}

	r, err := c.ImagePull(
		ctx,
		imageRef,
		image.PullOptions{},
	)
	if err != nil {
		return err
	}

	_, err = io.Copy(io.Discard, r)
	assert.ErrNil(err)

	err = r.Close()
	assert.ErrNil(err)

	return nil
}

func create(
	cfg *container.Config,
	hostCfg *container.HostConfig,
	networkCfg *network.NetworkingConfig,
) (string, error) {
	ctx := context.Background()

	cont, err := c.ContainerCreate(
		ctx, cfg, hostCfg, networkCfg, nil, "",
	)
	if err != nil {
		panic(err)
	}

	return cont.ID, nil
}

func start(
	containerID string,
	retry int,
) error {
	var err error = nil
	ctx := context.Background()
	for range retry {
		err = c.ContainerStart(ctx, containerID, container.StartOptions{})
		if err == nil {
			return nil
		}
	}

	return ErrContainterCannotStart
}

func existsContaienr(
	containerID string,
) (bool, error) {
	ctx := context.Background()
	summary, err := c.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return false, err
	}

	for _, item := range summary {
		if item.ID == containerID {
			return true, nil
		}
	}

	return false, nil
}

func createBridgedNetwork(
	networkName string,
) (string, error) {
	ctx := context.Background()
	created, err := c.NetworkCreate(
		ctx,
		networkName,
		network.CreateOptions{
			Labels: map[string]string{
				"zeus.object.type": "network",
			},
		},
	)
	if err != nil {
		return "", err
	}

	return created.ID, nil
}
