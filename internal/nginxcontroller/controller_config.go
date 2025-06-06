// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package nginxcontroller

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/raphaeldichler/zeus/internal/assert"
)

var (
	matching map[string]MatchingType = map[string]MatchingType{
		"exact":  ExactMatching,
		"prefix": PrefixMatching,
	}
)

type ApplyRequest struct {
	Servers []ServerRequest `json:"servers"`
}

func NewApplyRequest() *ApplyRequest {
	return &ApplyRequest{
		Servers: make([]ServerRequest, 0),
	}
}

type ServerRequestOption func(cfg *ServerRequest)

type ServerRequestOptions struct {
	Options []ServerRequestOption
}

func NewServerRequestOptions() *ServerRequestOptions {
	return &ServerRequestOptions{
		Options: nil,
	}
}

func (self *ServerRequestOptions) Add(opt ...ServerRequestOption) {
	self.Options = append(self.Options, opt...)
}

func WithCertificate(
	privkeyPem string,
	fullchainPem string,
) ServerRequestOption {
	return func(cfg *ServerRequest) {
		cfg.Certificate = &CertificateRequest{
			PrivkeyPem:   privkeyPem,
			FullchainPem: fullchainPem,
		}
	}
}

func WithLocation(
	path string,
	matching string,
	serviceEndpoint string,
) ServerRequestOption {
	return func(cfg *ServerRequest) {
		cfg.Locations = append(cfg.Locations, LocationRequest{
			Path:            path,
			Matching:        matching,
			ServiceEndpoint: serviceEndpoint,
		})
	}
}

func WithDomain(domain string) ServerRequestOption {
	return func(cfg *ServerRequest) {
		cfg.Domain = domain
	}
}

func WithIPv6Enabled(ipv6Enabled bool) ServerRequestOption {
	return func(cfg *ServerRequest) {
		cfg.IPv6Enabled = ipv6Enabled
	}
}

func (self *ApplyRequest) AddServer(opts ...ServerRequestOption) {
	server := new(ServerRequest)
	for _, opt := range opts {
		opt(server)
	}

	self.Servers = append(self.Servers, *server)
}

type ServerRequest struct {
	Domain      string              `json:"domain"`
	Certificate *CertificateRequest `json:"certificate"`
	Locations   []LocationRequest   `json:"locations"`
	IPv6Enabled bool                `json:"ipv6Enabled"`
}

type LocationRequest struct {
	Path            string `json:"path"`
	Matching        string `json:"matching"`
	ServiceEndpoint string `json:"serviceEndpoint"`
}

type CertificateRequest struct {
	PrivkeyPem   string `json:"privkeyPem"`
	FullchainPem string `json:"fullchainPem"`
}

func ValidateApplyRequest(
	w http.ResponseWriter,
	r *http.Request,
	req ApplyRequest,
) bool {
	return true
}

func (self *CertificateRequest) ToCertificate() *TlsCertificate {
	fullchainPem, err := base64.StdEncoding.DecodeString(self.FullchainPem)
	assert.ErrNil(err)
	privkeyPem, err := base64.StdEncoding.DecodeString(self.PrivkeyPem)
	assert.ErrNil(err)

	return &TlsCertificate{
		FullchainFilePath: "",
		PrivkeyFilePath:   "",
		Fullchain:         fullchainPem,
		Privkey:           privkeyPem,
	}
}

func (self *Controller) Apply(
	w http.ResponseWriter,
	r *http.Request,
	command *ApplyRequest,
) {
	d, err := openDirectory()
	if err != nil {
		replyInternalServerError(w, "Failed to open directory to store data. "+err.Error())
		return
	}
	defer d.close()
	cfg := NewNginxConfig()

	for _, server := range command.Servers {
		var tls *TlsCertificate = nil
		if cert := server.Certificate; cert != nil {
			tls = cert.ToCertificate()
		}

		sc := NewServerConfig(
			server.Domain,
			server.IPv6Enabled,
			tls,
		)
		for _, loc := range server.Locations {
			m, ok := matching[loc.Matching]
			assert.True(ok, "matching type must already be validated")

			lc := NewLocationConfig(
				loc.Path,
				m,
				fmt.Sprintf(`return 200 "%s@%s"`, server.Domain, loc.Path),
				"add_header Content-Type text/plain",
				/*
					"proxy_pass "+loc.ServiceEndpoint,
					"proxy_set_header Host $host",
					"proxy_set_header X-Real-IP $remote_addr",
					"proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for",
					"proxy_set_header X-Forwarded-Proto $scheme",
				*/
			)

			sc.SetLocation(lc)
		}

		cfg.SetServerConfig(sc)
	}

	if err := self.StoreAndApplyConfig(w, cfg, d); err != nil {
		return
	}
	self.config = cfg

	w.WriteHeader(http.StatusCreated)
}
