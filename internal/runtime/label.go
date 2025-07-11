// Copyright 2025 The Zeus Authors.
// Licensed under the MIT License. See the LICENSE file for details.

package runtime

import "github.com/raphaeldichler/zeus/internal/util/assert"

type Label struct {
	key   string
	value string
}

type ObjectLabel int

const (
	IngressObject ObjectLabel = iota + 1
	NetworkObject
)

const (
	labelObjectType      = "zeus.object.type"
	labelObjectImage     = "zeus.object.image"
	labelApplicationName = "zeus.application.name"
)

var objectLabelMapping map[ObjectLabel]string = map[ObjectLabel]string{
	IngressObject: "ingress",
	NetworkObject: "network",
}

// zeus.object.type={object}
func ObjectTypeLabel(object ObjectLabel) Label {
	e, ok := objectLabelMapping[object]
	assert.True(ok, "object label must exists")

	return Label{key: labelObjectType, value: e}
}

// zeus.object.image={image}
func ObjectImageLabel(image string) Label {
	return Label{key: labelObjectImage, value: image}
}

// zeus.application.name={name}
func ApplicationNameLabel(name string) Label {
	return Label{key: labelApplicationName, value: name}
}
