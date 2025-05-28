// Copyright 2025 The Zeus Authors.
// Licensed under the MIT License. See the LICENSE file for details.

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
		runtime.WithImage("nginx:1.27"),
		runtime.WithPulling(),
		runtime.WithExposeTcpPort("80", "80"),
		runtime.WithExposeTcpPort("443", "443"),
		runtime.WithConnectedToNetwork(network),
		runtime.WithObjectTypeLabel("ingress"),
		runtime.WithCmd("nginx", "-g", "daemon off;", "-c", "/etc/nginx/nginx.conf"),
		runtime.WithCopyIntoBeforeStart(DefaultHttpConfig()),
	)
	if err != nil {
		return nil, err
	}

	for _, path := range []string{
		NginxInternalCertificatePath,
		NginxInternalServerPath,
		NginxInternalLocationPath,
	} {
		if err := c.EnsurePathExists(path); err != nil {
			return nil, err
		}
	}

	return c, nil
}
