// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package ingress

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"

	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
	"github.com/raphaeldichler/zeus/internal/assert"
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

type NginxHttp01Provider struct {
	controller *NginxController
}

func (self *NginxHttp01Provider) Present(
	domain string,
	token string,
	keyAuth string,
) error {
	if err := self.controller.SetAcmeChallengeLocation(
		domain,
		token,
		keyAuth,
	); err != nil {
		return err
	}

	if err := self.controller.ApplyConfig(); err != nil {
		return err
	}

	return nil
}

func (self *NginxHttp01Provider) CleanUp(domain, token, keyAuth string) error {
	return nil
}

func ObtainCertificate(
	controller *NginxController,
	domains []string,
	email string,
) error {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	user := &LetEntryptUser{
		email: email,
		key:   privateKey,
	}

	config := lego.NewConfig(user)
	//config.CADirURL = lego.LEDirectoryStaging
	client, err := lego.NewClient(config)

	// create an account with the email address provided
	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return err
	}
	user.registration = reg

	err = client.Challenge.SetHTTP01Provider(&NginxHttp01Provider{controller: controller})
	assert.ErrNil(err)

  _, err = client.Certificate.Obtain(certificate.ObtainRequest{
		Domains: domains,
		Bundle:  true,
	})
	if err != nil {
		return err
	}

	//client.Certificate.Renew()
	//certificate, privateKey := certificates.Certificate, certificates.PrivateKey

	return nil
}

