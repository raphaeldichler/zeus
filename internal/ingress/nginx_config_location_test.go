// Copyright 2025 The Zeus Authors.
// Licensed under the MIT License. See the LICENSE file for details.

package ingress

import (
	"testing"
)

func TestLocationIdentifierExactDifferPrefix(t *testing.T) {
	prefixLocationID := LocationIdentifier{
		Path:     "/foo",
		Matching: LocationPrefix,
	}

	exactLocationID := LocationIdentifier{
		Path:     "/foo",
		Matching: LocationExact,
	}

	prefixConfigFilename := prefixLocationID.FilePath()
	exactConfigFilename := exactLocationID.FilePath()
	if prefixConfigFilename == exactConfigFilename {
		t.Errorf("prefix and exact should be different, but they match")
	}
}
