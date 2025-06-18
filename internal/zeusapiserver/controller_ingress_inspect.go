// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package zeusapiserver

import (
	"net/http"

	"github.com/raphaeldichler/zeus/internal/ingress"
	"github.com/raphaeldichler/zeus/internal/record"
)

const IngressInspectAPIPath = "/v1.0/applications/{application}/ingress"

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
		// error response -
		return
	}

	container, ok := ingress.SelectIngressContainer(state)
	if !ok {
		return
	}

	inspect, err := container.Inspect()
	if err != nil {
		// error response -
		return
	}

	// problem. obtaining current information about the container might error
	// ->
	_ = InspectResponse{
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

}
