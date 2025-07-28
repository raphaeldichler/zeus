// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package dnscontroller

import (
	"context"
	"math/big"
	"net"
	"os"

	"github.com/raphaeldichler/zeus/internal/util/assert"
	log "github.com/raphaeldichler/zeus/internal/util/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	SocketPath = "/run/zeus/dns.sock"
)

type Controller struct {
	UnimplementedDNSControllerServer
	server   *grpc.Server
	listener net.Listener
	log      *log.Logger

	networkPart uint8
	plg         *ZeusDns

	internalDNS *dnsEntryState
	externalDNS *dnsEntryState
}

func New(dnsPlugin *ZeusDns) (*Controller, error) {
	if _, err := os.Stat(SocketPath); err == nil {
		if err := os.Remove(SocketPath); err != nil {
			return nil, err
		}
	}

	listen, err := net.Listen("unix", SocketPath)
	if err != nil {
		return nil, err
	}

	if err := os.Chmod(SocketPath, 0666); err != nil {
		return nil, err
	}

	s := grpc.NewServer()
	srv := &Controller{
		server:      s,
		listener:    listen,
		log:         log.New("dns", "controller"),
		networkPart: 0,
		plg:         dnsPlugin,
		internalDNS: newDNSEntryState(0),
		externalDNS: newDNSEntryState(100),
	}
	RegisterDNSControllerServer(s, srv)

	return srv, nil
}

func (self *Controller) Run() error {
	return self.server.Serve(self.listener)
}

func (c *Controller) SetDNSEntry(
	ctx context.Context,
	req *DNSSetRequest,
) (*DNSSetResponse, error) {
	networkPart := networkHashToIpPart(req.NetworkHash)
	if c.networkPart == 0 {
		c.networkPart = networkPart
	}

	if c.networkPart != networkPart {
		return nil, status.Error(codes.Unknown, "network part must not change")
	}

	var (
		internal []string = nil
		external []string = nil
	)
	for _, e := range req.Entries {
		switch e.Type {
		case DNSEntryType_Internal:
			internal = append(internal, e.Domain)
		case DNSEntryType_External:
			external = append(external, e.Domain)
		}
	}

	c.internalDNS.update(internal)
	c.externalDNS.update(external)

	ipMap := make(map[string]string)
	ipMap = c.internalDNS.appendTo(ipMap)
	ipMap = c.externalDNS.appendTo(ipMap)
	c.plg.setIpMap(ipMap)

	return c.toResponse(), nil
}

func (c *Controller) toResponse() *DNSSetResponse {
	r := new(DNSSetResponse)

	for _, entries := range []map[string]string{
		c.internalDNS.entries,
		c.externalDNS.entries,
	} {
		for domain, ip := range entries {
			e := &DNSEntry{
				Domain: domain,
				IP:     ip,
			}
			r.DNSEntries = append(r.DNSEntries, e)
		}

	}

	return r
}

func networkHashToIpPart(networkHash string) uint8 {
	n := new(big.Int)
	n.SetString(networkHash, 16)

	mod := new(big.Int)
	mod.Mod(n, big.NewInt(254))

	ipPart := mod.Int64() + 1
	assert.True(ipPart <= 254, "ip part must be in range [1, 254]")

	return uint8(ipPart)
}
