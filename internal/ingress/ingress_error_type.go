// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package ingress

import "github.com/raphaeldichler/zeus/internal/record"

func Foo() record.IngressErrorRecord {
	return record.IngressErrorRecord{}
}

func ErrTypeFailedCreatingIngressContainer() record.IngressErrorEntryRecord {
	return record.IngressErrorEntryRecord{
		Type:       "FailedCreatingIngressContainer",
		Identifier: "",
		Message:    "",
	}
}
