// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package runtime

import (
	"fmt"

	"github.com/raphaeldichler/zeus/internal/util/assert"
)

type serviceAddressRing struct {
	index       int
	baseAddress [3]uint8
	ring        [256]uint8
}

func newServiceAddressRing(baseAddress [3]uint8) *serviceAddressRing {
	s := &serviceAddressRing{
		index:       0,
		baseAddress: baseAddress,
	}

	for i := range 256 {
		s.ring[i] = uint8(i)
	}

	return s
}

func (s *serviceAddressRing) next() string {
	part := s.ring[s.index]
	s.index += 1
	assert.True(s.index < len(s.ring), "ring overflow")

	return fmt.Sprintf(
		"%d.%d.%d.%d",
		s.baseAddress[0],
		s.baseAddress[1],
		s.baseAddress[2],
		part,
	)
}

func (s *serviceAddressRing) returnIP(part uint8) {
	assert.True(s.index > 0, "index ring underflow")

	returnIdx := -1
	for idx := range s.index {
		if s.ring[idx] == part {
			returnIdx = idx
			break
		}
	}
	assert.True(returnIdx != -1, "part not found in ring")

	s.index -= 1
	s.ring[returnIdx] = s.ring[s.index]
	s.ring[s.index] = part

}
