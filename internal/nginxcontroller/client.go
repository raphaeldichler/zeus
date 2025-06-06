// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package nginxcontroller

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/raphaeldichler/zeus/internal/assert"
	"github.com/raphaeldichler/zeus/internal/log"
)

const (
	DefaultClientTimeout = 5 * time.Second
)

type Client struct {
	client *http.Client
	log    *log.Logger
}

func NewClient(
	application string,
	socketPath string,
) *Client {
	logger := log.New(
		application, "nginx-controller-client",
	)

	dialer := func(_ context.Context, _ string, _ string) (net.Conn, error) {
		return net.Dial("unix", socketPath)
	}

	client := &http.Client{
		Transport: &http.Transport{DialContext: dialer},
		Timeout:   DefaultClientTimeout,
	}

	return &Client{
		client: client,
		log:    logger,
	}
}

func (self *Client) SetAcme(
	domain string,
	token string,
	keyAuth string,
) (*http.Response, error) {
	request := AcmeCreateRequest{
		Domain:  domain,
		Token:   token,
		KeyAuth: keyAuth,
	}

	b, err := json.Marshal(request)
	assert.ErrNil(err)
	body := bytes.NewReader(b)

	req, err := http.NewRequest("POST", DeleteAcmeAPIPath, body)
	assert.ErrNil(err)
	req.Header.Set("content-type", "application/json")

	return self.client.Do(req)
}

func (self *Client) DeleteAcme(
	domain string,
	token string,
) (*http.Response, error) {
	request := AcmeDeleteRequest{
		Domain: domain,
		Token:  token,
	}

	b, err := json.Marshal(request)
	assert.ErrNil(err)
	body := bytes.NewReader(b)

	req, err := http.NewRequest("DELETE", DeleteAcmeAPIPath, body)
	assert.ErrNil(err)
	req.Header.Set("content-type", "application/json")

	return self.client.Do(req)
}

func (self *Client) SetConfig(
	request *ApplyRequest,
) (*http.Response, error) {
	b, err := json.Marshal(request)
	assert.ErrNil(err)
	body := bytes.NewReader(b)

	req, err := http.NewRequest("POST", DeleteAcmeAPIPath, body)
	assert.ErrNil(err)
	req.Header.Set("content-type", "application/json")

	return self.client.Do(req)
}
