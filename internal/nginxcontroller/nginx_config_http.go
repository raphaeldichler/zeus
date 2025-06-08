// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package nginxcontroller

import (
	"fmt"
	"os"

	"github.com/raphaeldichler/zeus/internal/assert"
)

type MatchingType int

const (
	PrefixMatching MatchingType = iota
	ExactMatching
)

const NginxPidFilePath string = "/run/nginx.pid"

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
			fmt.Sprintf("pid %s", NginxPidFilePath),
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
	assert.NotNil(server, "a server must exists")

	server.SetLocation(loc)
}

func (self *NginxConfig) GetOrCreateHttpServerConfig(
	domain string,
) *ServerConfig {
	server := self.GetHttpServerConfig(domain)
	if server == nil {
		server = NewServerConfig(domain, false, nil)
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
	if server == nil {
		return nil
	}

	return server.RemoveLocation(&LocationsConfig{
		Path:     path,
		Matching: matching,
	})
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

type TlsCertificate struct {
	FullchainFilePath string
	PrivkeyFilePath   string
	Fullchain         []byte
	Privkey           []byte
}

func (self *TlsCertificate) write(w *ConfigBuilder) {
	w.writeln("ssl_certificate ", self.FullchainFilePath, ";")
	w.writeln("ssl_certificate_key ", self.PrivkeyFilePath, ";")
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
