// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package nginxcontroller

import (
	"os"

	"github.com/raphaeldichler/zeus/internal/assert"
)

type MatchingType int

const (
	PrefixMatching MatchingType = iota
	ExactMatching
)

type NginxConfig struct {
	GeneralEntries []string
	EventEntries   []string
	HttpEntries    []string
	Servers        []*ServerConfig
}

func NewNginxConfig() *NginxConfig {
	return &NginxConfig{
		GeneralEntries: []string{
			"worker_processes 1",
			"pid /run/nginx.pid",
			"user nginx",
		},
		EventEntries: []string{
			"worker_connections 1024",
		},
		HttpEntries: []string{
			"include /etc/nginx/mime.types",
			"default_type application/octet-stream",
			"keepalive_timeout 65",
			"sendfile on",
			"gzip on",
		},
		Servers: make([]*ServerConfig, 0),
	}
}

func (self *NginxConfig) Store(d directory) error {
	w := NewConfigBuilder()

	for _, e := range self.GeneralEntries {
		assert.IsAsciiString(e, "only ascii chars allowed inside the config")
		assert.EndsNotWith(e, ';', "cannot end with ';' already appended")
		w.writeln(e, ";")
	}

	w.writeln("events {")
	w.intend()
	for _, e := range self.EventEntries {
		assert.IsAsciiString(e, "only ascii chars allowed inside the config")
		assert.EndsNotWith(e, ';', "cannot end with ';' already appended")
		w.writeln(e, ";")
	}
	w.unintend()
	w.writeln("}")

	w.writeln("http {")
	w.intend()
	for _, e := range self.HttpEntries {
		assert.IsAsciiString(e, "only ascii chars allowed inside the config")
		assert.EndsNotWith(e, ';', "cannot end with ';' already appended")
		w.writeln(e, ";")
	}

	for _, server := range self.Servers {
		if tls := server.Tls; tls != nil {
			if err := tls.store(d); err != nil {
				return err
			}
		}

		server.write(w)
	}

	w.unintend()
	w.writeln("}")

	err := os.WriteFile(NginxConfigPath, w.content(), 0600)
	if err != nil {
		return err
	}

	return nil
}

func (self *NginxConfig) SetServerConfig(s *ServerConfig) {
	self.Servers = append(self.Servers, s)
}

func (self *NginxConfig) GetHttpServerConfig(
	domain string,
) *ServerConfig {
	for _, server := range self.Servers {
		if server.Domain == domain && server.Tls == nil {
			return server
		}
	}

	return nil
}

func (self *NginxConfig) SetHttpLocation(
	domain string,
	loc *LocationsConfig,
) {
	server := self.GetOrCreateHttpServerConfig(domain)
	server.SetLocation(loc)
}

func (self *NginxConfig) GetOrCreateHttpServerConfig(
	domain string,
) *ServerConfig {
	server := self.GetHttpServerConfig(domain)
	if server == nil {
		server = &ServerConfig{
			Domain:    domain,
			Tls:       nil,
			IPv6:      false,
			Locations: make([]*LocationsConfig, 0),
		}
		self.Servers = append(self.Servers, server)
	}

	return server
}

func (self *NginxConfig) DeleteHttpLocation(
	domain string,
	path string,
	matching MatchingType,
) *LocationsConfig {
	server := self.GetHttpServerConfig(domain)
	for idx, loc := range server.Locations {
		if loc.Path == path && loc.Matching == matching {
			loc := server.Locations[idx]
			server.Locations[idx] = server.Locations[len(server.Locations)-1]
			server.Locations = server.Locations[:len(server.Locations)-1]

			return loc
		}
	}

	return nil
}

type TlsCertificate struct {
	FullchainFilePath string
	PrivkeyFilePath   string
	Fullchain         []byte
	Privkey           []byte
}

func (self *TlsCertificate) write(w *ConfigBuilder) {
	w.writeln("ssl_certificate ", self.FullchainFilePath)
	w.writeln("ssl_certificate_key ", self.PrivkeyFilePath)
}

func (self *TlsCertificate) store(d directory) error {
	fullchainPath, err := d.storeFile("pem", self.Fullchain)
	if err != nil {
		return err
	}

	privkeyPath, err := d.storeFile("pem", self.Privkey)
	if err != nil {
		return err
	}

	self.FullchainFilePath = fullchainPath
	self.PrivkeyFilePath = privkeyPath

	return nil
}

type ServerConfig struct {
	Domain    string
	Tls       *TlsCertificate
	IPv6      bool
	Locations []*LocationsConfig
}

func NewServerConfig(
	domain string,
	ipv6Enabled bool,
	tls *TlsCertificate,
) *ServerConfig {
	return &ServerConfig{
		Domain:    domain,
		Tls:       tls,
		IPv6:      ipv6Enabled,
		Locations: make([]*LocationsConfig, 0),
	}
}

func (self *ServerConfig) SetLocation(l *LocationsConfig) {
	for idx, e := range self.Locations {
		if e.Path == l.Path && e.Matching == l.Matching {
			self.Locations[idx] = self.Locations[len(self.Locations)-1]
			self.Locations = self.Locations[:len(self.Locations)-1]
			break
		}
	}

	self.Locations = append(self.Locations, l)
}

func (self *ServerConfig) IsTlsEnabled() bool {
	return self.Tls != nil
}

func (self *ServerConfig) write(w *ConfigBuilder) {
	w.writeln("server {")
	w.intend()

	listenIpv4 := "listen 80;"
	listenIpv6 := "listen [::]:80;"
	additionHttp2 := ""
	if self.IsTlsEnabled() {
		listenIpv4 = "listen 443 ssl;"
		listenIpv6 = "listen [::]:443 ssl;"
		additionHttp2 = "http2 on;"
	}

	w.writeln(listenIpv4)
	if self.IPv6 {
		w.writeln(listenIpv6)
	}
	w.writeln(additionHttp2)

	w.writeln("server_name ", self.Domain, ";")

	if self.IsTlsEnabled() {
		self.Tls.write(w)
	}

	for _, loc := range self.Locations {
		loc.write(w)
	}

	w.unintend()
	w.writeln("}")
}

type LocationsConfig struct {
	Path     string
	Matching MatchingType
	Entries  []string
}

func NewLocationConfig(
	path string,
	matching MatchingType,
	entries ...string,
) *LocationsConfig {
	return &LocationsConfig{
		Path:     path,
		Matching: matching,
		Entries:  entries,
	}
}

func (self *LocationsConfig) write(w *ConfigBuilder) {
	prefix := "location "
	if self.Matching == ExactMatching {
		prefix = "location = "
	}
	w.writeln(prefix, self.Path, " {")
	w.intend()

	for _, e := range self.Entries {
		assert.EndsNotWith(e, ';', "cannot end with ';' already appended")
		w.writeln(e, ";")
	}

	w.unintend()
	w.writeln("}")
}
