// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package formatter

import (
	"encoding/json"

	"github.com/raphaeldichler/zeus/internal/util/assert"
)

type JSON struct{}

func NewJSONFormatter() *JSON {
	return &JSON{}
}

func (p *JSON) Marshal(obj any) string {
	jsonBytes, err := json.MarshalIndent(obj, "", "  ")
	assert.ErrNil(err)

	return string(jsonBytes)
}
