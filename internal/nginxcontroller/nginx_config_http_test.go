// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package nginxcontroller

import "testing"

func TestNginxControllerHttpConfigSetServerConfig(t *testing.T) {
	cfg := NewNginxConfig()

  sc1 := NewServerConfig("example.com", false, nil)
  sc2 := NewServerConfig("www.example.com", false, nil)

  cfg.SetServerConfig(sc1)
  cfg.SetServerConfig(sc2)

  sc1Get := cfg.GetHttpServerConfig("example.com")
  if !sc1.Equal(sc1Get) {
    t.Errorf("Failed to set and get server config. Wanted '%v', got '%v'", sc1, sc1Get)
  }

  sc2Get := cfg.GetHttpServerConfig("www.example.com")
  if !sc2.Equal(sc2Get) {
    t.Errorf("Failed to set and get server config. Wanted '%v', got '%v'", sc2, sc2Get)
  }

  l3 := NewLocationConfig("/path", ExactMatching)
  cfg.SetHttpLocation("app.example.com", l3)
  sc3Get := cfg.GetHttpServerConfig("app.example.com")
  if sc3Get.Domain != "app.example.com" || sc3Get.IsTlsEnabled() {
    t.Errorf("Setted an HTTP location to 'app.example.com', but got back '%v'", sc3Get)
  }

  l3Removed := sc3Get.RemoveLocation(l3)
  if l3Removed == nil || !l3.Equal(l3Removed) {
    t.Errorf("Removed location, but got wrong location back. want '%v', got '%v'", l3, l3Removed)
  }
}


func TestNginxControllerHttpConfigGetOrCreate(t *testing.T) {
	cfg := NewNginxConfig()

  if c := cfg.GetHttpServerConfig("example.com"); c != nil {
    t.Errorf("Got wrong server back, wanted none, got '%v'", c)
  }

  sc := cfg.GetOrCreateHttpServerConfig("example.com")
  if sc == nil || sc.Domain != "example.com" || sc.IsTlsEnabled() {
    t.Errorf("Got wrong server back, wanted HTTP with domain 'example.com', got '%v'", sc)
  }


  l := NewLocationConfig("/a", ExactMatching)
  cfg.SetHttpLocation("example.com", l)
  if sc.Locations.size() != 1 {
    t.Errorf("Setted location over pointer. Wanted server location contain 1 entry, got %d", sc.Locations.size())
  }
}
