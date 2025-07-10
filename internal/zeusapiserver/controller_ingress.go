// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package zeusapiserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/raphaeldichler/zeus/internal/assert"
	"github.com/raphaeldichler/zeus/internal/ingress"
	"github.com/raphaeldichler/zeus/internal/ingress/errtype"
	"github.com/raphaeldichler/zeus/internal/record"
	bboltErr "go.etcd.io/bbolt/errors"
)

const (
	ingressApplyAPIPath   = "/v1.0/applications/{application}/ingress"
	ingressInspectAPIPath = "/v1.0/applications/{application}/ingress"
)

func IngressApplyAPIPath(apiVersion string, application string) string {
	switch apiVersion {
	case "v1.0":
		return strings.Replace(ingressApplyAPIPath, "{application}", application, 1)
	default:
		assert.Unreachable("cover all cases of api version")
	}
	return ""
}

func IngressInspectAPIPath(apiVersion string, application string) string {
	switch apiVersion {
	case "v1.0":
		return strings.Replace(ingressInspectAPIPath, "{application}", application, 1)
	default:
		assert.Unreachable("cover all cases of api version")
	}
	return ""
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
	Errors       []IngressErrorEntryRecord    `json:"errors"`
}

type IngressErrorEntryRecord struct {
	Type       string `json:"type"`
	Identifier string `json:"identifier"`
	Message    string `json:"message"`
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

	fmt.Println(command)

	err := self.records.tx(
		application(command.Application),
		func(r *record.ApplicationRecord) error {
			ingress := r.Ingress
			if ingress == nil {
				ingress = record.NewIngressRecord()
				r.Ingress = ingress
			}

			serverFilter := newServerFilter(ingress.Servers)
			ingress.Servers = nil

			for _, rule := range command.IngressApplyRequestBody.Rules {
				fmt.Println(rule.Host)
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
					fmt.Printf("%s - %s - %s\n", path.Path, path.Matching, path.Service.Name)
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

			fmt.Print("host", server.Host)
			fmt.Println(" | path", path)
			servers = append(servers, ServerInspectResponse{
				Host:     server.Host,
				Path:     path,
				Backends: state.Service.GetEndpoint(serverPath.Service),
			})
		}
	}

	fmt.Println("servers")
	fmt.Println(servers)

	return servers
}

func buildCertificatResponse(state *record.ApplicationRecord) []CertificateInspectResponse {
	certificaters := make([]CertificateInspectResponse, 0)

	for _, server := range state.Ingress.Servers {
		if server.Tls == nil {
			certificaters = append(certificaters, CertificateInspectResponse{
				Host:     server.Host,
				Enabled:  false,
				Status:   "-",
				Deadline: "-",
				Email:    "-",
			})

		} else {
			status := "Obtain"
			deadline := "-"
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

func buildErrorResponse(state *record.ApplicationRecord) []IngressErrorEntryRecord {
	errors := make([]IngressErrorEntryRecord, 0)

	for _, err := range state.Ingress.Errors {
		errors = append(errors, IngressErrorEntryRecord{
			Type:       err.Type,
			Identifier: err.Identifier,
			Message:    err.Message,
		})
	}

	return errors
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
	fmt.Println(command)
	state, err := self.records.get(application(command.Application))
	if err != nil {
		replyBadRequest(w, "Application is not enabled")
		return
	}

	i := state.Ingress
	if !i.Enabled() {
		replyBadRequest(w, "Ingress does not exist")
		return
	}

	response := InspectResponse{
		Name:      "ingress",
		StartTime: i.Metadata.CreateTime.String(),
		IP:        "-",
		Container: ContainerInspectResponse{
			ContainerID: "-",
			Image:       i.Metadata.Image,
			ImageID:     "-",
			State:       "Not Created",
		},
		Servers:      buildServerResponse(state),
		Certificates: buildCertificatResponse(state),
		Errors:       buildErrorResponse(state),
	}

	container, ok := ingress.SelectIngressContainer(state)
	if !ok {
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(response)
		assert.ErrNil(err)
		return
	}

	inspect, err := container.Inspect()
	if err != nil {
		state.Ingress.SetError(
			errtype.FailedInteractionWithDockerDaemon(errtype.DockerInspectContainer, err),
		)
	} else {
		response.IP = inspect.NetworkSettings.IPAddress
		response.Container.State = inspect.State.Status
		response.Container.ContainerID = inspect.ID
		response.Container.ImageID = inspect.Image
	}

	response.Errors = buildErrorResponse(state)

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(response)
	assert.ErrNil(err)
}
