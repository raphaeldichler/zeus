// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package ingress

import (
	"time"

	"github.com/raphaeldichler/zeus/internal/assert"
	"github.com/raphaeldichler/zeus/internal/log"
	"github.com/raphaeldichler/zeus/internal/record"
)

const (
	TlsRenewThreshold = time.Hour * 24
)

type IngressControllerManager struct {
	log log.Logger
}

func (self *IngressControllerManager) Sync(state *record.ApplicationRecord) error {
	self.log.Info("Starting IngressControllerManager.Sync() - syncing ingress controllers")
  if err := self.syncTlsCertificates(state); err != nil {
    return err
  }



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


			self.log.Info("ACME certificate creation request initialized for domain %q", server.Host)
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
