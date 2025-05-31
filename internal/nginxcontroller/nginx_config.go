// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package nginxcontroller


type MatchingType int
const (
  PrefixMatching MatchingType = iota
  ExactMatching 
)

type NginxConfig struct {
  Servers []*ServerConfig
}

func NewNginxConfig() *NginxConfig {
  return &NginxConfig {
    Servers: make([]*ServerConfig, 0),
  }
}

func (self *NginxConfig) Content() []byte  {
  w := NewConfigBuilder()

  for _, server := range self.Servers {
    server.write(w)
  }

  return w.content()
}

func (self *NginxConfig) AddServerConfig(s *ServerConfig) {
  self.Servers = append(self.Servers, s)
}

func (self *NginxConfig) GetHttpServerConfig(
  domain string,
) *ServerConfig {
  var result *ServerConfig = nil
  for _, server := range self.Servers {
    if server.Domain == domain && server.Tls == nil {
      result = server
      break
    }
  }

  return result
}

func (self *NginxConfig) GetOrCreateHttpServerConfig(
  domain string,
) *ServerConfig {
  result := self.GetHttpServerConfig(domain)
  if result == nil {
    result = &ServerConfig {
      Domain: domain,
      Tls: nil,
      IPv6: false,
      Locations: make([]*LocationsConfig, 0),
    }
  }

  return result
}

func (self *NginxConfig) DeleteHttpLocation(
  domain string,
  path string,
  matching MatchingType,
) bool {
  server := self.GetHttpServerConfig(domain)
  for idx, loc := range server.Locations {
    if loc.Path == path && loc.Matching == matching {
      server.Locations[idx] = server.Locations[len(server.Locations) - 1]
      server.Locations = server.Locations[:len(server.Locations) - 1]

      return true
    }
  }

  return false
}

func (self *NginxConfig) ExistServerConfig(
  domain string,
) bool {
  server := self.GetHttpServerConfig(domain)

  return server != nil
}

type TlsCertificate struct {
  FullchainFilePath string
  PrivkeyFilePath string
}

func (self *TlsCertificate) write(w *ConfigBuilder) {
  w.writeln("ssl_certificate ", self.FullchainFilePath)
  w.writeln("ssl_certificate_key ", self.PrivkeyFilePath)
}

type ServerConfig struct {
	Domain string 
  Tls *TlsCertificate
	IPv6 bool 
  Locations []*LocationsConfig 
}

func NewServerConfig(
  domain string,
  ipv6Enabled bool,
  tls *TlsCertificate,
) *ServerConfig {
  return &ServerConfig {
    Domain: domain,
    Tls: tls,
    IPv6: ipv6Enabled,
    Locations: make([]*LocationsConfig, 0),
  }
}

func (self *ServerConfig) AddLocation(l *LocationsConfig) {
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
	Path string
	Matching MatchingType
  Entries []string
}

func NewLocationConfig(
  path string,
  matching MatchingType,
  entries ...string,
) *LocationsConfig {
  return &LocationsConfig{
    Path: path,
    Matching: matching,
    Entries: entries,
  }
}


func (self *LocationsConfig) write(w *ConfigBuilder) {
  prefix := "location "
	if self.Matching == ExactMatching {
    prefix = "location = "
	} 
	w.writeln(prefix, self.Path, " {")
	w.intend()

	w.unintend()
	w.writeln("}")
}
