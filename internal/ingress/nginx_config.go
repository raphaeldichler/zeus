// Copyright 2025 The Zeus Authors.
// Licensed under the MIT License. See the LICENSE file for details.

package ingress

import (
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
    Matching: LocationExact,
    Entries: []string{
      "root /var/www/certbot",
    },
  }
  RedirectToHttpsLocation = NginxServerLocation {
    Path: "/",
    Matching: LocationExact,
    Entries: []string{
      "return 301 https://\\$host\\$request_uri", 
    },
  }
)

type configWriter struct {
  builder strings.Builder
  currentIndent uint32
}

func (self *configWriter) openBlock() *configWriter {
  self.builder.WriteString("{\n")
  self.currentIndent += 1
  return self
}

func (self *configWriter) closeBlock() *configWriter {
  // how to write with indent?
  // add some method
  self.builder.WriteString("}\n")
  self.currentIndent -= 1
  return self
}



type NginxServerLocation struct {
  Path string
  Matching LocationMatchingType
  Entries []string
}

type NginxServer struct {
  ServerName string
  TlsEnabled bool
  Entries []string
  Locations []NginxServerLocation
}


func (self *NginxServerLocation) String() string {
  var b strings.Builder

  b.WriteString("location ")
  if self.Matching == LocationExact {
    b.WriteString("= ")
  }

  b.WriteString(self.Path)
  b.WriteString(" {\n")
  for _, k := range self.Entries {
    assert.EndsNotWith(k, ';', "cannot end with ';' already appended")
    b.WriteString(k)
    b.WriteString(";\n")
  }
  b.WriteString("}")

  return b.String()
}

func (self *NginxServer) String() string {
  var b strings.Builder

  b.WriteString("server {\n")
  if self.TlsEnabled {
    b.WriteString("listen 443 ssl;\n")
    b.WriteString("server_name ")
    b.WriteString("")

  } else {

  }


  return b.String()
}




