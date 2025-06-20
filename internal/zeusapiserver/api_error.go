// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package zeusapiserver

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type BadRequest struct {
	Message string `json:"message"`
}

func replyBadRequest(w http.ResponseWriter, message string, args ...any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	msg := fmt.Sprintf(message, args...)
	json.NewEncoder(w).Encode(BadRequest{Message: msg})
}
