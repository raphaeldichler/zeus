// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package record

import (
	"time"

	"github.com/raphaeldichler/zeus/internal/util/assert"
)

const (
	MatchingExact  = "exact"
	MatchingPrefix = "prefix"
)

type TlsState int

const (
	TlsObtain TlsState = iota
	TlsRenew
)

type RecordIngress struct {
	Metadata IngressMetadataRecord
	Errors   []*IngressErrorEntryRecord
	Servers  []*ServerRecord
}

type IngressMetadataRecord struct {
	CreateTime time.Time
	Image      string
}

type ServerRecord struct {
	Host string
	IPv6 bool
	Tls  *TlsRecord
	HTTP HttpRecord
}

type HttpRecord struct {
	Paths []PathRecord
}

type PathRecord struct {
	Path     string
	Matching string
	Service  RecordKey
}

type TlsRecord struct {
	CertificateEmail string
	State            TlsState
	Expires          time.Time
	PrivkeyPem       []byte
	FullchainPem     []byte
}

type IngressErrorEntryRecord struct {
	Type       string
	Identifier string
	Message    string
}

func (self *RecordIngress) Enabled() bool {
	if self == nil {
		return false
	}
	return len(self.Servers) > 0
}

func NewIngressRecord() *RecordIngress {
	return &RecordIngress{
		Metadata: IngressMetadataRecord{
			CreateTime: time.Now(),
			Image:      "zeus-nginx:v0.1",
		},
	}
}

func (self *RecordIngress) NoErrors() bool {
	return len(self.Errors) == 0
}

func (self *RecordIngress) SetError(entry IngressErrorEntryRecord) {
	self.Errors = append(self.Errors, &IngressErrorEntryRecord{
		Type:       entry.Type,
		Identifier: entry.Identifier,
		Message:    entry.Message,
	})
}

func (self *RecordIngress) HasError(entry IngressErrorEntryRecord) bool {
	for _, err := range self.Errors {
		if err.Type == entry.Type && err.Identifier == entry.Identifier {
			return true
		}
	}

	return false
}

func (self *RecordIngress) Sync(other *RecordIngress) {
	assert.True(self.Metadata.Image == other.Metadata.Image, "image must be the same")
	assert.True(self.Metadata.CreateTime.Equal(other.Metadata.CreateTime), "create time must be the same")

	self.Errors = other.Errors
	self.Servers = other.Servers
}
