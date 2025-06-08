// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package ingress

import (
	"testing"
	"time"

	"github.com/raphaeldichler/zeus/internal/assert"
	"github.com/raphaeldichler/zeus/internal/record"
	"github.com/raphaeldichler/zeus/internal/runtime"
)

func TestIngressContainerSelectAndCreate(t *testing.T) {
	application := "testing-select-create-" + ID()
	image := buildIngressContainer(application)

	var (
		network *runtime.Network   = nil
		c       *runtime.Container = nil
	)
	defer func() {
		if c != nil {
			c.Shutdown()
		}
		if network != nil {
			network.Cleanup()
		}
	}()

	network, err := runtime.CreateNewNetwork(application)
	assert.ErrNil(err)
	assert.NotNil(network, "must create network")

	state := record.ApplicationRecord{
		Ingress: record.NewIngressRecord(),
	}

	state.Ingress.Metadata.Image = image
	state.Ingress.Metadata.Name = application
	state.Ingress.Metadata.CreateTime = time.Now()

	c = SelectOrCreateIngressContainer(&state, application, nil)
	if !state.Ingress.Errors.NoErrors() {
		t.Fatalf("wanted to get no errors, but got '%v'", state)
	}
	if c == nil {
		t.Fatalf("select or create should return a non nil value")
	}

	c1 := SelectOrCreateIngressContainer(&state, application, nil)
	if !c.Equal(c1) {
		t.Errorf("reselecting must return same container, but was not")
	}
}
