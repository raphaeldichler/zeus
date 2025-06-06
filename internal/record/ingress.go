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

type RecordIngress struct {
	Errors   IngressErrorRecord
	Metadata IngressMetadataRecord
	Servers  []ServerRecord
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
	Server  []IngressErrorEntryRecord
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

func (self *IngressErrorRecord) SetTlsError(
	errorType string,
	identifier string,
	message string,
) {
	self.TLS = append(
		self.TLS,
		IngressErrorEntryRecord{
			Type:       errorType,
			Identifier: identifier,
			Message:    message,
		},
	)
}

func (self *IngressErrorRecord) SetServerError(
	errorType string,
	identifier string,
	message string,
) {
	self.Server = append(
		self.Server,
		IngressErrorEntryRecord{
			Type:       errorType,
			Identifier: identifier,
			Message:    message,
		},
	)
}

func (self *IngressErrorRecord) SetIngressError(
	errorType string,
	identifier string,
	message string,
) {
	self.Server = append(
		self.Server,
		IngressErrorEntryRecord{
			Type:       errorType,
			Identifier: identifier,
			Message:    message,
		},
	)
}

func (self *IngressErrorRecord) ExistsTlsError(
	errorType string,
	identifier string,
) bool {
	for _, err := range self.TLS {
		if err.Type == errorType && err.Identifier == identifier {
			return true
		}
	}

	return false
}
