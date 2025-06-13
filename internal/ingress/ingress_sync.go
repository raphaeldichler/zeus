// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package ingress

import (
	"context"
	"fmt"
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

func Sync(state *record.ApplicationRecord) {
	log := state.Logger("ingress-daemon")

	log.Info("Starting syncing ingress controllers")
	defer log.Info("Completed syncing ingress controllers")
	if !state.Ingress.Enabled() {
		return
	}

	container, ok := SelectOrCreateIngressContainer(state)
	if !ok {
		return
	}
	assert.NotNil(container, "ok result must return valid container")

	client, err := nginxcontroller.NewClient(state.Metadata.Application)
	if err != nil {
		state.Ingress.Errors.SetIngressError(
			errtype.FailedInteractionWithNginxController(errtype.NginxClientConnection, err),
		)
		return
	}

	for _, server := range state.Ingress.Servers {
		tls := server.Tls
		if tls == nil {
			continue
		}
		if tls.State == record.TlsRenew && tls.Expires.Sub(time.Now()) > TlsRenewThreshold {
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()
		resp, err := client.GenerateCertificates(ctx, &nginxcontroller.GenerateCertificateRequest{
      Type: nginxcontroller.GenerateCertificateType_AuthoritySigned,
      CertificateEmail: tls.CertificateEmail, 
      Domain: server.Host,
    })
    if err != nil {
      state.Ingress.Errors.SetTlsError(
        errtype.FailedInteractionWithNginxController(errtype.NginxApply, err),
      )
      continue
    }
    assert.True(resp.Fullchain != "", "a fullchain must exists")
    assert.True(resp.Privkey != "", "a privkey must exists")

    tls.FullchainPem = []byte(resp.Fullchain)
    tls.PrivkeyPem = []byte(resp.Privkey)
    tls.State = record.TlsRenew
		tls.Expires = time.Now().Add(TlsNewRenewThreshold)
	}

  ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
  defer cancel()
  _, err = client.SetIngressConfig(ctx, buildIngressConfigRequest(state))
  if err != nil {
    state.Ingress.Errors.SetIngressError(
      errtype.FailedInteractionWithNginxController(errtype.NginxApply, err),
    )
  }
}

func buildIngressConfigRequest(state *record.ApplicationRecord) *nginxcontroller.IngressRequest {
  req := nginxcontroller.NewIngressRequestBuilder()

  req.AddEventEntries("worker_connections 1024")
  req.AddGeneralEntries(
    "worker_processes 1",
    "pid /run/nginx.pid",
    "user nginx",
  )
  req.AddHttpEntries(
    "include /etc/nginx/mime.types",
    "default_type application/octet-stream",
    "keepalive_timeout 65",
    "sendfile on",
    "gzip on",
  )

	for _, server := range state.Ingress.Servers {
		if state.Ingress.Errors.ExistsTlsError(errtype.FailedObtainCertificateQuery(server.Host)) {
			continue
		}

    s := req.AddServer(
      server.Host,
      server.IPv6,
    )

    if tls := server.Tls; tls != nil {
      s.AddTLS(
        string(tls.FullchainPem),
        string(tls.PrivkeyPem),
      )
    }

    fmt.Println(server)
    for _, loc := range server.HTTP.Paths {
      matching := nginxcontroller.Matching_Prefix
      if nginxcontroller.Matching_Exact.String() == loc.Matching {
        matching = nginxcontroller.Matching_Exact
      }

      s.AddLocation(
        loc.Path,
        matching,
        "return 200 'http-content'",
        "add_header Content-Type text/plain",
      )
    }
	}

	return req.Build()
}
