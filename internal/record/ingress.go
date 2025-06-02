// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package record

import "time"

const (
	IngressKey RecordKey = "/v1.0/ingress"
)

type TlsState int

const (
	TlsObtain TlsState = iota
	TlsRenew
)

type RecordIngress struct {
	Servers []ServerRecord
}

type ServerRecord struct {
	Host string
	IPv6 bool
	Tls  *TlsRecord
	HTTP []HTTP
}

type HTTP struct {
	Paths []Path
}

type Path struct {
	Path    string
	Type    string
	Service RecordKey
}

type TlsRecord struct {
	CertificateEmail string
	State            TlsState
	Expires          time.Time
	PrivkeyPem       []byte
	FullchainPem     []byte
}
