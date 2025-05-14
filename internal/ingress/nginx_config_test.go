// Copyright 2025 The Zeus Authors.
// Licensed under the MIT License. See the LICENSE file for details.

package ingress

import (
	"testing"
)



func TestLocationConfig(t *testing.T) {
  cfg := NginxServerLocation{
    Path: "/example",
    Matching: LocationPrefix,
    Entries: []string {
      "proxy_pass http://localhost:8080",
    },
  }

  writer := newConfigWriter(0)
  cfg.write(writer)

  cfgStr := writer.string()
  expectCfg := "location /example {\n\tproxy_pass http://localhost:8080;\n}\n"

  if cfgStr != expectCfg {
    t.Errorf("producing config string failed. want = %q, got %q", expectCfg, cfgStr)

  }
}

func TestLocationExactConfig(t *testing.T) {
  cfg := NginxServerLocation{
    Path: "/example",
    Matching: LocationExact,
    Entries: []string {
      "proxy_pass http://localhost:8080",
    },
  }

  writer := newConfigWriter(0)
  cfg.write(writer)

  cfgStr := writer.string()
  expectCfg := "location = /example {\n\tproxy_pass http://localhost:8080;\n}\n"

  if cfgStr != expectCfg {
    t.Errorf("producing config string failed. want = %q, got %q", expectCfg, cfgStr)

  }
}
