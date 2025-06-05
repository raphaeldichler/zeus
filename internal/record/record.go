// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package record

type RecordKey string

type ApplicationRecord struct {
	Ingress RecordIngress
	Service RecordService
}

// Persists the application state
func (self *ApplicationRecord) Apply() error {
	return nil
}
