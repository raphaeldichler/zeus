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
	"github.com/raphaeldichler/zeus/internal/log"
	"github.com/raphaeldichler/zeus/internal/record"
	"go.etcd.io/bbolt"
	bboltErr "go.etcd.io/bbolt/errors"
)

var (
	RecordKey                = []byte("record")
	ErrBadRequestApplication = errors.New("bad request: application")
)

const (
	createApplicationAPIPath     string = "/v1.0/applications"
	deleteApplicationAPIPath     string = "/v1.0/applications/{application}"
	enableApplicationAPIPath     string = "/v1.0/applications/{application}/enable"
	disableApplicationAPIPath    string = "/v1.0/applications/{application}/disable"
	inspectApplicationAPIPath    string = "/v1.0/applications/{application}"
	inspectAllApplicationAPIPath string = "/v1.0/applications"
)

func CreateApplicationAPIPath() string {
	return createApplicationAPIPath
}

func DeleteApplicationAPIPath(application string) string {
	return strings.Replace(deleteApplicationAPIPath, "{application}", application, 1)
}

func InspectApplicationAPIPath(application string) string {
	return strings.Replace(inspectApplicationAPIPath, "{application}", application, 1)
}

func InspectAllApplicationAPIPath() string {
	return inspectAllApplicationAPIPath
}

type ApplicationController struct {
	records *RecordCollection
	logger  *log.Logger
}

func NewApplication(records *RecordCollection) *ApplicationController {
	return &ApplicationController{
		records: records,
		logger:  log.New("zeus", "application-controller"),
	}
}

func decodeApplicationName(application string, w http.ResponseWriter) error {
	if strings.TrimSpace(application) != application {
		replyBadRequest(w, "Application name must not have leading or trailing whitespace")
		return ErrBadRequestApplication
	}

	if len(application) == 0 {
		replyBadRequest(w, "Application name must not be empty")
		return ErrBadRequestApplication
	}

	if len(application) >= bbolt.MaxKeySize {
		replyBadRequest(w, fmt.Sprintf("Application name exceeds maximum allowed length of %d", bbolt.MaxKeySize))
		return ErrBadRequestApplication
	}

	if !isLetter(application) {
		replyBadRequest(w, "Application name must contain only letters")
		return ErrBadRequestApplication
	}

	return nil
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
}

type DeleteApplicationRequest struct {
	Application application
}

func (self *ApplicationController) DecoderDeleteApplicationRequest(
	w http.ResponseWriter,
	r *http.Request,
	out *DeleteApplicationRequest,
) error {
	a := r.PathValue("application")

	if err := decodeApplicationName(a, w); err != nil {
		self.logger.Error("Invalid application name in delete request: %q", a)
		return err
	}

	out.Application = application(a)
	return nil
}

func (self *ApplicationController) DeleteApplication(
	w http.ResponseWriter,
	r *http.Request,
	command *DeleteApplicationRequest,
) {
	assert.True(command.Application.valid(), "decoder must validate the application")

	app := string(command.Application)
	self.logger.Info("Received request to delete application: %q", app)

	err := self.records.delete(command.Application)
	if err != nil {
		self.logger.Error("Failed to delete application %q: not found", app)
		replyBadRequest(w, "Cannot delete application, because not found.")
		return
	}

	self.logger.Info("Application %q deleted successfully", app)
	w.WriteHeader(http.StatusNoContent)
}

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

func EnableApplicationAPIPath(applicaiton string) string {
	return strings.Replace(enableApplicationAPIPath, "{application}", applicaiton, 1)
}

func DisableApplicationAPIPath(applicaiton string) string {
	return strings.Replace(disableApplicationAPIPath, "{application}", applicaiton, 1)
}

type DisableApplicationRequest struct {
	Application application
}

type EnableApplicationRequest struct {
	Application application
}

func (self *ApplicationController) DecoderEnableApplicationRequest(
	w http.ResponseWriter,
	r *http.Request,
	out *EnableApplicationRequest,
) error {
	a := r.PathValue("application")

	if err := decodeApplicationName(a, w); err != nil {
		self.logger.Error("Invalid application name in enable request: %q", a)
		return err
	}

	out.Application = application(a)
	return nil
}

func (self *ApplicationController) EnableApplication(
	w http.ResponseWriter,
	r *http.Request,
	command *EnableApplicationRequest,
) {
	appName := string(command.Application)
	err := self.records.enableIfNonElse(command.Application)

	switch {
	case errors.Is(err, ErrApplicationEnabled):
		self.logger.Error("Enable failed: application %q cannot be enabled because other is this enabled", appName)
		replyBadRequest(w, "An application is already enabled.")
		return

	case errors.Is(err, bboltErr.ErrBucketNotFound):
		self.logger.Error("Enable failed: application not found: %q", appName)
		replyBadRequest(w, "Application does not exist")
		return

	case err != nil:
		assert.Unreachable("cover all cases of enableIfNonElse")
	}

	self.logger.Info("Application enabled: %q", appName)
	w.WriteHeader(http.StatusNoContent)
}

func (self *ApplicationController) DecoderDisableApplicationRequest(
	w http.ResponseWriter,
	r *http.Request,
	out *DisableApplicationRequest,
) error {
	a := r.PathValue("application")

	if err := decodeApplicationName(a, w); err != nil {
		self.logger.Error("Invalid application name in disable request: %q", a)
		return err
	}

	out.Application = application(a)
	return nil
}

func (self *ApplicationController) DisableApplication(
	w http.ResponseWriter,
	r *http.Request,
	out *DisableApplicationRequest,
) {
	appName := string(out.Application)
	err := self.records.tx(out.Application, func(rec *record.ApplicationRecord) error {
		rec.Metadata.Enabled = false
		return nil
	})
	if err != nil {
		self.logger.Error("Disable failed: application not found: %q", appName)
		replyBadRequest(w, "Application does not exist")
		return
	}

	self.logger.Info("Application disabled: %q", appName)
	w.WriteHeader(http.StatusNoContent)
}
