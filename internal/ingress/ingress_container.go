// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package ingress

import (
	"os"
	"time"
	"path/filepath"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/go-connections/nat"
	"github.com/raphaeldichler/zeus/internal/assert"
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
func SelectOrCreateIngressContainer(
	state *record.ApplicationRecord,
) (container *runtime.Container) {
  defer func() {
    if container == nil {
      return
    }

    if ok := validateContainer(container, state.Metadata.Application, state); ok {
      return
    }

    if err := container.Shutdown(); err != nil {
      state.Ingress.Errors.SetIngressError(
        errtype.FailedInteractionWithDockerDaemon(errtype.DockerStopContainer, err),
      )
    }
    container = nil
  }()

  // todo: change to select one container
  selectedContainers, err := runtime.SelectContainer(
    runtime.ObjectTypeLabel(runtime.IngressObject),
    runtime.ObjectImageLabel(state.Ingress.Metadata.Image),
    runtime.ApplicationNameLabel(state.Metadata.Application),
  )
  if err != nil {
    state.Ingress.Errors.SetIngressError(
      errtype.FailedInteractionWithDockerDaemon(errtype.DockerSelectContainer, err),
    )
    return nil
  }

  switch len(selectedContainers) {
  case 0:
    break

  case 1:
    c, err := selectedContainers[0].NewContainer(state.Metadata.Application)
    if err != nil {
      state.Ingress.Errors.SetIngressError(
        errtype.FailedInteractionWithDockerDaemon(errtype.DockerCreateContainer, err),
      )
      return nil
    }
    return c

  default:
    assert.Unreachable(
      "Too many ingress container exists in the current context. Possible external tampering.",
    )
  }

  networks, err := runtime.SelectNetworks(
    runtime.ObjectTypeLabel(runtime.NetworkObject),
    runtime.ApplicationNameLabel(state.Metadata.Application),
  )
  if err != nil {
    state.Ingress.Errors.SetIngressError(
      errtype.FailedInteractionWithDockerDaemon(errtype.DockerCreateContainer, err),
    )
  }

  var n *runtime.Network = nil
  switch len(networks) {
  case 1:
    n = networks[0].NewNetwork(state.Metadata.Application)

  default:
    assert.Unreachable("Network must exists, is created on application start")
  }
  assert.NotNil(n, "network must be selected")

  socket := filepath.Join("/run/zeus", state.Metadata.Application, "ingress")
  err = os.MkdirAll(socket, 0777)


  container, ok := nginxcontroller.CreateContainer(state)
  if !ok {
    return nil
  }

	return container
}

func validateContainer(c *runtime.Container, application string, state *record.ApplicationRecord) bool {
	assert.NotNil(c, "at this state the container was set correctly")

	inspect, err := c.Inspect()
	if err != nil {
		state.Ingress.Errors.SetIngressError(
			errtype.FailedInteractionWithDockerDaemon(errtype.DockerInspectContainer, err),
		)
		return false
	}

  socket := filepath.Join("/run/zeus", state.Metadata.Application, "ingress")
	mounts := inspect.Mounts
	if len(mounts) != 1 {
		return false
	}
	socketMount := mounts[0]
	correctSocketMount := container.MountPoint{
		Type:        mount.TypeBind,
		Source:      socket,
		Destination: nginxcontroller.SocketMountPath,
		Mode:        "",
		RW:          true,
		Propagation: "rprivate",
	}
	if socketMount != correctSocketMount {
		return false
	}

	portBindings := inspect.HostConfig.PortBindings
	if portBindings == nil {
		return false
	}
	tcp443, ok := portBindings[nat.Port("443/tcp")]
	if !ok {
		return false
	}
	correctTcp443 := nat.PortBinding{HostPort: "443", HostIP: ""}
	if len(tcp443) != 1 || tcp443[0] != correctTcp443 {
		return false
	}
	tcp80, ok := portBindings[nat.Port("80/tcp")]
	if !ok {
		return false
	}
	correctTcp80 := nat.PortBinding{HostPort: "80", HostIP: ""}
	if len(tcp80) != 1 || correctTcp80 != tcp80[0] {
		return false
	}

	return true
}
