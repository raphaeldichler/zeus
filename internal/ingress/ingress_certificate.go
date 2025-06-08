// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package ingress

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"strings"
	"time"

	"github.com/raphaeldichler/zeus/internal/assert"
	"github.com/raphaeldichler/zeus/internal/ingress/errtype"
	"github.com/raphaeldichler/zeus/internal/nginxcontroller"
	"github.com/raphaeldichler/zeus/internal/record"
)

const (
	// Time until the ingress-daemon will consider the tls certificate to renew.
	// Aka if the time left until the certificate expires is less than this threshold
	TlsRenewThreshold = time.Hour * 24
	// Time until we will renew the certificate
	TlsNewRenewThreshold = time.Hour * 24 * 40
)

func CertificateProviderBuilder(daemon *IngressDaemon, state *record.ApplicationRecord) CertificateProvider {
	switch state.Deployment {

	case record.Development:
		return &DevelopmentCertificate{}

	case record.Production:
		client := nginxcontroller.NewClient(daemon.application, hostNginxControllerSocketPath(daemon.application))
		acmeProvider := NewNginxChallengeProvider(client, daemon.application)

		return &ProductionCertificate{acmeProvider: acmeProvider}

	default:
		assert.Unreachable("All development types need to be covered")
	}

	return nil
}

type CertificateProvider interface {
	GenerateCertificates(state *record.ApplicationRecord)
}

type ProductionCertificate struct {
	acmeProvider *NginxChallengeProvider
}

func (self *ProductionCertificate) GenerateCertificates(state *record.ApplicationRecord) {
	assert.NotNil(self.acmeProvider, "using the acme provider in the next steps, it must exists")

	for _, server := range state.Ingress.Servers {
		tls := server.Tls
		if tls == nil {
			continue
		}

		if tls.State == record.TlsRenew && tls.Expires.Sub(time.Now()) > TlsRenewThreshold {
			continue
		}

		// we dont differatiate between renew and obtain. renew can result in obtaining a new one
		// therefore we just obtain a new one to reduce the possiblilties of failures
		certBundle, err := ObtainCertificate(
			self.acmeProvider,
			server.Host,
			tls.CertificateEmail,
		)
		if err != nil {
			state.Ingress.Errors.SetTlsError(
				errtype.FailedObtainCertificate(server.Host, err),
			)
			continue
		}

		tls.FullchainPem = certBundle.FullchainPem
		tls.PrivkeyPem = certBundle.PrivKeyPem
		tls.State = record.TlsRenew
		tls.Expires = time.Now().Add(TlsNewRenewThreshold)
	}
}

func toLocalhostDomain(domain string) string {
	if strings.HasSuffix(domain, "localhost") {
		return domain
	}

	levels := strings.Split(domain, ".")
	assert.True(len(levels) >= 2, "at least a top and second level domain must exists")

	return strings.Join(append(levels[:len(levels)-2], "localhost"), ".")
}

type DevelopmentCertificate struct{}

func (self *DevelopmentCertificate) GenerateCertificates(state *record.ApplicationRecord) {
	for _, server := range state.Ingress.Servers {
		tls := server.Tls
		if tls == nil {
			continue
		}

		if tls.State == record.TlsRenew && tls.Expires.Sub(time.Now()) > TlsRenewThreshold {
			continue
		}

		domain := toLocalhostDomain(server.Host)
		server.Host = domain

		priv, err := rsa.GenerateKey(rand.Reader, 2048)
		assert.ErrNil(err)
		expires := time.Now().Add(TlsNewRenewThreshold)

		serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
		assert.ErrNil(err)

		template := x509.Certificate{
			SerialNumber: serial,
			Subject: pkix.Name{
				CommonName: domain,
			},
			NotBefore:             time.Now(),
			NotAfter:              expires,
			KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
			ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
			BasicConstraintsValid: true,
			DNSNames:              []string{domain},
		}

		certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
		assert.ErrNil(err)

		certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
		keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

		tls.FullchainPem = certPEM
		tls.PrivkeyPem = keyPEM
		tls.State = record.TlsRenew
		tls.Expires = expires
	}
}
