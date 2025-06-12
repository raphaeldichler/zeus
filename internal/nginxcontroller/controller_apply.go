// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package nginxcontroller

const ApplyAPIPath = "/apply"

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
