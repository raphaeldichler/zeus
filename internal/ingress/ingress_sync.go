// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package ingress

import (
	"encoding/base64"
	"errors"
	"io"
	"net/http"

	"github.com/raphaeldichler/zeus/internal/assert"
	"github.com/raphaeldichler/zeus/internal/ingress/errtype"
	"github.com/raphaeldichler/zeus/internal/log"
	"github.com/raphaeldichler/zeus/internal/nginxcontroller"
	"github.com/raphaeldichler/zeus/internal/record"
	"github.com/raphaeldichler/zeus/internal/runtime"
)

const (
	IngressDaemonName          = "ingress-daemon"
	IngressContainerDaemonName = "ingress-nginx"
)

type IngressDaemon struct {
	application string

	container   *runtime.Container
	client      *nginxcontroller.Client
	certificate CertificateProvider

	log *log.Logger
}

// Creates a new IngressDaemon which handles the state of the ingress object.
func NewIngress(application string) *IngressDaemon {
	logger := log.New(application, IngressDaemonName)

	return &IngressDaemon{
		application: application,
		client:      nil,
		container:   nil,
		log:         logger,
	}
}

func (self *IngressDaemon) Sync(state *record.ApplicationRecord) {
	self.log.Info("Starting syncing ingress controllers")
	defer self.log.Info("Completed syncing ingress controllers")
	if !state.Ingress.Enabled() {
		return
	}

	if self.container == nil {
		self.container = SelectOrCreateIngressContainer(state, self.application, self.container)
		if self.container == nil {
			return
		}

		self.client = nginxcontroller.NewClient(self.application, hostNginxControllerSocketPath(self.application))
		self.certificate = CertificateProviderBuilder(self, state)
	}
	assert.NotNil(self.certificate, "on container selection/creation the certifgicate provider must be selected   ")
	self.certificate.GenerateCertificates(state)

	response, err := self.client.SetConfig(self.buildIngressConfigRequest(state))
	if err != nil {
		state.Ingress.Errors.SetIngressError(
			errtype.FailedInteractionWithNginxController(errtype.NginxSend, err),
		)
		return
	}
	defer response.Body.Close()
	assert.True(
		response.StatusCode == http.StatusCreated || response.StatusCode == http.StatusInternalServerError,
		"a request cannot result in a bad request or any other problems",
	)

	if response.StatusCode == http.StatusInternalServerError {
		body, _ := io.ReadAll(response.Body)
		err = errors.New(string(body))

		state.Ingress.Errors.SetIngressError(
			errtype.FailedInteractionWithNginxController(errtype.NginxApply, err),
		)
	}
}

func (self *IngressDaemon) buildIngressConfigRequest(state *record.ApplicationRecord) *nginxcontroller.ApplyRequest {
	req := nginxcontroller.NewApplyRequest()
	for _, server := range state.Ingress.Servers {
		if state.Ingress.Errors.ExistsTlsError(errtype.FailedObtainCertificateQuery(server.Host)) {
			continue
		}

		opts := nginxcontroller.NewServerRequestOptions()
		opts.Add(
			nginxcontroller.WithDomain(server.Host),
			nginxcontroller.WithIPv6Enabled(server.IPv6),
		)

		if tls := server.Tls; tls != nil {
			privkeyBase64 := base64.StdEncoding.EncodeToString(tls.PrivkeyPem)
			fullchainBase64 := base64.StdEncoding.EncodeToString(tls.FullchainPem)

			opts.Add(
				nginxcontroller.WithCertificate(
					privkeyBase64, fullchainBase64,
				),
			)
		}

		for _, paths := range server.HTTP.Paths {
			serviceEndpoint := state.Service.GetEndpoint(paths.Service)

			opts.Add(
				nginxcontroller.WithLocation(
					paths.Path,
					paths.Matching,
					serviceEndpoint,
				),
			)
		}

		req.AddServer(opts.Options...)
	}

	return req
}
