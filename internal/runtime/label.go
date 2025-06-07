// Copyright 2025 The Zeus Authors.
// Licensed under the MIT License. See the LICENSE file for details.

package runtime

import "github.com/raphaeldichler/zeus/internal/assert"

type Label struct {
	value string
	key   string
}

func ObjectTypeLabel(object string) Label {
	assert.In(object, []string{"ingress", "network"}, "can only set valid object as labels")

	return Label{value: "zeus.object.type", key: object}
}

func ObjectImageLabel(image string) Label {
	return Label{value: "zeus.object.image", key: image}
}

func ApplicationNameLabel(name string) Label {
	return Label{value: "zeus.application.name", key: name}
}
