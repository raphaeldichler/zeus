// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package zeusapiserver

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"unicode"

	"github.com/raphaeldichler/zeus/internal/assert"
	"github.com/raphaeldichler/zeus/internal/record"
)

var ErrBadRequestApplication = errors.New("bad request: application")

const (
	createApplicationAPIPath     string = "/v1.0/applications"
)

func CreateApplicationAPIPath() string {
	return createApplicationAPIPath
}

type CreateApplicationRequest struct {
	Application    application
	DeploymentType record.DeploymentType
}

type JsonCreateApplicationRequest struct {
	Application    string `json:"application"`
	DeploymentType string `json:"deploymentType"`
}

func NewCreateApplicationRequestAsJsonBody(
  application string,
  deploymentType string,
) io.Reader {
  data := &JsonCreateApplicationRequest{
		Application:    application,
		DeploymentType: deploymentType,
	}

  jsonData, err := json.Marshal(data)
  assert.ErrNil(err)

  return bytes.NewReader(jsonData)
}

func isLetter(s string) bool {
	return !strings.ContainsFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r)
	})
}

func (self *ApplicationController) DecoderCreateApplicationRequest(
	w http.ResponseWriter,
	r *http.Request,
	out *CreateApplicationRequest,
) error {
	jsonRequest := new(JsonCreateApplicationRequest)

	if err := json.NewDecoder(r.Body).Decode(jsonRequest); err != nil {
		self.logger.Error("Failed to decode JSON body: %v", err)
		replyBadRequest(w, "Invalid JSON payload")
		return err
	}

	if err := decodeApplicationName(jsonRequest.Application, w); err != nil {
		self.logger.Error("Invalid application name: %q", jsonRequest.Application)
		return err
	}

	switch jsonRequest.DeploymentType {
	case "production":
		out.DeploymentType = record.Production

	case "development":
		out.DeploymentType = record.Development

	default:
		self.logger.Error("Invalid deployment type: %q", jsonRequest.DeploymentType)
		replyBadRequest(w, "Deployment type must be either 'production' or 'development'")
		return ErrBadRequestApplication
	}

	out.Application = application(jsonRequest.Application)
	return nil
}

func (self *ApplicationController) CreateApplication(
	w http.ResponseWriter,
	r *http.Request,
	command *CreateApplicationRequest,
) {
	assert.True(command.Application.valid(), "decoder must validate the application")

	app := string(command.Application)
	self.logger.Info("Received request to create application: %q", app)

	err := self.records.add(command.Application, command.DeploymentType)
	if err != nil {
		self.logger.Error("Application creation failed for %q: already exists", app)
		replyBadRequest(w, "Application already exists.")
		return
	}

	self.logger.Info("Application %q created successfully", app)
	w.WriteHeader(http.StatusCreated)
	w.Header().Set(
		"Location",
		strings.Replace(inspectApplicationAPIPath, "{application}", app, 1),
	)
}
