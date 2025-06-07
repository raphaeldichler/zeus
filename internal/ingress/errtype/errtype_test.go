// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package errtype

import (
	"errors"
	"testing"
)

func TestIngressErrtypeDockerDaemonInteraction(t *testing.T) {
	errorMsg := "Testing message"
	err := errors.New(errorMsg)

	types := []struct {
		id   DockerDaemonInteraction
		name string
	}{
		{id: DockerCreate, name: "create"},
		{id: DockerSelect, name: "select"},
		{id: DockerStop, name: "stop"},
		{id: DockerInspect, name: "inspect"},
	}

	for _, tt := range types {
		t.Run(tt.name, func(t *testing.T) {
			record := FailedInteractionWithDockerDaemon(tt.id, err)
			if record.Message != errorMsg {
				t.Errorf("expected message to be '%s', got '%s'", errorMsg, record.Message)
			}
		})
	}
}

func TestIngressErrtypeNginxControllerInteraction(t *testing.T) {
	errorMsg := "Testing message"
	err := errors.New(errorMsg)

	types := []struct {
		id   NginxControllerInteraction
		name string
	}{
		{id: NginxSend, name: "send"},
		{id: NginxApply, name: "apply"},
	}

	for _, tt := range types {
		t.Run(tt.name, func(t *testing.T) {
			record := FailedInteractionWithNginxController(tt.id, err)
			if record.Message != errorMsg {
				t.Errorf("expected message to be '%s', got '%s'", errorMsg, record.Message)
			}
		})
	}
}
