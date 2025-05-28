// Copyright 2025 The Zeus Authors.
// Licensed under the MIT License. See the LICENSE file for details.

package ingress

import (
	"path/filepath"

	"github.com/raphaeldichler/zeus/internal/assert"
)

const (
	NginxInternalLocationPath = "/etc/nginx/locations"
	NginxInternalServerPath   = "/etc/nginx/sites"
)

type ServerIdentifier struct {
	Domain     string
	TlsEnabled bool
}

func (self *ServerIdentifier) serverId() string {
	port := "80"
	if self.TlsEnabled {
		port = "443"
	}

	return self.Domain + "#" + port
}

func (self *ServerIdentifier) FilePath() string {
	filename := self.serverId() + ".conf"
	return filepath.Join(
		NginxInternalServerPath, filename,
	)
}

// The directory in which the location file is placed in.
func (self *ServerIdentifier) LocationDirectory() string {
	return filepath.Join(
		NginxInternalLocationPath, self.serverId(),
	)
}

type ServerConfig struct {
	ServerIdentifier

	Ipv6    bool
	Entries []string
}

func NewServerConfig(
	serverId ServerIdentifier,
	ipv6Enabled bool,
	entries ...string,
) *ServerConfig {
	return &ServerConfig{
		ServerIdentifier: serverId,
		Ipv6:             ipv6Enabled,
		Entries:          entries,
	}
}

func (self *ServerConfig) FileContent() []byte {
	w := NewConfigBuilder()

	w.writeln("server {")
	w.intend()

	listenIpv4 := "listen 80;"
	listenIpv6 := "listen [::]:80;"
	additionHttp2 := ""
	if self.TlsEnabled {
		listenIpv4 = "listen 443 ssl;"
		listenIpv6 = "listen [::]:443 ssl;"
		additionHttp2 = "http2 on;"
	}

	w.writeln(listenIpv4)
	if self.Ipv6 {
		w.writeln(listenIpv6)
	}
	w.writeln(additionHttp2)

	w.writeln("server_name ", self.Domain, ";")
	for _, entry := range self.Entries {
		assert.EndsNotWith(entry, ';', "cannot end with ';' already appended")
		w.writeln(entry, ";")
	}

	w.writeln("include ", filepath.Join(NginxInternalLocationPath, self.serverId(), "*.conf"), ";")
	w.unintend()
	w.writeln("}")

	return w.content()
}
