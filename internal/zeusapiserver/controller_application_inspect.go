// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package zeusapiserver

import (
	"encoding/json"
	"net/http"

	"github.com/raphaeldichler/zeus/internal/assert"
)

type InspectAllApplicationRequest struct{}

type InspectAllApplicationResponse struct {
	Applications []InspectApplicationResponse `json:"applications"`
}

type InspectApplicationRequest struct {
	Application application
}

type InspectApplicationResponse struct {
	Application    string `json:"application"`
	DeploymentType string `json:"deploymentType"`
	Enabled        bool   `json:"enabled"`
}

func (self *ApplicationController) DecoderInspectApplicationRequest(
	w http.ResponseWriter,
	r *http.Request,
	out *InspectApplicationRequest,
) error {
	a := r.PathValue("application")

	if err := decodeApplicationName(a, w); err != nil {
		self.logger.Error("Invalid application name in inspect request: %q", a)
		return err
	}

	out.Application = application(a)
	return nil
}

func (self *ApplicationController) DecoderInspectAllApplicationRequest(
	w http.ResponseWriter,
	r *http.Request,
	out *InspectAllApplicationRequest,
) error {
	// No validation needed
	return nil
}

func (self *ApplicationController) InspectAllApplication(
	w http.ResponseWriter,
	r *http.Request,
	_ *InspectAllApplicationRequest,
) {
	records := self.records.all()
	response := InspectAllApplicationResponse{}
	self.logger.Info("Inspecting all applications")

	for _, e := range records {
		response.Applications = append(
			response.Applications,
			InspectApplicationResponse{
				Application:    e.Metadata.Application,
				DeploymentType: e.Metadata.Deployment.String(),
				Enabled:        e.Metadata.Enabled,
			},
		)
	}

	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(response)
	assert.ErrNil(err)
}

func (self *ApplicationController) InspectApplication(
	w http.ResponseWriter,
	r *http.Request,
	command *InspectApplicationRequest,
) {
	assert.True(command.Application.valid(), "decoder must validate the application")

	appName := string(command.Application)
	app, err := self.records.get(command.Application)
	if err != nil {
		self.logger.Error("Application not found during inspection: %q", appName)
		replyBadRequest(w, "Application does not exist")
		return
	}

	self.logger.Info("Inspected application: %q", appName)

	err = json.NewEncoder(w).Encode(
		InspectApplicationResponse{
			Application:    app.Metadata.Application,
			DeploymentType: app.Metadata.Deployment.String(),
			Enabled:        app.Metadata.Enabled,
		},
	)
	assert.ErrNil(err)
	w.WriteHeader(http.StatusOK)
}
