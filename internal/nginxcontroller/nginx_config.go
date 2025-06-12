// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package nginxcontroller

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"

	"github.com/raphaeldichler/zeus/internal/assert"
)

type directory string

func openDirectory() (directory, error) {
	tmp, err := os.MkdirTemp("", "")
	if err != nil {
		return "", err
	}

	return directory(tmp), nil
}

func (d directory) close() error {
	return os.RemoveAll(string(d))
}

func (d directory) store(filename string, content []byte) (string, error) {
	path := filepath.Join(string(d), filename)
	return path, os.WriteFile(path, content, 0600)
}

func (d directory) storeFile(ext string, content []byte) (string, error) {
	assert.StartsNotWith(ext, '.', "the method appends a '.' to the filename")

	b := make([]byte, 16)
	rand.Read(b)
	filename := hex.EncodeToString(b) + "." + ext

	return d.store(filename, content)
}

func (self *IngressRequest) storeAsNginxConfig(d directory) error {
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
		if err := server.write(d, w); err != nil {
			return err
		}
	}

	w.unintend()
	w.writeln("}")

	err := os.WriteFile(NginxConfigPath, w.content(), 0600)
	if err != nil {
		return err
	}

	return nil
}

func newHTTPServer(
	domain string,
	entries ...string,
) *Server {
	return &Server{
		Domain:    domain,
		Tls:       nil,
		Entries:   entries,
		Locations: nil,
	}
}

func newLocation(
	path string,
	matching Matching,
	entries ...string,
) *Location {
	return &Location{
		Path:     path,
		Matching: matching,
		Entries:  entries,
	}
}

func (self *Location) write(w *ConfigBuilder) {
	prefix := "location "
	if self.Matching == Matching_Exact {
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

func (self *IngressRequest) setHTTPLocation(
	domain string,
	loc *Location,
) {
	var httpServer *Server = nil
	for _, server := range self.Servers {
		if server.Domain == domain && server.Tls == nil {
			httpServer = server
			break
		}
	}
	if httpServer == nil {
		httpServer = newHTTPServer(domain)
		self.Servers = append(self.Servers, httpServer)
	}

  httpServer.Locations = append(httpServer.Locations, loc)
}

func (self *IngressRequest) deleteHTTPLocation(
	domain string,
	path string,
	matching Matching,
) *Location {
	var httpServer *Server = nil
  var httpServerIdx = -1
	for idx, server := range self.Servers {
		if server.Domain == domain && server.Tls == nil {
			httpServer = server
      httpServerIdx = idx
			break
		}
	}
	if httpServer == nil {
		return nil
	}

  assert.True(httpServerIdx != -1, "idx must be set correctly")
	for idx, location := range httpServer.Locations {
		if location.Matching == matching && location.Path == path {
			httpServer.Locations[idx] = httpServer.Locations[len(httpServer.Locations)-1]
			httpServer.Locations = httpServer.Locations[:len(httpServer.Locations)-1]

      // if no locations exists anymore it was set via SetHTTPLocation so we remove the server
      if len(httpServer.Locations) == 0 {
        self.Servers[idx] = self.Servers[len(httpServer.Locations)-1]
        self.Servers = self.Servers[:len(self.Servers)-1]
      }

			return location
		}
	}

	return nil
}

func (self *Server) write(d directory, w *ConfigBuilder) error {
	w.writeln("server {")
	w.intend()
	tls := self.Tls

	listen := "listen 80;"
	if tls != nil {
		listen = "listen 443 ssl;"
	}
	w.writeln(listen)
	w.writeln("server_name ", self.Domain, ";")

	for _, entry := range self.Entries {
		assert.EndsNotWith(entry, ';', "cannot end with ';' already appended")
		assert.StartsNotWithString(entry, "listen", "listen cannot be an entry")
		assert.StartsNotWithString(entry, "server_name", "server_name cannot be an entry")
		w.writeln(entry, ";")
	}

	if tls := self.Tls; tls != nil {
		fullchainPath, err := d.storeFile("pem", []byte(tls.Fullchain))
		if err != nil {
			return err
		}

		privkeyPath, err := d.storeFile("pem", []byte(tls.Privkey))
		if err != nil {
			return err
		}

		w.writeln("ssl_certificate ", fullchainPath, ";")
		w.writeln("ssl_certificate_key ", privkeyPath, ";")
	}

	for _, loc := range self.Locations {
		loc.write(w)
	}

	w.unintend()
	w.writeln("}")

	return nil
}
