// Copyright 2025 The Zeus Authors.
// Licensed under the MIT License. See the LICENSE file for details.

package ingress

import (
	"fmt"
	"strings"

	"github.com/raphaeldichler/zeus/internal/assert"
)

type LocationMatchingType uint32

const (
  LocationPrefix LocationMatchingType = 0
  LocationExact LocationMatchingType = 1
)

var (
  AcmeChallengeLocation = NginxServerLocation{
    Path: "/.well-known/acme-challenge/ ",
    Matching: LocationPrefix,
    Entries: []string{
      "root /var/www/certbot",
    },
  }
  RedirectToHttpsLocation = NginxServerLocation {
    Path: "/",
    Matching: LocationPrefix,
    Entries: []string{
      "return 301 https://\\$host\\$request_uri", 
    },
  }
)

type configWriter struct {
  builder strings.Builder
  currentIndent uint32
}


func newConfigWriter(startIndent uint32) *configWriter {
  return &configWriter{
    currentIndent: startIndent,
  }
}

func (self *configWriter) writeLine(arr ...string) *configWriter {
  self.builder.WriteString(
    strings.Repeat("\t", int(self.currentIndent)),
  )
  /*for _ = range self.currentIndent {
    self.builder.WriteRune('\t')
  }*/

  for _, e := range arr {
    self.builder.WriteString(e)
  }

  self.builder.WriteRune('\n')
  return self
}

func (self *configWriter) intend() *configWriter {
  self.currentIndent++
  return self
}

func (self *configWriter) unintend() *configWriter {
  self.currentIndent--
  return self
}

func (self *configWriter) string() string {
  return self.builder.String()
}

type NginxServerLocation struct {
  Path string
  Matching LocationMatchingType
  Entries []string
}

type NginxServer struct {
  ServerName string
  TlsEnabled bool
  Ipv6 bool
  Entries []string
  Locations []NginxServerLocation
}

func (self *NginxServer) publicSSLCertificate() string {
  return fmt.Sprintf("/etc/nginx/ssl/live/%s/fullchain.pem", self.ServerName)
}

func (self *NginxServer) privateSSLCertificate() string {
  return fmt.Sprintf("/etc/nginx/ssl/live/%s/privkey.pem", self.ServerName)
}


func (self *NginxServerLocation) write(writer *configWriter) {
  {
    fmt.Println("Hello World")

  }
  switch self.Matching {
  case LocationExact:
    writer.writeLine("location = ", strings.TrimSpace(self.Path), " {")
  
  case LocationPrefix:
    writer.writeLine("location ", strings.TrimSpace(self.Path), " {")

  default:
    assert.Unreachable("invalid matching value")

  }

  writer.intend()

  for _, k := range self.Entries {
    assert.EndsNotWith(k, ';', "cannot end with ';' already appended")
    writer.writeLine(k, ";")
  }

  writer.unintend()
  writer.writeLine("}")
}

func (self *NginxServer) httpRedirectToHttps(writer *configWriter) {
  // should it be a custom server instance? makes most sence, we redo stuff there, right?
  writer.writeLine("server {")
  writer.intend()

  writer.writeLine("listen 80;")
  if self.Ipv6 {
    writer.writeLine("listen [::]:80;")
  }
  
  writer.writeLine("server_name ", self.ServerName, ";")
  writer.writeLine("server_tokens off;")
  
  AcmeChallengeLocation.write(writer)
  RedirectToHttpsLocation.write(writer)

  writer.unintend()
  writer.writeLine("}")
}

func (self *NginxServer) string(writer *configWriter) {
  if self.TlsEnabled {
    self.httpRedirectToHttps()
    

  }
  

  writer.writeLine("server {")
  writer.intend()
  defer writer.unintend()


  if self.TlsEnabled {
    writer.writeLine("listen 443 default_server ssl http2")
    if self.Ipv6 {
      writer.writeLine("listen [::]:443 ssl http2")
    }

    writer.writeLine("server_name ", self.ServerName, ";")
    writer.writeLine("ssl_certificate ", self.privateSSLCertificate(), ";")
    writer.writeLine("ssl_certificate_key", self.privateSSLCertificate(), ";")

  }

}

