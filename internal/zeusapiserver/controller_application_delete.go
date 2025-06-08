// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package zeusapiserver

import (
	"net/http"

	"github.com/raphaeldichler/zeus/internal/assert"
)

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
