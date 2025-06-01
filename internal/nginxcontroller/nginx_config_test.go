// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package nginxcontroller

import "testing"

func TestNginxControllerNginxConfigAddHttpLocation(t *testing.T) {
	domain := "localhost"
	cfg := NewNginxConfig()

	l1 := NewLocationConfig(
		"/l1",
		PrefixMatching,
	)
	l2 := NewLocationConfig(
		"/l2",
		PrefixMatching,
	)

	cfg.SetHttpLocation(domain, l1)
	sc := cfg.GetHttpServerConfig(domain)
	if len(sc.Locations) != 1 {
		t.Errorf("failed to set http location. expected 1 location set, got %d", len(sc.Locations))
	}

	cfg.SetHttpLocation(domain, l2)
	if len(sc.Locations) != 2 {
		t.Errorf("failed to set http location. expected 2 location set, got %d", len(sc.Locations))
	}

	l3 := NewLocationConfig(
		"/l1",
		PrefixMatching,
	)
	cfg.SetHttpLocation(domain, l3)
	if len(sc.Locations) != 2 {
		t.Errorf("failed to set http location. expected 2 location set, got %d", len(sc.Locations))
	}
}
