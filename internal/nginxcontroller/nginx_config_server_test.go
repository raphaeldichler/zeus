// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package nginxcontroller

import "testing"

func TestNginxControllerServerConfigEqual(t *testing.T) {
	sc1 := NewServerConfig("example.com", false, &TlsCertificate{})
	sc2 := NewServerConfig("www.example.com", false, &TlsCertificate{})

	if sc1.Equal(sc2) {
		t.Errorf("server config should not match, but they do")
	}

	sc3 := NewServerConfig("example.com", false, &TlsCertificate{})
	if !sc1.Equal(sc3) {
		t.Errorf("server config should match, but they do not")
	}
}

func TestNginxControllerServerConfigSetRemoveLocation(t *testing.T) {
	sc := NewServerConfig("example.com", false, &TlsCertificate{})

	l1 := NewLocationConfig("/a", ExactMatching)
	l2 := NewLocationConfig("/b", ExactMatching)
	l3 := NewLocationConfig("/a", ExactMatching)
	l4 := NewLocationConfig("/a", PrefixMatching)

	sc.SetLocation(l1)
	sc.SetLocation(l2)
	sc.SetLocation(l3)
	sc.SetLocation(l4)

	if sc.Locations.size() != 3 {
		t.Errorf("server config should contain 3 different configs, but got '%d'", sc.Locations.size())
	}

	l1Return := sc.RemoveLocation(l1)
	if l1.Path != l1Return.Path || l1.Matching != l1Return.Matching {
		t.Errorf("removed location should equal in path and match, but got '%q'", l1Return)
	}
	if sc.Locations.size() != 2 {
		t.Errorf("server config should contain 2 different configs, but got '%d'", sc.Locations.size())
	}

	l2Return := sc.RemoveLocation(l2)
	if l2.Path != l2Return.Path || l2.Matching != l2Return.Matching {
		t.Errorf("removed location should equal in path and match, but got '%q'", l2Return)
	}
	if sc.Locations.size() != 1 {
		t.Errorf("server config should contain 1 different configs, but got '%d'", sc.Locations.size())
	}

	l3Return := sc.RemoveLocation(l1)
	if l3Return != nil {
		t.Errorf("location was already remove, but got still a non nil location '%q'", l3Return)
	}
	if sc.Locations.size() != 1 {
		t.Errorf("server config should contain 1 different configs, but got '%d'", sc.Locations.size())
	}

	l4Return := sc.RemoveLocation(l4)
	if l4.Path != l4Return.Path || l4.Matching != l4Return.Matching {
		t.Errorf("location was already remove, but got still a non nil location '%q'", l3Return)
	}
	if sc.Locations.size() != 0 {
		t.Errorf("server config should contain 0 different configs, but got '%d'", sc.Locations.size())
	}
}
