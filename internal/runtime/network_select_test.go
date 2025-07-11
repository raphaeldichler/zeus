// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package runtime

import (
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/raphaeldichler/zeus/internal/util/assert"
)

func TestRuntimeSelectNetwork(t *testing.T) {
	testID := rand.IntN(1000000)

	application := fmt.Sprintf("testing-%d", testID)
	network, err := CreateNewNetwork(application)
	assert.ErrNil(err)
	defer network.Cleanup()

	selected, err := SelectNetworks(
		ApplicationNameLabel(application),
		ObjectTypeLabel(NetworkObject),
	)
	assert.ErrNil(err)

	if len(selected) != 1 {
		t.Errorf(
			"failed to select correct network, expected to get 1, but got %d", len(selected),
		)
	}

	if selected[0].name != network.name {
		t.Errorf(
			"failed to select correct network, expected name to be '%s', but got %s", network.name, selected[0].name,
		)
	}
}
