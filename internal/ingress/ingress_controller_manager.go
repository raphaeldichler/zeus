// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package ingress

type IngressControllerManager struct {
	controllers map[string]NginxController
}

func (self *IngressControllerManager) Sync() error {

	return nil
}
