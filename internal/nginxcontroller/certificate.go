// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package nginxcontroller

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"path/filepath"
	"time"

	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
	"github.com/raphaeldichler/zeus/internal/assert"
)

const (
	CADirURL = lego.LEDirectoryStaging
	// Time until the ingress-daemon will consider the tls certificate to renew.
	// Aka if the time left until the certificate expires is less than this threshold
	TlsRenewThreshold = time.Hour * 24
)

func acmeLocationPath(token string) string {
	return filepath.Join("/.well-known/acme-challenge/", token)
}

type letEntryptUser struct {
	email        string
	registration *registration.Resource
	key          crypto.PrivateKey
}

func (self *letEntryptUser) GetEmail() string {
	return self.email
}

func (self *letEntryptUser) GetRegistration() *registration.Resource {
	return self.registration
}

func (self *letEntryptUser) GetPrivateKey() crypto.PrivateKey {
	return self.key
}

type certificateBundle struct {
	FullchainPem []byte
	PrivKeyPem   []byte
}

func obtainCertificate(
	p challenge.Provider,
	domain string,
	email string,
) (*certificateBundle, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	user := &letEntryptUser{
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

	return &certificateBundle{
		FullchainPem: certificates.Certificate,
		PrivKeyPem:   certificates.PrivateKey,
	}, nil
}

func (self *Controller) Present(
	domain string,
	token string,
	keyAuth string,
) error {
	d, err := openDirectory()
	if err != nil {
		return err
	}
	defer time.AfterFunc(time.Minute, func() { d.close() })

	self.config.setHTTPLocation(
		domain,
		newLocation(
			acmeLocationPath(token),
			Matching_Exact,
			fmt.Sprintf(`return 200 "%s"`, keyAuth),
			"add_header Content-Type text/plain",
		),
	)

	if err := self.storeAndApplyConfig(d); err != nil {
		return err
	}

	return nil
}

func (self *Controller) CleanUp(
	domain string,
	token string,
	keyAuth string,
) error {
	d, err := openDirectory()
	if err != nil {
		return err
	}
	defer time.AfterFunc(time.Minute, func() { d.close() })

	loc := self.config.deleteHTTPLocation(domain, acmeLocationPath(token), Matching_Exact)
	assert.NotNil(loc, "cleanup must clean a valid location")

	if err := self.storeAndApplyConfig(d); err != nil {
		self.config.setHTTPLocation(domain, loc)
		return err
	}

	return nil
}
