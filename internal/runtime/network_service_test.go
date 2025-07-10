// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package runtime

import "testing"

func TestServiceAddressRing(t *testing.T) {
	ring := newServiceAddressRing([3]uint8{10, 10, 10})

	if ring.next() != "10.10.10.0" {
		t.Error("first address is not correct")
	}

	if ring.next() != "10.10.10.1" {
		t.Error("second address is not correct")
	}

	ring.returnIP(0)
	if ring.next() != "10.10.10.0" {
		t.Error("first address is not correct")
	}
}
