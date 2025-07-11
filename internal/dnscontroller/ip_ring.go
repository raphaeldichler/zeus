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
type ipRing struct {
	openParts   [255]uint8
	networkPart uint8
	idx         uint8
}

func newIpRing(networkPart uint8) *ipRing {
	ring := new(ipRing)

	ring.networkPart = networkPart
	ring.idx = 0
	for i := 0; i < 255; i++ {
		ring.openParts[i] = 0
	}

	return ring
}

func (i *ipRing) next() string {
	next := i.openParts[i.idx]
	i.idx += 1
	assert.True(i.idx <= 255, "idx must not exceed 255")

	return fmt.Sprintf("10.%d.10.%d", i.networkPart, next)
}

// Returns the ip back to the pool, if it was not given away via next() the function returns false.
func (i *ipRing) returnToPool(ip string) bool {
	returnedIpParts := strings.Split(ip, ".")
	assert.True(len(returnedIpParts) == 4, "ip must be in the form of '10.x.10.y'")
	assert.True(returnedIpParts[0] == "10", "ip must be in the form of '10.x.10.y'")
	assert.True(returnedIpParts[1] == fmt.Sprintf("%d", i.networkPart), "ip must be in the form of '10.x.10.y'")
	assert.True(returnedIpParts[2] == "10", "ip must be in the form of '10.x.10.y'")

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
