// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package errtype

import (
	"github.com/raphaeldichler/zeus/internal/record"
	"github.com/raphaeldichler/zeus/internal/util/assert"
)

type DockerDaemonInteraction int

const (
	DockerCreateContainer DockerDaemonInteraction = iota + 1
	DockerSelectContainer
	DockerStopContainer
	DockerInspectContainer
	DockerCreateNetwork
)

var dockerDaemomnInteractionMapping map[DockerDaemonInteraction]string = map[DockerDaemonInteraction]string{
	DockerCreateContainer:  "create",
	DockerSelectContainer:  "select",
	DockerStopContainer:    "stop",
	DockerInspectContainer: "inspect",
	DockerCreateNetwork:    "network",
}

func FailedInteractionWithDockerDaemon(
	identifier DockerDaemonInteraction,
	err error,
) record.IngressErrorEntryRecord {
	assert.True(err != nil, "the error must exist")
	id, ok := dockerDaemomnInteractionMapping[identifier]
	assert.True(ok, "identifier must exists")

	return record.IngressErrorEntryRecord{
		Type:       "FailedInteractionWithDockerDaemon",
		Identifier: id,
		Message:    err.Error(),
	}
}
