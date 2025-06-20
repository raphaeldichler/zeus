// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package zeusapiserver

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/raphaeldichler/zeus/internal/assert"
	"github.com/raphaeldichler/zeus/internal/ingress"
	"github.com/raphaeldichler/zeus/internal/record"
	bboltErr "go.etcd.io/bbolt/errors"
)

const (
	IngressApplyAPIPath   = "/v1.0/applications/{application}/ingress"
	IngressInspectAPIPath = "/v1.0/applications/{application}/ingress"
)

func InggressApplyAPIPath(application string) string {
	return strings.Replace(IngressApplyAPIPath, "{application}", application, 1)
}

type IngressInspectRequest struct {
	Application string
}

type InspectResponse struct {
	Name         string                       `json:"name"`
	StartTime    string                       `json:"startTime"`
	IP           string                       `json:"ip"`
	Container    ContainerInspectResponse     `json:"container"`
	Servers      []ServerInspectResponse      `json:"servers"`
	Certificates []CertificateInspectResponse `json:"certificates"`
}

type ServerInspectResponse struct {
	Host     string `json:"host"`
	Path     string `json:"path"`
	Backends string `json:"backends"`
}

type CertificateInspectResponse struct {
	Host     string `json:"host"`
	Enabled  bool   `json:"enabled"`
	Status   string `json:"status"`
	Deadline string `json:"deadline"`
	Email    string `json:"email"`
}

type ContainerInspectResponse struct {
	ContainerID string `json:"containerId"`
	Image       string `json:"image"`
	ImageID     string `json:"imageId"`
	State       string `json:"state"`
}

type IngressApplyRequestBody struct {
	IPv6  bool `json:"ipv6" yaml:"ipv6"`
	Rules []struct {
		Host string `json:"host" yaml:"host"`
		Tls  struct {
			Enabled          bool   `json:"enabled" yaml:"enabled"`
			CertificateEmail string `json:"certificateEmail" yaml:"certificateEmail"`
		} `json:"tls" yaml:"tls"`
		Http struct {
			Paths []struct {
				Path     string `json:"path" yaml:"path"`
				Matching string `json:"matching" yaml:"matching"`
				Service  struct {
					Name string `json:"name" yaml:"name"`
					Port string `json:"port" yaml:"port"`
				} `json:"service" yaml:"service"`
			} `json:"paths" yaml:"paths"`
		} `json:"http" yaml:"http"`
	} `json:"rules" yaml:"rules"`
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
	defer self.orchestrator.ping()

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

func buildServerResponse(state *record.ApplicationRecord) []ServerInspectResponse {
	servers := make([]ServerInspectResponse, 0)

	for _, server := range state.Ingress.Servers {
		for _, serverPath := range server.HTTP.Paths {
			path := serverPath.Path
			if serverPath.Matching == record.MatchingExact {
				path = "= " + serverPath.Path
			}

			servers = append(servers, ServerInspectResponse{
				Host:     server.Host,
				Path:     path,
				Backends: state.Service.GetEndpoint(serverPath.Service),
			})
		}
	}

	return servers
}

func buildCertificatResponse(state *record.ApplicationRecord) []CertificateInspectResponse {
	certificaters := make([]CertificateInspectResponse, 0)

	for _, server := range state.Ingress.Servers {
		if server.Tls == nil {
			certificaters = append(certificaters, CertificateInspectResponse{
				Host:     server.Host,
				Enabled:  false,
				Status:   "",
				Deadline: "",
				Email:    "",
			})

		} else {
			status := "Obtain"
			deadline := ""
			if server.Tls.State == record.TlsRenew {
				status = "Renew"
				deadline = server.Tls.Expires.String()
			}

			certificaters = append(certificaters, CertificateInspectResponse{
				Host:     server.Host,
				Enabled:  true,
				Status:   status,
				Deadline: deadline,
				Email:    server.Tls.CertificateEmail,
			})

		}
	}

	return certificaters
}

func GetIngressInspectRequestDecoder(
	w http.ResponseWriter,
	r *http.Request,
	out *IngressInspectRequest,
) error {
	out.Application = r.PathValue("application")
	return nil
}

func (self *ZeusController) GetIngressInspect(
	w http.ResponseWriter,
	r *http.Request,
	command *IngressInspectRequest,
) {
	state, err := self.records.get(application(command.Application))

	i := state.Ingress
	if !i.Enabled() {
		replyBadRequest(w, "Ingress does not exist")
		return
	}

	container, ok := ingress.SelectIngressContainer(state)
	if !ok {
		replyBadRequest(w, "Failed to interact with container daemon")
		return
	}

	inspect, err := container.Inspect()
	if err != nil {
		replyBadRequest(w, "Failed to interact with container daemon. Inspection failed: %v", err)
		return
	}

	response := InspectResponse{
		Name:      i.Metadata.Name,
		StartTime: i.Metadata.CreateTime.String(),
		IP:        inspect.NetworkSettings.IPAddress,
		Container: ContainerInspectResponse{
			ContainerID: inspect.ID,
			Image:       i.Metadata.Image,
			ImageID:     inspect.Image,
			State:       inspect.State.Status,
		},
		Servers:      buildServerResponse(state),
		Certificates: buildCertificatResponse(state),
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(response)
	assert.ErrNil(err)
}
