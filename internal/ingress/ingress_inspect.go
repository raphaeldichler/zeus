// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package ingress

import (
	"encoding/json"
	"io"

	"github.com/raphaeldichler/zeus/internal/record"
)

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

// GET  /v1.0/applications/poseidon/ingress/gateway
// POST /v1.0/applications/poseidon/ingress/gateway
//
// POST /v1.0/applications
// { "name": "poseidon", "deploymentType": "production|development" }
//
// zeus ingress apply -f/--file gateway.yml
// zeus ingress inspect -f/--format [json|pretty]
//
// Inspects the ingress-controller and writes the current state as a json
// object into the write. It is ensured that the bytes writen are a valid
// json object. If the writer throws an error Inspect will return the error
// otherwise no error will occre.
// if an error occure while reading container information an error response will be written
// into the writer
func (self *IngressDaemon) Inspect(state *record.ApplicationRecord, w io.Writer) error {
	ingress := state.Ingress
	if !ingress.Enabled() {
		// error response -
		return nil
	}

	inspect, err := self.container.Inspect()
	if err != nil {
		// error response -
		return nil
	}

	// problem. obtaining current information about the container might error
	// ->

	response := InspectResponse{
		Name:      ingress.Metadata.Name,
		StartTime: ingress.Metadata.CreateTime.String(),
		IP:        inspect.NetworkSettings.IPAddress,
		Container: ContainerInspectResponse{
			ContainerID: inspect.ID,
			Image:       ingress.Metadata.Image,
			ImageID:     inspect.Image,
			State:       inspect.State.Status,
		},
		Servers:      buildServerResponse(state),
		Certificates: buildCertificatResponse(state),
	}

	if err := json.NewEncoder(w).Encode(&response); err != nil {
		return err
	}
	return nil
}
