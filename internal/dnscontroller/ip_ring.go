// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package dnscontroller

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/raphaeldichler/zeus/internal/util/assert"
)

// ipRing is a simple ring buffer to store the ip addresses which can be used to set the dns entries.
// 10.network.identifier.ring
type ipRing struct {
	openParts   [255]uint8
	network uint8
  identifier  uint8
	idx         uint8
}

func newIpRing(
  network uint8,
  identifier uint8,
) *ipRing {
	ring := new(ipRing)

	ring.network = network
  ring.identifier = identifier
	ring.idx = 0
  for i := range 255 {
		ring.openParts[i] = uint8(i)
	}

	return ring
}

func (i *ipRing) next() string {
	next := i.openParts[i.idx]
	i.idx += 1
	assert.True(i.idx <= 255, "idx must not exceed 255")

	return fmt.Sprintf("10.%d.%d.%d", i.network, i.identifier, next)
}

// Returns the ip back to the pool, if it was not given away via next() the function returns false.
func (i *ipRing) returnToPool(ip string) bool {
	returnedIpParts := strings.Split(ip, ".")
	assert.True(len(returnedIpParts) == 4, "ip must be in the form of '10.network.identifier.y'")

	returnedPart, err := strconv.Atoi(returnedIpParts[3])
	assert.ErrNil(err)
	assert.True(returnedPart >= 0 && returnedPart <= 254, "ip part must be in range [1, 254]")

	var (
		found bool  = false
		idx   uint8 = 0
	)
	for ; idx < i.idx; idx++ {
		if i.openParts[idx] == uint8(returnedPart) {
			found = true
			break
		}
	}
	if !found {
		return false
	}

	i.idx -= 1
	i.openParts[idx] = i.openParts[i.idx]
	i.openParts[i.idx] = uint8(returnedPart)

	return true
}

// Returns true if num elements are available.
func (i *ipRing) hasOpen(num int) bool {
	return (len(i.openParts) - int(i.idx)) <= num
}
