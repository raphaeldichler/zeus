// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package nginxcontroller

import "github.com/raphaeldichler/zeus/internal/assert"

type LocationsConfig struct {
	Path     string
	Matching MatchingType
	Entries  []string
}

func NewLocationConfig(
	path string,
	matching MatchingType,
	entries ...string,
) *LocationsConfig {
	return &LocationsConfig{
		Path:     path,
		Matching: matching,
		Entries:  entries,
	}
}

func (self LocationsConfig) Equal(other *LocationsConfig) bool {
	return self.Path == other.Path && self.Matching == other.Matching
}

func (self *LocationsConfig) write(w *ConfigBuilder) {
	prefix := "location "
	if self.Matching == ExactMatching {
		prefix = "location = "
	}
	w.writeln(prefix, self.Path, " {")
	w.intend()

	for _, e := range self.Entries {
		assert.EndsNotWith(e, ';', "cannot end with ';' already appended")
		w.writeln(e, ";")
	}

	w.unintend()
	w.writeln("}")
}
