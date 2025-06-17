// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package zeusapiserver

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/raphaeldichler/zeus/internal/log"
	"go.etcd.io/bbolt"
)

var RecordKey = []byte("record")

const (
	inspectApplicationAPIPath    string = "/v1.0/applications/{application}"
	inspectAllApplicationAPIPath string = "/v1.0/applications"
)

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
