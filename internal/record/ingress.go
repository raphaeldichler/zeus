// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package record

import "time"

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

func (self *RecordIngress) Enabled() bool {
	return len(self.Servers) > 0
}
