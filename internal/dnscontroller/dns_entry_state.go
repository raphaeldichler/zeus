// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package dnscontroller

import (
	"slices"

	"github.com/raphaeldichler/zeus/internal/util/assert"
)


type dnsEntryState struct {
	// maps the domain to the ip address which the DNS is pointing to
	entries map[string]string
	// the ip ring which is used to assign ip addresses to the domains
	ring *ipRing
}

func newDNSEntryState(networkPart uint8) *dnsEntryState {
	return &dnsEntryState{
		entries: make(map[string]string),
		ring:    newIpRing(networkPart),
	}
}

func (d *dnsEntryState) update(domains []string) {
	// to reduce the actual update which are need to be done, we dont change domains ip addresses
	// if they are already set. else we would have the possiblity that we reassign a domain a different ip
	for domain, ip := range d.entries {
		if !slices.Contains(domains, domain) {
			ok := d.ring.returnToPool(ip)
			assert.True(ok, "ip must be returned to the pool")

			delete(d.entries, domain)
		}
	}

	for _, domain := range domains {
		_, ok := d.entries[domain]
		if ok {
			continue
		}

		ip := d.ring.next()
		d.entries[domain] = ip
	}
}

func (d *dnsEntryState) appendTo(ipMap map[string]string) map[string]string {
	for domain, ip := range d.entries {
		_, ok := ipMap[domain]
		assert.False(ok, "domain must not be in ipMap")
		ipMap[domain] = ip
	}

	return ipMap
}
