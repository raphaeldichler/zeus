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
	"github.com/raphaeldichler/zeus/internal/nginxcontroller"
	"github.com/raphaeldichler/zeus/internal/record"
)

func Sync(state *record.ApplicationRecord) {
	log := state.Logger("ingress-daemon")

	log.Info("Starting syncing ingress controllers")
	defer log.Info("Completed syncing ingress controllers")
	if !state.Ingress.Enabled() {
		return
	}

	container := SelectOrCreateIngressContainer(state)
	if container == nil {
		return
	}

	client := nginxcontroller.NewClient(
		state.Metadata.Application, hostNginxControllerSocketPath(state.Metadata.Application),
	)
	certificate := CertificateProviderBuilder(state)
	assert.NotNil(certificate, "on container selection/creation the certifgicate provider must be selected")
	certificate.GenerateCertificates(state)

	response, err := client.SetConfig(buildIngressConfigRequest(state))
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

func buildIngressConfigRequest(state *record.ApplicationRecord) *nginxcontroller.ApplyRequest {
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
