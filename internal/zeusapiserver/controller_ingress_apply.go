// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package zeusapiserver

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/raphaeldichler/zeus/internal/record"
	bboltErr "go.etcd.io/bbolt/errors"
)

const IngressApplyAPIPath = "/v1.0/applications/{application}/ingress"

func InggressApplyAPIPath(application string) string {
	return strings.Replace(IngressApplyAPIPath, "{application}", application, 1)
}

type IngressApplyRequestBody struct {
	IPv6  bool `json:"ipv6"`
	Rules []struct {
		Host string `json:"host"`
		Tls  struct {
			Enabled          bool   `json:"enabled"`
			CertificateEmail string `json:"certificateEmail"`
		} `json:"tls"`
		Http struct {
			Paths []struct {
				Path     string `json:"path"`
				Matching string `json:"matching"`
				Service  struct {
					Name string `json:"name"`
					Port string `json:"port"`
				} `json:"service"`
			} `json:"paths"`
		} `json:"http"`
	} `json:"rules"`
}

type IngressApplyRequest struct {
	Application string
	IngressApplyRequestBody
}

func PostIngressApplyRequestDecoder(
	w http.ResponseWriter,
	r *http.Request,
	out *IngressApplyRequest,
) error {
	out.Application = r.PathValue("application")

	if err := json.NewDecoder(r.Body).Decode(&out.IngressApplyRequestBody); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return err
	}

	return nil
}

type serverFilter struct {
	servers []*record.ServerRecord
}

func newServerFilter(servers []*record.ServerRecord) *serverFilter {
	return &serverFilter{
		servers: servers,
	}
}

func (self *serverFilter) getTls(host string) *record.TlsRecord {
	for _, server := range self.servers {
		if server.Host == host && server.Tls != nil {
			return server.Tls
		}
	}
	return nil
}

func (self *ZeusController) PostIngressApply(
	w http.ResponseWriter,
	r *http.Request,
	command *IngressApplyRequest,
) {
	err := self.records.tx(
		application(command.Application),
		func(r *record.ApplicationRecord) error {
			ingress := r.Ingress
			serverFilter := newServerFilter(ingress.Servers)
			ingress.Servers = nil

			for _, rule := range command.IngressApplyRequestBody.Rules {
				server := &record.ServerRecord{
					Host: rule.Host,
					IPv6: command.IngressApplyRequestBody.IPv6,
					Tls:  nil,
					HTTP: record.HttpRecord{
						Paths: nil,
					},
				}
				ingress.Servers = append(ingress.Servers, server)

				if rule.Tls.Enabled {
					tls := serverFilter.getTls(rule.Host)
					expires := time.Unix(0, 0)
					state := record.TlsObtain
					var (
						privkeyPem   []byte = nil
						fullchainPem []byte = nil
					)
					if tls != nil {
						expires = tls.Expires
						state = tls.State
						privkeyPem = tls.PrivkeyPem
						fullchainPem = tls.FullchainPem
					}

					tls = &record.TlsRecord{
						CertificateEmail: rule.Tls.CertificateEmail,
						State:            state,
						Expires:          expires,
						PrivkeyPem:       privkeyPem,
						FullchainPem:     fullchainPem,
					}
					server.Tls = tls
				}

				for _, path := range rule.Http.Paths {
					loc := record.PathRecord{
						Path:     path.Path,
						Matching: path.Matching,
						Service:  record.RecordKey(path.Service.Name),
					}
					server.HTTP.Paths = append(server.HTTP.Paths, loc)
				}
			}

			return nil
		},
	)

	if errors.Is(err, bboltErr.ErrBucketNotFound) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
}
