// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package nginxcontroller

type ServerConfig struct {
	Domain        string
	Tls           *TlsCertificate
	IPv6          bool
	Locations     Set[LocationsConfig]
	DefaultServer bool
}

func NewServerConfig(
	domain string,
	ipv6Enabled bool,
	tls *TlsCertificate,
	defaultServer bool,
) *ServerConfig {
	return &ServerConfig{
		Domain:        domain,
		Tls:           tls,
		IPv6:          ipv6Enabled,
		Locations:     NewSet[LocationsConfig](),
		DefaultServer: defaultServer,
	}
}

func (self ServerConfig) Equal(other *ServerConfig) bool {
	return self.Domain == other.Domain && self.IsTlsEnabled() == other.IsTlsEnabled()
}

func (self *ServerConfig) RemoveLocation(l *LocationsConfig) *LocationsConfig {
	return self.Locations.remove(l)
}

func (self *ServerConfig) SetLocation(l *LocationsConfig) {
	self.Locations.set(l)
}

func (self *ServerConfig) IsTlsEnabled() bool {
	return self.Tls != nil
}

func (self *ServerConfig) write(w *ConfigBuilder) {
	w.writeln("server {")
	w.intend()

	defaultServerAddition := ""
	if self.DefaultServer {
		defaultServerAddition = " default_server"
	}

	listenIpv4 := "listen 80"
	listenIpv6 := "listen [::]:80"
	additionHttp2 := ""
	if self.IsTlsEnabled() {
		listenIpv4 = "listen 443 ssl"
		listenIpv6 = "listen [::]:443 ssl"
		additionHttp2 = "http2 on;"
	}

	w.writeln(listenIpv4, defaultServerAddition, ";")
	if self.IPv6 {
		w.writeln(listenIpv6, defaultServerAddition, ";")
	}
	w.writeln(additionHttp2)

	w.writeln("server_name ", self.Domain, ";")

	if self.IsTlsEnabled() {
		self.Tls.write(w)
	}

	for _, loc := range self.Locations.entries() {
		loc.write(w)
	}

	if self.DefaultServer {
		fallback := &LocationsConfig{
			Path:     "/",
			Matching: PrefixMatching,
			Entries: []string{
				"return 444",
			},
		}
		fallback.write(w)
	}

	w.unintend()
	w.writeln("}")
}
