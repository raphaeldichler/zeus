// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package ingress

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/raphaeldichler/zeus/internal/assert"
	"github.com/raphaeldichler/zeus/internal/log"
	"github.com/raphaeldichler/zeus/internal/nginxcontroller"
	"github.com/raphaeldichler/zeus/internal/record"
)

const (
	TlsRenewThreshold = time.Hour * 24
)

type IngressControllerManager struct {
  // HTTP client to the nginx controller
  nginxControllerClient *http.Client 

	log log.Logger
}

// setting http over socket
func NewIngress() *IngressControllerManager {
  dialer := func (ctx context.Context, _ string, _ string) (net.Conn, error) {
    return net.Dial("unix", "some path to the socket")
  }

  transport := &http.Transport{DialContext: dialer}
  client := &http.Client{
    Transport: transport,
    Timeout: 5 * time.Second,
  }

  return &IngressControllerManager {
    nginxControllerClient: client,
  }
}


func (self *IngressControllerManager) Sync(state *record.ApplicationRecord) error {
	self.log.Info("Starting IngressControllerManager.Sync() - syncing ingress controllers")
	self.syncTlsCertificates(state)

  req := nginxcontroller.NewApplyRequest()
  for _, server := range state.Ingress.Servers {
    tls := server.Tls
    if tls != nil && tls.State == record.TlsObtain {
      // we skip it, tls sync failed - server 
      // dependency error - tls certificate failed so this fails
      continue
    }

    opts := nginxcontroller.NewServerRequestOptions()
    if server.IPv6 {
      opts.Add(nginxcontroller.WithIPv6())
    }

    if tls != nil {
      opts.Add(
        nginxcontroller.WithCertificate(
          string(tls.PrivkeyPem),
          string(tls.FullchainPem),
        ),
      )
    }

    for _, paths := range server.HTTP.Paths {
      serviceEndpoint := state.Service.GetEndpoint(paths.Service)

      // we need to load the service entpoint by this key paths.Service
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

  response, err := self.nginxControllerClient.Post("/apply", "application/json", nil)


	self.log.Info("Completed syncing ingress controllers")
	return nil
}

func (self *IngressControllerManager) syncTlsCertificates(state *record.ApplicationRecord) {
	for _, server := range state.Ingress.Servers {
		tls := server.Tls
		if tls == nil {
			// no certificate is required for this server
			continue
		}

		if tls.State == record.TlsObtain {
			// obtain a certificate for this domain. we set acme endpoints and sign the certificate

			certBundle, err := ObtainCertificate(
				nil,
				server.Host,
				tls.CertificateEmail,
			)
			if err != nil {
				// failed to obtain the certificate

				continue
			}

			tls.FullchainPem = certBundle.FullchainPem
			tls.PrivkeyPem = certBundle.PrivKeyPem
			tls.State = record.TlsRenew
			// todo: renew time?

			continue
		}

		assert.True(tls.State == record.TlsRenew, "tls must be in renew state")
		if time.Since(tls.Expires) > TlsRenewThreshold {

			continue
		}
	}

}

func (self *IngressControllerManager) Inspect() error {

	return nil
}
