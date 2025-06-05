// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package zeusapiserver

import (
	"encoding/json"
	"net/http"
)

const IngressApplyAPIPath = "/v1.0/applications/{application}/ingress"

type IngressApplyRequestBody struct {
	IPv6  bool `json:"ipv6"`
	Rules []struct {
		Host string `json:"host"`
		Tls  bool   `json:"tls"`
		Http struct {
			Paths []struct {
				Path     string `json:"path"`
				Matching string `json:"matching"`
				Service  struct {
					Name string `json:"name"`
					Port string `json:"port"`
				} `json:"service"`
			} `json:"paths"`
		} `json:"http"`
	} `json:"rules"`
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

func (self *ZeusController) PostIngressApplyRequestValidator(
	w http.ResponseWriter,
	r *http.Request,
	out IngressApplyRequest,
) bool {
	_, ok := self.Applications[out.Application]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return false
	}

	return true
}

func (self *ZeusController) PostIngressApply(
	w http.ResponseWriter,
	r *http.Request,
	command *IngressApplyRequest,
) {

}
