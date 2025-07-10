// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package record

type RecordService struct{

  Services []ServiceSpec
}

type ServiceSpec struct {
  ServiceName RecordKey
  Network *ServiceNetwork
  Container *ServiceContainer
}

type ServiceNetwork struct {
  // Domain name of the service
  Name string
  // port name to port number
  PortMapping map[string]string
}

type ServiceContainer struct {
	Image string
}

func (self *RecordService) GetEndpoint(service RecordKey) string {
	return ""
}
