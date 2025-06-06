// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package ingress

import (
	"errors"
	"net/http"

	"github.com/raphaeldichler/zeus/internal/assert"
	"github.com/raphaeldichler/zeus/internal/log"
	"github.com/raphaeldichler/zeus/internal/nginxcontroller"
)

var ErrFailedConfigAcmeConfig = errors.New("")

type NginxChallengeProvider struct {
	client *nginxcontroller.Client
	log    *log.Logger
}

func NewNginxChallengeProvider(
	client *nginxcontroller.Client,
	application string,
) *NginxChallengeProvider {
	logger := log.New(application, "")

	return &NginxChallengeProvider{
		client: client,
		log:    logger,
	}
}

func (self *NginxChallengeProvider) Present(
	domain string,
	token string,
	keyAuth string,
) error {
	response, err := self.client.SetAcme(
		domain, token, keyAuth,
	)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	assert.True(
		response.StatusCode == http.StatusCreated || response.StatusCode == http.StatusInternalServerError,
		"a request cannot result in a bad request or any other problems",
	)

	if response.StatusCode == http.StatusInternalServerError {
		return ErrFailedConfigAcmeConfig
	}

	return nil
}

func (self *NginxChallengeProvider) CleanUp(
	domain string,
	token string,
	_ string,
) error {
	response, err := self.client.DeleteAcme(
		domain, token,
	)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	assert.True(
		response.StatusCode == http.StatusCreated || response.StatusCode == http.StatusInternalServerError,
		"a request cannot result in a bad request or any other problems",
	)

	if response.StatusCode == http.StatusInternalServerError {
		return ErrFailedConfigAcmeConfig
	}

	return nil
}
