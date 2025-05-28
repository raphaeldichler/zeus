// Copyright 2025 The Zeus Authors.
// Licensed under the MIT License. See the LICENSE file for details.

package ingress

import (
	"bytes"
	"strings"

	"github.com/raphaeldichler/zeus/internal/assert"
)

type ConfigBuilder struct {
	builder       bytes.Buffer
	currentIndent int
}

func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{
		currentIndent: 0,
	}
}

func (self *ConfigBuilder) writeln(arr ...string) {
	assert.GreaterThanOrEqual(self.currentIndent, 0, "indent cannot be negative")

	self.builder.WriteString(
		strings.Repeat("\t", int(self.currentIndent)),
	)

	for _, e := range arr {
		assert.IsAsciiString(e, "")
		self.builder.WriteString(e)
	}

	self.builder.WriteRune('\n')
}

func (self *ConfigBuilder) intend() {
	self.currentIndent++
}

func (self *ConfigBuilder) unintend() {
	self.currentIndent--
}

func (self *ConfigBuilder) content() []byte {
	return self.builder.Bytes()
}
