// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package ingress

import (
	"fmt"
	"path/filepath"
	"slices"
	"time"

	"github.com/raphaeldichler/zeus/internal/log"
	"github.com/raphaeldichler/zeus/internal/runtime"
)

type NginxController struct {
	container *runtime.Container
	cache     *NginxCache

	log *log.Logger
}

func NewNginxController(
	container *runtime.Container,
	application string,
) *NginxController {
	log := log.New(application, "nginx-controller")

	return &NginxController{
		container: container,
		cache:     NewNginxCache(),
		log:       log,
	}
}

func (self *NginxController) FindObsoleteServers(servers []ServerIdentifier) []ServerIdentifier {
  result := make([]ServerIdentifier, 0)
  for _, serverID := range self.cache.serverKeys() {
    if !slices.Contains(servers, serverID) {
      result = append(result, serverID)
    }
  }

  return result
}

func (self *NginxController) FindObsoleteLocations(servers []LocationIdentifier) []LocationIdentifier {
  result := make([]LocationIdentifier, 0)
  for _, locationID := range self.cache.locationKeys() {
    if !slices.Contains(servers, locationID) {
      result = append(result, locationID)
    }
  }

  return result
}

func (self *NginxController) SetAcmeChallengeLocation(
	domain string,
	token string,
	keyAuth string,
) error {
	self.log.Info("Setting ACME-Challenge location for domain '%s' (containerID: %s)", domain, self.container)

	return self.SetLocation(&LocationConfig{
		LocationIdentifier: LocationIdentifier{
			ServerIdentifier: ServerIdentifier{
				Domain:     domain,
				TlsEnabled: false,
        IPv6: false,
			},
			Path:     filepath.Join("/.well-known/acme-challenge/", token),
			Matching: LocationExact,
		},
		Entries: []string{
			fmt.Sprintf(`return 200 "%s"`, keyAuth),
			"add_header Content-Type text/plain",
		},
	})
}

func (self *NginxController) SetLocation(
	location *LocationConfig,
) error {
	if err := self.container.AssertPathExists(location.LocationDirectory()); err != nil {
		return err
	}

	if self.cache.isLocationCached(location) {
		return nil
	}

	if err := self.container.CopyInto(location); err != nil {
		self.log.Error("Failed to copy location config into container (%s)", self.container)
		return err
	}
	self.cache.set(location)

	return nil
}

func (self *NginxController) UnsetLocation(
	locationId *LocationIdentifier,
) error {
	if self.cache.isKeyCached(locationId.FilePath()) {
		if err := self.container.RemoveFile(locationId.FilePath()); err != nil {
			return err
		}

		self.cache.unset(locationId.FilePath())
	}

	return nil
}

func (self *NginxController) SetHTTPServer(
	cfg *ServerConfig,
) error {
	if err := self.container.AssertPathExists(NginxInternalServerPath); err != nil {
		return err
	}

	if self.cache.isCached(cfg) {
		return nil
	}

	serverLocationPath := cfg.LocationDirectory()
	if err := self.container.EnsurePathExists(serverLocationPath); err != nil {
		return err
	}

	if cfg.TlsEnabled {
		certificateId := CertificateIdentifier{cfg.Domain}

		if err := self.container.AssertPathExists(certificateId.PublicKeyPath()); err != nil {
			return err
		}
		if err := self.container.AssertPathExists(certificateId.PrivateKeyPath()); err != nil {
			return err
		}

		cfg.Entries = append(cfg.Entries, fmt.Sprintf("ssl_certificate %s", certificateId.PublicKeyPath()))
		cfg.Entries = append(cfg.Entries, fmt.Sprintf("ssl_certificate_key %s", certificateId.PrivateKeyPath()))
	}

	if err := self.container.CopyInto(cfg); err != nil {
		self.log.Error("Failed to copy server config into container (%s)", self.container)
		return err
	}
	self.cache.set(cfg)

	return nil
}

func (self *NginxController) UnsetCertificate(
	certificateId CertificateIdentifier,
) error {
	if self.cache.isKeyCached(certificateId.PrivateKeyPath()) {
		if err := self.container.RemoveFile(certificateId.PrivateKeyPath()); err != nil {
			return err
		}

		self.cache.unset(certificateId.PrivateKeyPath())
	}

	if self.cache.isKeyCached(certificateId.PublicKeyPath()) {
		if err := self.container.RemoveFile(certificateId.PublicKeyPath()); err != nil {
			return err
		}

		self.cache.unset(certificateId.PublicKeyPath())
	}

	return nil
}

func (self *NginxController) SetCertificate(
	certificate *Certificate,
) error {
	if err := self.container.AssertPathExists(NginxInternalCertificatePath); err != nil {
		return err
	}

	if err := self.container.EnsurePathExists(certificate.DirectoryPath()); err != nil {
		return err
	}

	if !self.cache.isCached(certificate.PrivateKey()) {
		if err := self.container.CopyInto(certificate.PrivateKey()); err != nil {
			self.log.Error("Failed to copy private certificates into container (%s)", self.container)
			return err
		}
		self.cache.set(certificate.PrivateKey())
	}

	if !self.cache.isCached(certificate.PublicKey()) {
		if err := self.container.CopyInto(certificate.PublicKey()); err != nil {
			self.log.Error("Failed to copy public certificates into container (%s)", self.container)
			return err
		}
		self.cache.set(certificate.PublicKey())
	}

	return nil
}

func (self *NginxController) ApplyConfig() error {
	err := self.container.Sighup()
  if err != nil {
	  return err
  }
	// todo: just a hack for the moment, add a verificiation step to ensure is applied
	time.Sleep(time.Second)
	return nil
}

func (self *NginxController) Transaction(tx func(*NginxController) error) error {
  return self.cache.transaction(func(c *NginxCache) error {
		ctr := NginxController{
			container: self.container,
			cache:     c,
			log:       self.log,
		}

    if err := tx(&ctr); err != nil {
      return err
    }

    if err := ctr.ApplyConfig(); err != nil {
      return err
    }

    return nil
	})
}
