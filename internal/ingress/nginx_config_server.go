// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package ingress

import (
	"path/filepath"
	"strings"

	"github.com/raphaeldichler/zeus/internal/assert"
)

const (
	NginxInternalLocationPath = "/etc/nginx/locations"
	NginxInternalServerPath   = "/etc/nginx/sites"
)

type ServerIdentifier struct {
	Domain     string
	TlsEnabled bool
	IPv6    bool
}

func (self *ServerIdentifier) serverId() string {
  port := "#80"
	if self.TlsEnabled {
		port = "#443"
	}

  ip := "#ipv4"
  if self.IPv6 {
    ip = "#ipv6"
  }

	return self.Domain + port + ip
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

	Entries []string
}

func ServerIdentifierFromPath(path string) (ServerIdentifier, bool) {
  _, file := filepath.Split(path)
  assert.EndsWith(file, ".conf", "we only parse paths which include .conf")
  file = file[:len(file) - len(".conf")]

  return parseServerIdentifierString(file)
}

func parseServerIdentifierString(serverID string) (ServerIdentifier, bool) {
  parts := strings.Split(serverID, "#")
  if len(parts) != 3 {
    return ServerIdentifier{}, false
  }

  portPart := parts[1]
  if ok := portPart == "443" || portPart == "80"; !ok {
    return ServerIdentifier{}, false
  }
  tlsEnabled := portPart == "443"

  ipPart := parts[2]
  if ok := ipPart == "ipv4" || ipPart == "ipv6"; !ok {
    return ServerIdentifier{}, false
  }
  ipv6Enabled := ipPart == "ipv6"

  return ServerIdentifier {
    Domain: parts[0],
    TlsEnabled: tlsEnabled,
    IPv6: ipv6Enabled,
  }, true
}



func NewServerConfig(
	serverId ServerIdentifier,
	entries ...string,
) *ServerConfig {
	return &ServerConfig{
		ServerIdentifier: serverId,
		Entries:          entries,
	}
}

func (self *ServerConfig) Equal(other *ServerConfig) bool {

  return true
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
	if self.IPv6 {
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
