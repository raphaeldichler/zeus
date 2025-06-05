// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package zeusapiserver

import "net/http"

const IngressInspectAPIPath = "/v1.0/applications/{application}/ingress"

type IngressInspectRequest struct {
	Application string
}

func GetIngressInspectRequestDecoder(
	w http.ResponseWriter,
	r *http.Request,
	out *IngressInspectRequest,
) error {
	out.Application = r.PathValue("application")
	return nil
}

func (self *ZeusController) GetIngressInspectRequestValidation(
	w http.ResponseWriter,
	r *http.Request,
	out IngressInspectRequest,
) bool {
	_, ok := self.Applications[out.Application]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return false
	}

	return true
}

func (self *ZeusController) GetIngressInspect(
	w http.ResponseWriter,
	r *http.Request,
	command *IngressInspectRequest,
) {
	// so we need to do fetch the state?
}
