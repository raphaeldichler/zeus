// Copyright 2025 The Zeus Authors.
// Licensed under the MIT License. See the LICENSE file for details.

package ingress

import (
	"path/filepath"

	"github.com/raphaeldichler/zeus/internal/assert"
)

const (
	NginxInternalMimeTypes = "/etc/nginx/mime.types"
)

// Describes a HTTP config which can be set.
type HttpConfig struct {
	GeneralEntries []string
	EventEntries   []string
	HttpEntries    []string
}

func DefaultHttpConfig() *HttpConfig {
	return &HttpConfig{
		GeneralEntries: []string{
			"worker_processes 1",
		},
		EventEntries: []string{
			"worker_connections 1024",
		},
		HttpEntries: []string{
			"include " + NginxInternalMimeTypes,
			"keepalive_timeout 65",
			"sendfile on",
			"gzip on",
			"include " + filepath.Join(NginxInternalServerPath, "*.conf"),
		},
	}
}

func (self *HttpConfig) FilePath() string {
	return "/etc/nginx/nginx.conf"
}

func (self *HttpConfig) FileContent() []byte {
	w := NewConfigBuilder()

	for _, e := range self.GeneralEntries {
		assert.IsAsciiString(e, "only ascii chars allowed inside the config")
		assert.EndsNotWith(e, ';', "cannot end with ';' already appended")
		w.writeln(e, ";")
	}

	blocks := []struct {
		block   string
		entries []string
	}{
		{
			block:   "events",
			entries: self.EventEntries,
		},
		{
			block:   "http",
			entries: self.HttpEntries,
		},
	}
	for _, b := range blocks {
		w.writeln(b.block, " {")
		w.intend()

		for _, e := range b.entries {
			assert.IsAsciiString(e, "only ascii chars allowed inside the config")
			assert.EndsNotWith(e, ';', "cannot end with ';' already appended")
			w.writeln(e, ";")
		}

		w.unintend()
		w.writeln("}")
	}

	return w.content()
}
