// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package nginxcontroller

import (
	"time"

	"github.com/raphaeldichler/zeus/internal/assert"
	"github.com/raphaeldichler/zeus/internal/ingress/errtype"
	"github.com/raphaeldichler/zeus/internal/record"
	"github.com/raphaeldichler/zeus/internal/runtime"
)

var nginxRunningCheckBackoff = time.Millisecond * 500

func CreateContainer(state *record.ApplicationRecord) (container *runtime.Container, ok bool) {
  network, err := runtime.TrySelectOneNetwork(
    state.Metadata.Application,
    runtime.ObjectTypeLabel(runtime.NetworkObject),
    runtime.ApplicationNameLabel(state.Metadata.Application),
  )
  if err != nil {
    state.Ingress.Errors.SetIngressError(
      errtype.FailedInteractionWithDockerDaemon(errtype.DockerCreateContainer, err),
    )
  }
  assert.NotNil(network, "Network must exists, is created on application start")

  container, err = runtime.NewContainer(
    state.Metadata.Application,
    runtime.WithImage(state.Ingress.Metadata.Image),
    runtime.WithPulling(),
    runtime.WithExposeTcpPort("80", "80"),
    runtime.WithExposeTcpPort("443", "443"),
    runtime.WithConnectedToNetwork(network),
    runtime.WithLabels(
      runtime.ObjectTypeLabel(runtime.IngressObject),
      runtime.ObjectImageLabel(state.Ingress.Metadata.Image),
      runtime.ApplicationNameLabel(state.Metadata.Application),
    ),
    runtime.WithMount("socket", SocketMountPath),
  )
  if err != nil {
    state.Ingress.Errors.SetIngressError(
      errtype.FailedInteractionWithDockerDaemon(errtype.DockerCreateContainer, err),
    )
    return nil, false
  }

  runs := 0
  for {
    exists, err := container.ExitsPath(NginxPidFilePath)
    if runs == 3 {
      state.Ingress.Errors.SetIngressError(
        errtype.FailedInteractionWithDockerDaemon(errtype.DockerInspectContainer, err),
      )
      return nil, false
    }
    if err != nil {
      runs += 1
    }
    if exists {
      break
    }
    time.Sleep(nginxRunningCheckBackoff)
  }

	return container, true
}
