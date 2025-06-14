// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package zeusapiserver

import (
	"net"
	"os"

	"github.com/raphaeldichler/zeus/internal/server"
)

const SocketPath = "/run/zeus/zeusd.sock"

type ZeusController struct {
	server *server.HttpServer

	records      *RecordCollection
	application  *ApplicationController
	orchestrator *orchestrator
}

func New() (*ZeusController, error) {
	if _, err := os.Stat(SocketPath); err == nil {
		if err := os.Remove(SocketPath); err != nil {
			return nil, err
		}
	}

	listen, err := net.Listen("unix", SocketPath)
	if err != nil {
		return nil, err
	}

	records, err := OpenAndCreateRecordCollection()
	if err != nil {
		return nil, err
	}
	applicationController := NewApplication(records)
	orchestrator := newOrchestrator(records)

	self := &ZeusController{
		application:  applicationController,
		orchestrator: orchestrator,
	}
	self.server = server.New(
		server.WithListener(listen),
		// applications
		server.Get(
			InspectAllApplicationAPIPath,
			applicationController.InspectAllApplication,
			server.WithRequestDecoder(applicationController.DecoderInspectAllApplicationRequest),
		),
		server.Get(
			InspectApplicationAPIPath,
			applicationController.InspectApplication,
			server.WithRequestDecoder(applicationController.DecoderInspectApplicationRequest),
		),
		server.Post(
			CreateApplicationAPIPath,
			applicationController.CreateApplication,
			server.WithRequestDecoder(applicationController.DecoderCreateApplicationRequest),
		),
		server.Delete(
			DeleteApplicationAPIPath,
			applicationController.DeleteApplication,
			server.WithRequestDecoder(applicationController.DecoderDeleteApplicationRequest),
		),
		server.Post(
			EnableApplicationAPIPath,
			applicationController.EnableApplication,
			server.WithRequestDecoder(applicationController.DecoderEnableApplicationRequest),
		),
		server.Post(
			DisableApplicationAPIPath,
			applicationController.DisableApplication,
			server.WithRequestDecoder(applicationController.DecoderDisableApplicationRequest),
		),
		// Ingress
		server.Get(
			IngressInspectAPIPath,
			self.GetIngressInspect,
			server.WithRequestDecoder(GetIngressInspectRequestDecoder),
		),
		server.Post(
			IngressApplyAPIPath,
			self.PostIngressApply,
			server.WithRequestDecoder(PostIngressApplyRequestDecoder),
		),
	)

	return self, nil
}

func (self *ZeusController) Run() error {
	defer func() {
		self.orchestrator.close()
		self.records.cleanup()
	}()

	return self.server.Run()
}
