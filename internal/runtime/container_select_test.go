// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package runtime

import (
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/raphaeldichler/zeus/internal/util/assert"
)

func TestRuntimeSelectContainer(t *testing.T) {
	testID := rand.IntN(1000000)
	application := fmt.Sprintf("testing-%d", testID)

	labels := []Label{
		ObjectTypeLabel(IngressObject),
		ApplicationNameLabel(application),
	}

	container, err := CreateNewContainer(
		application,
		WithImage("traefik/whoami:latest"),
		WithPulling(),
		WithLabels(labels...),
	)
	assert.ErrNil(err)
	defer container.Shutdown()

	selected, err := SelectContainer(labels...)
	assert.ErrNil(err)

	if len(selected) != 1 {
		t.Errorf(
			"failed to select correct container, expected to get 1, but got %d", len(selected),
		)
	}

	if selected[0].id != container.id {
		t.Errorf(
			"failed to select correct network, expected id to be '%s', but got %s", container.id, selected[0].id,
		)
	}
}

func TestRuntimeSelectContainerWithNetwork(t *testing.T) {
	testID := rand.IntN(1000000)
	application := fmt.Sprintf("testing-%d", testID)
	network, err := CreateNewNetwork(application)
	assert.ErrNil(err)
	fmt.Println("created network ", network)

	labels := []Label{
		ObjectTypeLabel(IngressObject),
		ApplicationNameLabel(application),
	}

	container, err := CreateNewContainer(
		application,
		WithImage("traefik/whoami:latest"),
		WithPulling(),
		WithLabels(labels...),
		WithConnectedToNetwork(network),
	)
	assert.ErrNil(err)
	defer container.Shutdown()
	fmt.Println("created container ", container)

	selected, err := SelectContainer(labels...)
	assert.ErrNil(err)

	if len(selected) != 1 {
		t.Errorf(
			"failed to select correct container, expected to get 1, but got %d", len(selected),
		)
	}

	selectedContainer, err := selected[0].NewContainer(application)
	assert.ErrNil(err)

	if !selectedContainer.Equal(container) {
		t.Errorf("wrong selected container, want '%v' got '%v'", container, selectedContainer)
	}
}
