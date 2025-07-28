// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package dnscontroller

import (
	"context"
	"math/big"
	"net"

	"github.com/raphaeldichler/zeus/internal/util/assert"
	log "github.com/raphaeldichler/zeus/internal/util/logger"
	"github.com/raphaeldichler/zeus/internal/util/socket"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var SocketFileEnvironmentManager = socket.NewFileEnvironmentManager(
  "/run/zeus/dns",
  "dns.sock",
  0666,
)

const (
  internalDNSIdentifier uint8 = 0
  externalDNSIdentifier uint8 = 100
)

type Controller struct {
	UnimplementedDNSControllerServer
	server   *grpc.Server
	listener net.Listener
	log      *log.Logger

	plg *ZeusDns

  networkHash string
	internalDNS *dnsEntryState
	externalDNS *dnsEntryState
}

func New(
  dnsPlugin *ZeusDns,
  networkHash string,
) (*Controller, error) {
  listen, err := SocketFileEnvironmentManager.Listen()
  if err != nil {
    return nil, err
  }

  network := networkHashToIpPart(networkHash)
	s := grpc.NewServer()
	srv := &Controller{
		server:      s,
		listener:    listen,
		log:         log.New("dns", "controller"),
		plg:         dnsPlugin,
    networkHash: networkHash,
		internalDNS: newDNSEntryState(network, internalDNSIdentifier),
		externalDNS: newDNSEntryState(network, externalDNSIdentifier),
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
  if req.NetworkHash != c.networkHash {
    return nil, status.Error(codes.Unknown, "network part differs from controller")
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
