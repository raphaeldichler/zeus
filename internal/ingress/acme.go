// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package ingress

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"

	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
	"github.com/raphaeldichler/zeus/internal/assert"
)

const (
	CADirURL = lego.LEDirectoryStaging
)

type LetEntryptUser struct {
	email        string
	registration *registration.Resource
	key          crypto.PrivateKey
}

func (self *LetEntryptUser) GetEmail() string {
	return self.email
}

func (self *LetEntryptUser) GetRegistration() *registration.Resource {
	return self.registration
}

func (self *LetEntryptUser) GetPrivateKey() crypto.PrivateKey {
	return self.key
}

type CertificateBundle struct {
	FullchainPem []byte
	PrivKeyPem   []byte
}

func ObtainCertificate(
	p challenge.Provider,
	domain string,
	email string,
) (*CertificateBundle, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	user := &LetEntryptUser{
		email: email,
		key:   privateKey,
	}

	config := lego.NewConfig(user)
	config.CADirURL = CADirURL
	client, err := lego.NewClient(config)
	assert.ErrNil(err)

	err = client.Challenge.SetHTTP01Provider(p)
	assert.ErrNil(err)

	// create an account with the email address provided
	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return nil, err
	}
	user.registration = reg

	certificates, err := client.Certificate.Obtain(certificate.ObtainRequest{
		Domains: []string{domain},
		Bundle:  true,
	})
	if err != nil {
		return nil, err
	}

	return &CertificateBundle{
		FullchainPem: certificates.Certificate,
		PrivKeyPem:   certificates.PrivateKey,
	}, nil
}
