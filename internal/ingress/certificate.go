// Copyright 2025 The Zeus Authors.
// Licensed under the MIT License. See the LICENSE file for details.

package ingress

import (
	"path/filepath"

	"github.com/raphaeldichler/zeus/internal/runtime"
)

const (
	NginxInternalCertificatePath = "/etc/letsencrypt/live/"
)

type CertificateIdentifier struct {
	domain string
}

func (self *CertificateIdentifier) PublicKeyPath() string {
	return filepath.Join(NginxInternalCertificatePath, self.domain, "fullchain.pem")
}

func (self *CertificateIdentifier) PrivateKeyPath() string {
	return filepath.Join(NginxInternalCertificatePath, self.domain, "privkey.pem")
}

func (self *CertificateIdentifier) DirectoryPath() string {
	return filepath.Join(NginxInternalCertificatePath, self.domain)
}

type Certificate struct {
	CertificateIdentifier

	privateKey     []byte
	certificatePEM []byte
}

func NewCertificate(
	domain string,
	privateKey []byte,
	certificatePEM []byte,
) *Certificate {
	return &Certificate{
		CertificateIdentifier: CertificateIdentifier{
			domain: domain,
		},
		privateKey:     privateKey,
		certificatePEM: certificatePEM,
	}
}

// Sets the certificate file for nginx to /etc/letsencrypt/live/{domain}/fullchain.pem;
func (self *Certificate) PublicKey() runtime.FileContent {
	return &runtime.BasicFileContent{
		Path:    self.PublicKeyPath(),
		Content: self.certificatePEM,
	}
}

// Sets the private key file for nginx to /etc/letsencrypt/live/{domain}/privkey.pem;
func (self *Certificate) PrivateKey() runtime.FileContent {
	return &runtime.BasicFileContent{
		Path:    self.PrivateKeyPath(),
		Content: self.privateKey,
	}
}
