// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package ingress

import (
	"github.com/raphaeldichler/zeus/internal/ingress/errtype"
	"github.com/raphaeldichler/zeus/internal/nginxcontroller"
	"github.com/raphaeldichler/zeus/internal/record"
	"github.com/raphaeldichler/zeus/internal/runtime"
)

const ZeusRootPath string = "/run/zeus/"

// Ensures that an ingress container is create and running.
//
// If no container exists a new one is created and started.
// On fatal errors no container will be set and the application state will be updated accrodingly.
//
// Note: It is assumed that an existing path to the socket mount exists, if a new container must be created.
func SelectOrCreateIngressContainer(
	state *record.ApplicationRecord,
) (container *runtime.Container, ok bool) {
	defer func() {
		if container == nil {
			ok = false
			return
		}

		if ok := nginxcontroller.ValidateContainer(container, state); ok {
			return
		}

		if err := container.Shutdown(); err != nil {
			state.Ingress.Errors.SetIngressError(
				errtype.FailedInteractionWithDockerDaemon(errtype.DockerStopContainer, err),
			)
		}
		container = nil
		ok = false
	}()

	container, err := runtime.TrySelectOneContainer(
		state.Metadata.Application,
		runtime.ObjectTypeLabel(runtime.IngressObject),
		runtime.ObjectImageLabel(state.Ingress.Metadata.Image),
		runtime.ApplicationNameLabel(state.Metadata.Application),
	)
	if err != nil {
		state.Ingress.Errors.SetIngressError(
			errtype.FailedInteractionWithDockerDaemon(errtype.DockerSelectContainer, err),
		)
	}
	if container != nil {
		return container, true
	}

	container, ok = nginxcontroller.CreateContainer(state)
	if !ok {
		return nil, false
	}

	return container, true
}

func SelectIngressContainer(
	state *record.ApplicationRecord,
) (container *runtime.Container, ok bool) {
	container, err := runtime.TrySelectOneContainer(
		state.Metadata.Application,
		runtime.ObjectTypeLabel(runtime.IngressObject),
		runtime.ObjectImageLabel(state.Ingress.Metadata.Image),
		runtime.ApplicationNameLabel(state.Metadata.Application),
	)
	if err != nil {
		state.Ingress.Errors.SetIngressError(
			errtype.FailedInteractionWithDockerDaemon(errtype.DockerSelectContainer, err),
		)
		return nil, false
	}

	return container, true
}
