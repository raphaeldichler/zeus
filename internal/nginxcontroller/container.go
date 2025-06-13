// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package nginxcontroller

import (
	"os"
	"path/filepath"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/go-connections/nat"
	"github.com/raphaeldichler/zeus/internal/assert"
	"github.com/raphaeldichler/zeus/internal/ingress/errtype"
	"github.com/raphaeldichler/zeus/internal/record"
	"github.com/raphaeldichler/zeus/internal/runtime"
)

var (
	nginxRunningCheckBackoff = time.Millisecond * 500
	hostSocketRoot           = "/run"
)

func init() {
	if root := os.Getenv("ZEUS_HOST_ROOT"); root != "" {
		hostSocketRoot = root
	}
}

// Returns the directory which is be used to store the socket for IPC
// between the container and the application.
func HostSocketDirectory(application string) string {
	return filepath.Join(hostSocketRoot, "zeus", application, "ingress")
}

// Creates an Ingress container which will be conntected to the network.
//
// If an error happends the error is written into the state and it returns nil, false. If
// the container creation succeeds a it returns a container, true.
func CreateContainer(state *record.ApplicationRecord) (container *runtime.Container, ok bool) {
	network, err := runtime.TrySelectApplicationNetwork(
		state.Metadata.Application,
	)
	if err != nil {
		state.Ingress.Errors.SetIngressError(
			errtype.FailedInteractionWithDockerDaemon(errtype.DockerCreateContainer, err),
		)
	}

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
		runtime.WithMount(HostSocketDirectory(state.Metadata.Application), SocketMountPath),
	)
	if err != nil {
		state.Ingress.Errors.SetIngressError(
			errtype.FailedInteractionWithDockerDaemon(errtype.DockerCreateContainer, err),
		)
		return nil, false
	}

	for runs := 0; ; {
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

func ValidateContainer(c *runtime.Container, state *record.ApplicationRecord) bool {
	assert.NotNil(c, "at this state the container was set correctly")

	inspect, err := c.Inspect()
	if err != nil {
		state.Ingress.Errors.SetIngressError(
			errtype.FailedInteractionWithDockerDaemon(errtype.DockerInspectContainer, err),
		)
		return false
	}

	mounts := inspect.Mounts
	if len(mounts) != 1 {
		return false
	}
	socketMount := mounts[0]
	correctSocketMount := container.MountPoint{
		Type:        mount.TypeBind,
		Source:      HostSocketDirectory(state.Metadata.Application),
		Destination: SocketMountPath,
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
