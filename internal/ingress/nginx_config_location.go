// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package ingress

import (
	"encoding/base64"
	"path/filepath"
	"slices"
	"strings"

	"github.com/raphaeldichler/zeus/internal/assert"
)

type LocationMatchingType uint32

const (
	LocationPrefix LocationMatchingType = 0
	LocationExact  LocationMatchingType = 1
)

type LocationIdentifier struct {
	ServerIdentifier

	Path     string
	Matching LocationMatchingType
}

func (self *LocationIdentifier) id() string {
	prefix := "location "
	if self.Matching == LocationExact {
		prefix += "= "
	}

	path := strings.TrimSpace(self.Path)
	return base64.RawStdEncoding.EncodeToString(
		[]byte(prefix + path),
	)
}

// The path to the location file.
func (self *LocationIdentifier) FilePath() string {
	filename := self.id() + ".conf"
	return filepath.Join(self.LocationDirectory(), filename)
}

func (self *LocationIdentifier) IsExactLocation() bool {
	return self.Matching == LocationExact
}

type LocationConfig struct {
	LocationIdentifier

	Entries []string
}

func (self *LocationConfig) Equal(other *LocationConfig) bool {
  if self.LocationIdentifier != other.LocationIdentifier {
    return false
  }

  return slices.Equal(self.Entries, other.Entries)
}

func NewLocationConfig(
	locationID LocationIdentifier,
	entries ...string,
) *LocationConfig {

	return &LocationConfig{
		LocationIdentifier: locationID,
		Entries:            entries,
	}
}

func (self *LocationConfig) FileContent() []byte {
	w := NewConfigBuilder()
	path := strings.TrimSpace(self.Path)

	if self.IsExactLocation() {
		w.writeln("location = ", path, " {")
	} else {
		w.writeln("location ", path, " {")
	}
	w.intend()

	for _, k := range self.Entries {
		assert.EndsNotWith(k, ';', "cannot end with ';' already appended")
		w.writeln(k, ";")
	}

	w.unintend()
	w.writeln("}")

	return w.content()
}
