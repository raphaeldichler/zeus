// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package errtype

import (
	"github.com/raphaeldichler/zeus/internal/assert"
	"github.com/raphaeldichler/zeus/internal/record"
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

type NginxControllerInteraction int

const (
	NginxSend NginxControllerInteraction = iota + 1
	NginxApply
)

var nginxControllerInteractionMapping map[NginxControllerInteraction]string = map[NginxControllerInteraction]string{
	NginxSend:  "send",
	NginxApply: "apply",
}

func FailedInteractionWithDockerDaemon(identifier DockerDaemonInteraction, err error) record.IngressErrorEntryRecord {
	assert.True(err != nil, "the error must exist")
	id, ok := dockerDaemomnInteractionMapping[identifier]
	assert.True(ok, "identifier must exists")

	return record.IngressErrorEntryRecord{
		Type:       "FailedInteractionWithDockerDaemon",
		Identifier: id,
		Message:    err.Error(),
	}
}

func FailedObtainCertificate(host string, err error) record.IngressErrorEntryRecord {
	assert.True(err != nil, "the error must exist")
	return failedObtainCertificate(host, err.Error())
}

func FailedObtainCertificateQuery(host string) record.IngressErrorEntryRecord {
	return failedObtainCertificate(host, "")
}

func failedObtainCertificate(host string, error string) record.IngressErrorEntryRecord {
	return record.IngressErrorEntryRecord{
		Type:       "FailedObtainCertificate",
		Identifier: host,
		Message:    error,
	}
}

func FailedInteractionWithNginxController(action NginxControllerInteraction, err error) record.IngressErrorEntryRecord {
	assert.True(err != nil, "the error must exist")
	id, ok := nginxControllerInteractionMapping[action]
	assert.True(ok, "identifier must exists")

	return record.IngressErrorEntryRecord{
		Type:       "FailedInteractionWithNginxController",
		Identifier: id,
		Message:    err.Error(),
	}
}
