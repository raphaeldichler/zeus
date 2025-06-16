// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package formatter

type Output interface {
	Marshal(obj any) string
}
