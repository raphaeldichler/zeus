// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package nginxcontroller

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	ErrorType    string `json:"error_type"`
	ErrorMessage string `json:"error_message"`
	DebugInfo    string `json:"debug_info,omitempty"`
	Details      any    `json:"details,omitempty"`
}

func reply(
	w http.ResponseWriter,
	status int,
	errorResponse *ErrorResponse,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(errorResponse)
}

func replyInternalServerError(
	w http.ResponseWriter,
	message string,
) {
	reply(
		w,
		http.StatusInternalServerError,
		&ErrorResponse{
			ErrorType:    "internal-error",
			ErrorMessage: message,
		},
	)
}

func replyBadRequest(
	w http.ResponseWriter,
	errorResponse *ErrorResponse,
) {
	reply(
		w,
		http.StatusBadRequest,
		errorResponse,
	)
}
