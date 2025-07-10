// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package optional

import (
	"testing"
)

func TestEmptyOptional(t *testing.T) {
	o := Empty[int]()
	if o.IsPresent() {
		t.Errorf("expected IsPresent to be false on empty optional, got true")
	}
}

func TestOfOptional(t *testing.T) {
	val := 42
	o := Of(&val)
	if !o.IsPresent() {
		t.Errorf("expected IsPresent to be true on optional with value, got false")
	}

	got := o.Get()
	if got == nil {
		t.Errorf("expected Get to return non-nil pointer, got nil")
	} else if *got != val {
		t.Errorf("expected Get to return %d, got %d", val, *got)
	}
}
