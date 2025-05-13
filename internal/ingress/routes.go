// Copyright 2025 The Zeus Authors.
// Licensed under the MIT License. See the LICENSE file for details.

package ingress

import (
	_ "embed"
)

// "/var/lib/zeus/poseidon/network/ingress/nginx/nginx.conf"
// "/var/lib/zeus/poseidon/network/ingress/nginx/conf.d/"

//go:embed nginx.conf
var NginxConfDefault string

