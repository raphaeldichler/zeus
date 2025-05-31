// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package nginxcontroller

import (
	"fmt"
	"path/filepath"
	"encoding/json"
	"net/http"
)

type AcmeCreateRequest struct {
  Domain string `json:"domain"`
  Token string `json:"token"`
  KeyAuth string `json:"KeyAuth"`
}

type AcmeDeleteRequest struct {
  Domain string `json:"domain"`
  Token string `json:"token"`
}

func acmeLocationPath(
  token string,
) string {
  return filepath.Join("/.well-known/acme-challenge/", token)
}

func (self *Controller) Acme(w http.ResponseWriter, r *http.Request) {
  switch r.Method {
  case "POST":
    req := new(AcmeCreateRequest)
    defer r.Body.Close()
    if err := json.NewDecoder(r.Body).Decode(req); err != nil {
      w.WriteHeader(http.StatusBadRequest)
      return
    }
    
    server := self.config.GetOrCreateHttpServerConfig(req.Domain)
    acmeLocation := NewLocationConfig(
      acmeLocationPath(req.Token),
      ExactMatching,
			fmt.Sprintf(`return 200 "%s"`, req.KeyAuth),
			"add_header Content-Type text/plain",
    )
    server.AddLocation(acmeLocation)

    err := openDirectory(func (d directory) error {
      filepath, err := d.storeFile("conf", self.config.Content())
      if err != nil {
        return err
      }

      if err := self.ReloadNginxConfig(filepath); err != nil {
        return err
      }

      return nil
    })
    if err != nil {
      w.Header().Set("Content-Type", "application/json")
      w.WriteHeader(http.StatusInternalServerError)
      json.NewEncoder(w).Encode(map[string]any {
        "error": "internal-error",
        "message": err.Error(),
      })
      return
    }

	  w.WriteHeader(http.StatusCreated)

  case "DELETE":
    req := new(AcmeDeleteRequest)
    defer r.Body.Close()
    if err := json.NewDecoder(r.Body).Decode(req); err != nil {
      w.WriteHeader(http.StatusBadRequest)
      return
    }

    if !self.config.ExistServerConfig(req.Domain) {
      w.WriteHeader(http.StatusBadRequest)
      return
    }

    deleted := self.config.DeleteHttpLocation(
      req.Domain,
      acmeLocationPath(req.Token),
      ExactMatching,
    )
    if !deleted {
      w.WriteHeader(http.StatusBadRequest)
      return
    }

    err := openDirectory(func (d directory) error {
      filepath, err := d.storeFile("conf", self.config.Content())
      if err != nil {
        return err
      }

      if err := self.ReloadNginxConfig(filepath); err != nil {
        return err
      }

      return nil
    })
    if err != nil {
      w.Header().Set("Content-Type", "application/json")
      w.WriteHeader(http.StatusInternalServerError)
      json.NewEncoder(w).Encode(map[string]any {
        "error": "internal-error",
        "message": err.Error(),
      })
      return
    }

	  w.WriteHeader(http.StatusNoContent)

  default:
    w.WriteHeader(http.StatusMethodNotAllowed)
    return
  }
}
