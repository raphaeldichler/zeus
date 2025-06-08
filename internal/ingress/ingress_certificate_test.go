// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package ingress

import (
	"bytes"
	"strings"
	"testing"

	"github.com/raphaeldichler/zeus/internal/assert"
	"github.com/raphaeldichler/zeus/internal/record"
)

func TestIngressCertificateDevelopment(t *testing.T) {
	state := record.ApplicationRecord{
		Ingress: record.NewIngressRecord(),
	}
	state.Ingress.Servers = append(
		state.Ingress.Servers,
		&record.ServerRecord{
			Host: "poseidon.com",
			IPv6: false,
			Tls: &record.TlsRecord{
				CertificateEmail: "some@mail.com",
				State:            record.TlsObtain,
			},
			HTTP: record.HttpRecord{
				Paths: []record.PathRecord{
					{
						Path:     "/",
						Matching: record.MatchingPrefix,
						Service:  "",
					},
				},
			},
		},
	)
	state.Ingress.Servers = append(
		state.Ingress.Servers,
		&record.ServerRecord{
			Host: "admin.poseidon.com",
			IPv6: false,
			Tls: &record.TlsRecord{
				CertificateEmail: "some@mail.com",
				State:            record.TlsObtain,
			},
			HTTP: record.HttpRecord{
				Paths: []record.PathRecord{
					{
						Path:     "/",
						Matching: record.MatchingPrefix,
						Service:  "",
					},
				},
			},
		},
	)

	certificate := DevelopmentCertificate{}
	certificate.GenerateCertificates(&state)

	firstPrivKey := make(map[string][]byte)
	firstFullchain := make(map[string][]byte)
	for _, s := range state.Ingress.Servers {
		firstPrivKey[s.Host] = s.Tls.PrivkeyPem
		firstFullchain[s.Host] = s.Tls.FullchainPem

		if !strings.HasSuffix(s.Host, "localhost") {
			t.Errorf("host must be renamed in development to localhost, but got '%s'", s.Host)
		}

		if len(firstPrivKey) == 0 {
			t.Errorf("no private key was set")
		}

		if len(firstFullchain) == 0 {
			t.Errorf("no public key was set")
		}
	}

	certificate.GenerateCertificates(&state)
	for _, s := range state.Ingress.Servers {
		privKey, ok := firstPrivKey[s.Host]
		assert.True(ok, "must exists")
		if !bytes.Equal(s.Tls.PrivkeyPem, privKey) {
			t.Errorf("new generate should not change the private key (expire time has not exeeded), but it did")
		}

		publicKey, ok := firstFullchain[s.Host]
		assert.True(ok, "must exists")
		if !bytes.Equal(s.Tls.FullchainPem, publicKey) {
			t.Errorf("new generate should not change the public key (expire time has not exeeded), but it did")
		}
	}
}

func TestIngressCertificateDevelopmentToLocalhostDomain(t *testing.T) {
	domains := map[string]string{
		"poseidon.com":           "localhost",
		"app.poseidon.com":       "app.localhost",
		"admin.app.poseidon.com": "admin.app.localhost",
	}

	for domain, exptect := range domains {
		if d := toLocalhostDomain(domain); d != exptect {
			t.Errorf("domain transformation failed. got '%s', wanted '%s'", d, exptect)
		}
	}
}
