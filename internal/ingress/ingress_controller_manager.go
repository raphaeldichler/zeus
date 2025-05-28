// Copyright 2025 The Zeus Authors.
// Licensed under the MIT License. See the LICENSE file for details.

package ingress

type IngressControllerManager struct {
	controllers map[string]NginxController
}

func (self *IngressControllerManager) Sync() error {

	return nil
}
