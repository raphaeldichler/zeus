// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package record

import (
	"time"
)

const (
	IngressKey RecordKey = "/v1.0/ingress"
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

func (t TlsState) GoString() string {
	switch t {
	case TlsObtain:
		return "TlsObtain"

	case TlsRenew:
		return "TlsRenew"

	default:
		return "Unknown"
	}
}

type RecordIngress struct {
	Errors   IngressErrorRecord
	Metadata IngressMetadataRecord
	Servers  []*ServerRecord
}

type IngressMetadataRecord struct {
	Name       string
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

type IngressErrorRecord struct {
	Ingress []IngressErrorEntryRecord
	TLS     []IngressErrorEntryRecord
}

type IngressErrorEntryRecord struct {
	Type       string
	Identifier string
	Message    string
}

func (self *RecordIngress) Enabled() bool {
	return len(self.Servers) > 0
}

func NewIngressRecord() RecordIngress {
	return RecordIngress{}
}

func (self *IngressErrorRecord) NoErrors() bool {
	return len(self.TLS) == 0 && len(self.Ingress) == 0
}

func (self *IngressErrorRecord) SetIngressError(entry IngressErrorEntryRecord) {
	self.TLS = append(self.TLS, entry)
}

func (self *IngressErrorRecord) SetTlsError(entry IngressErrorEntryRecord) {
	self.TLS = append(self.TLS, entry)
}

func (self *IngressErrorRecord) ExistsTlsError(entry IngressErrorEntryRecord) bool {
	for _, err := range self.TLS {
		if err.Type == entry.Type && err.Identifier == entry.Identifier {
			return true
		}
	}

	return false
}
