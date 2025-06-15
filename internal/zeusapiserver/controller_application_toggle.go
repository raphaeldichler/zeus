// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package zeusapiserver

import (
	"errors"
	"net/http"
	"strings"

	"github.com/raphaeldichler/zeus/internal/assert"
	"github.com/raphaeldichler/zeus/internal/record"
	bboltErr "go.etcd.io/bbolt/errors"
)

const (
	enableApplicationAPIPath  string = "/v1.0/applications/{application}/enable"
	disableApplicationAPIPath string = "/v1.0/applications/{application}/disable"
)

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
