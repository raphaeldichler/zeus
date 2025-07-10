// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package errtype

import (
	"github.com/raphaeldichler/zeus/internal/assert"
	"github.com/raphaeldichler/zeus/internal/record"
)

func FailedObtainCertificate(
  host string, 
  err error,
) record.IngressErrorEntryRecord {
	assert.True(err != nil, "the error must exist")
	return record.IngressErrorEntryRecord{
		Type:       "FailedObtainCertificate",
		Identifier: host,
		Message:    err.Error(),
	}
}

func FailedInteractionWithNginxController(
  domain string,
  err error,
) record.IngressErrorEntryRecord {
	assert.True(err != nil, "the error must exist")

	return record.IngressErrorEntryRecord{
		Type:       "FailedInteractionWithNginxController",
		Identifier: domain,
		Message:    err.Error(),
	}
}
