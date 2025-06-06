// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package nginxcontroller

import (
	"fmt"
	"net/http"
	"path/filepath"
	"regexp"
)

const (
	DeleteAcmeAPIPath = "/acme"
	SetAcmeAPIPath    = "/acme"
)

type ValidatableRequest interface {
	Validate() (*ErrorResponse, bool)
}

func IsValidDomain(domain string) bool {
	var domainRegex = regexp.MustCompile(`^(localhost|([a-zA-Z0-9-]{1,63}\.)+[a-zA-Z]{2,63})$`)
	return domainRegex.MatchString(domain)
}

type AcmeCreateRequest struct {
	Domain  string `json:"domain"`
	Token   string `json:"token"`
	KeyAuth string `json:"KeyAuth"`
}

func (self AcmeCreateRequest) Validate(
	w http.ResponseWriter,
	r *http.Request,
) bool {
	if !IsValidDomain(self.Domain) {
		return false
	}

	return true
}

type AcmeDeleteRequest struct {
	Domain string `json:"domain"`
	Token  string `json:"token"`
}

func (self AcmeDeleteRequest) Validate(
	w http.ResponseWriter,
	r *http.Request,
) bool {
	if !IsValidDomain(self.Domain) {
		return false
	}

	return true
}

func acmeLocationPath(token string) string {
	return filepath.Join("/.well-known/acme-challenge/", token)
}

func (self *Controller) DeleteAcme(
	w http.ResponseWriter,
	r *http.Request,
	command *AcmeDeleteRequest,
) {
	d, err := openDirectory()
	if err != nil {
		replyInternalServerError(w, "Failed to open directory to store data. "+err.Error())
		return
	}
	defer d.close()

	loc := self.config.DeleteHttpLocation(command.Domain, acmeLocationPath(command.Token), ExactMatching)
	if loc == nil {
		replyBadRequest(w, &ErrorResponse{
			ErrorType:    "invalid-delete-location",
			ErrorMessage: "The specified location to delete was not found for the given domain and token.",
		})
		return
	}

	if err := self.StoreAndApplyConfig(w, self.config, d); err != nil {
		self.config.SetHttpLocation(command.Domain, loc)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (self *Controller) SetAcme(
	w http.ResponseWriter,
	r *http.Request,
	command *AcmeCreateRequest,
) {
	d, err := openDirectory()
	if err != nil {
		replyInternalServerError(w, "Failed to open directory to store data. "+err.Error())
		return
	}
	defer d.close()

	self.config.SetHttpLocation(
		command.Domain,
		NewLocationConfig(
			acmeLocationPath(command.Token),
			ExactMatching,
			fmt.Sprintf(`return 200 "%s"`, command.KeyAuth),
			"add_header Content-Type text/plain",
		),
	)

	if err := self.StoreAndApplyConfig(w, self.config, d); err != nil {
		return
	}

	w.WriteHeader(http.StatusCreated)
}
