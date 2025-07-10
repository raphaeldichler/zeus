// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package ingress

import (
	runtimeErr "github.com/raphaeldichler/zeus/internal/runtime/errtype"
	"github.com/raphaeldichler/zeus/internal/nginxcontroller"
	"github.com/raphaeldichler/zeus/internal/record"
	"github.com/raphaeldichler/zeus/internal/runtime"
	"github.com/raphaeldichler/zeus/internal/util/optional"
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
) optional.Optional[runtime.Container] {
	optionalContainer, err := runtime.TrySelectOneContainer(
		state.Metadata.Application,
		runtime.ObjectTypeLabel(runtime.IngressObject),
		runtime.ApplicationNameLabel(state.Metadata.Application),
	)
	if err != nil {
		state.Ingress.SetError(
      runtimeErr.FailedInteractionWithDockerDaemon(runtimeErr.DockerSelectContainer, err),
		)

		return optional.Empty[runtime.Container]()
	}

	optionalContainer = optionalContainer.IfPresent(func(t *runtime.Container) *runtime.Container {
		if t.Image() == state.Ingress.Metadata.Image {
			return t
		}

		// to enable auto updates we remove an ingress container
		// if it has a different image. in the next steps
		// the code will see a nil container and create a new one with
		// the specified version
		if err := t.Shutdown(); err != nil {
      state.Ingress.SetError(
        runtimeErr.FailedInteractionWithDockerDaemon(runtimeErr.DockerStopContainer, err),
      )
		}

		return nil
	})
	if optionalContainer.IsPresent() {
		return optionalContainer
	}

	container, ok := nginxcontroller.CreateContainer(state)
	if !ok {
		return optional.Empty[runtime.Container]()
	}

	return optional.Of(container)
}

func SelectIngressContainer(
	state *record.ApplicationRecord,
) optional.Optional[runtime.Container] {
	optionalContainer, err := runtime.TrySelectOneContainer(
		state.Metadata.Application,
		runtime.ObjectTypeLabel(runtime.IngressObject),
		runtime.ApplicationNameLabel(state.Metadata.Application),
	)
	if err != nil {
    state.Ingress.SetError(
      runtimeErr.FailedInteractionWithDockerDaemon(runtimeErr.DockerSelectContainer, err),
    )

		return optional.Empty[runtime.Container]()
	}

	return optionalContainer
}
