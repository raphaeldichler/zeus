// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package zeusctl

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/raphaeldichler/zeus/internal/assert"
)

func FormatJSON(data io.Reader) string {
	var out bytes.Buffer
	rawJSON, err := io.ReadAll(data)
	assert.ErrNil(err)

	err = json.Indent(&out, rawJSON, "", "  ")
	assert.ErrNil(err)

	return out.String()
}
