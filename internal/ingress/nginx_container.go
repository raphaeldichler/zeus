// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package ingress

import (
	"github.com/raphaeldichler/zeus/internal/runtime"
)

func DefaultContainer(
	application string,
	network *runtime.Network,
) (*runtime.Container, error) {
	// big questions: should i create my own docker image and run it from there?
	c, err := runtime.NewContainer(
		application,
		"ingress-nginx",
		runtime.WithImage("raphaeldichler/zeus-nginx:1.0.0"),
		runtime.WithPulling(),
		runtime.WithExposeTcpPort("80", "80"),
		runtime.WithExposeTcpPort("443", "443"),
		runtime.WithConnectedToNetwork(network),
		runtime.WithObjectTypeLabel("ingress"),
		//runtime.WithCopyIntoBeforeStart(DefaultHttpConfig()),
	)
	if err != nil {
		return nil, err
	}

	return c, nil
}
