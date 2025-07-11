// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package nginxcontroller

import (
	"github.com/raphaeldichler/zeus/internal/util/assert"
)

type IngressRequestBuilder struct {
	servers []*Server
	general []string
	event   []string
	http    []string
}

type ServerRequestBuilder struct {
	request *Server
}

func NewIngressRequestBuilder() *IngressRequestBuilder {
	return &IngressRequestBuilder{}
}

func (i *IngressRequestBuilder) AddGeneralEntries(entries ...string) {
	i.general = append(i.general, entries...)
}

func (i *IngressRequestBuilder) AddEventEntries(entries ...string) {
	i.event = append(i.event, entries...)
}

func (i *IngressRequestBuilder) AddHttpEntries(entries ...string) {
	i.http = append(i.http, entries...)
}

func (i *IngressRequestBuilder) AddServer(
	domain string,
	ipv6 bool,
	entries ...string,
) *ServerRequestBuilder {
	request := &Server{
		Domain:    domain,
		Tls:       nil,
		Entries:   entries,
		Locations: nil,
		IPv6:      ipv6,
	}
	i.servers = append(i.servers, request)

	return &ServerRequestBuilder{request: request}
}

func (i *IngressRequestBuilder) Build() *IngressRequest {
	return &IngressRequest{
		GeneralEntries: i.general,
		EventEntries:   i.event,
		HttpEntries:    i.http,
		Servers:        i.servers,
	}
}

func (s *ServerRequestBuilder) AddTLS(
	fullchain string,
	privkey string,
) {
	assert.True(fullchain != "", "fullchain must exists")
	assert.True(privkey != "", "fullchain must exists")

	s.request.Tls = &TLS{
		Fullchain: fullchain,
		Privkey:   privkey,
	}
}

func (s *ServerRequestBuilder) AddLocation(
	path string,
	matching Matching,
	entries ...string,
) {
	s.request.Locations = append(
		s.request.Locations,
		&Location{
			Path:     path,
			Matching: matching,
			Entries:  entries,
		},
	)
}
