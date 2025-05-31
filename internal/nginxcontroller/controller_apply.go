// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package nginxcontroller

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/raphaeldichler/zeus/internal/assert"
)

var (
  matching map[string]MatchingType = map[string]MatchingType {
    "exact": ExactMatching,
    "prefix": PrefixMatching,
  }
)

type ApplyRequest struct {
  Servers []ServerRequest `json:"servers"`
}

type ServerRequest struct {
  Domain string `json:"domain"`
  Certificate *CertificateRequest `json:"certificate"`
  Locations []LocationRequest `json:"locations"`
  IPv6Enabled bool `json:"ipv6Enabled"`
}

type LocationRequest struct {
  Path string `json:"path"`
  Matching string `json:"matching"`
  ServiceEndpoint string `json:"serviceEndpoint"`
}

type CertificateRequest struct {
  PrivkeyPem string `json:"privkeyPem"`
  FullchainPem string `json:"fullchainPem"`
}

func (self *CertificateRequest) StoreCertificate(d directory) (*TlsCertificate, error) {
  b := make([]byte, 16)
  rand.Read(b)
  filename :=  hex.EncodeToString(b)

  fullchainPem, err := base64.StdEncoding.DecodeString(self.FullchainPem)
  assert.ErrNil(err)
  fullchainPath, err := d.store(filename + ".public.pem", fullchainPem)
  if err != nil {
    return nil, err
  }

  privkeyPem, err := base64.StdEncoding.DecodeString(self.PrivkeyPem)
  assert.ErrNil(err)
  privkeyPath, err := d.store(filename + ".private.pem", privkeyPem)
  if err != nil {
    return nil, err
  }

  return &TlsCertificate{
    FullchainFilePath: fullchainPath,
    PrivkeyFilePath: privkeyPath,
  }, nil
}

func (self *Controller) Apply(w http.ResponseWriter, r *http.Request) {
  if r.Method != "POST" {
    w.WriteHeader(http.StatusMethodNotAllowed)
    return
  }

  req := new(ApplyRequest)
  defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
    return
	}

  err := openDirectory(func (d directory) error {
    cfg := NewNginxConfig()
    for _, server := range req.Servers {
      var (
        tls *TlsCertificate
        err error
      )

      if cert := server.Certificate; cert != nil {
        tls, err = cert.StoreCertificate(d)
        if err != nil {
          return err
        }
      }

      sc := NewServerConfig(
        server.Domain,
        server.IPv6Enabled,
        tls,
      )
      cfg.AddServerConfig(sc)
    
      for _, loc := range server.Locations {
        m, ok := matching[loc.Matching]
        assert.True(ok, "matching type must already be validated")

        lc := LocationsConfig {
          Path: loc.Path,
          Matching: m,
          ServiceEndpoint: loc.ServiceEndpoint,
        }

        sc.AddLocation(lc)
      }
    }

    cfgPath, err := d.storeFile("conf", cfg.Content())
    if err != nil {
      return err
    }

    if err := self.ReloadNginxConfig(cfgPath); err != nil {
      return err
    }

    return nil
  })
  if err != nil  {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusInternalServerError)
    json.NewEncoder(w).Encode(map[string]any {
      "error": "internal-error",
      "message": err.Error(),
    })
    return
  }

	w.WriteHeader(http.StatusCreated)
}
