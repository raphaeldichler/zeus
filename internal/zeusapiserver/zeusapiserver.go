// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package zeusapiserver

import (
	"fmt"
	"net"
	"os"

	"github.com/raphaeldichler/zeus/internal/server"
	log "github.com/raphaeldichler/zeus/internal/util/logger"
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
		fmt.Println("error", err)
		return nil, err
	}
	applicationController := NewApplication(records)
	orchestrator := newOrchestrator(records, log.New("zeusapiserver", "orchestrator"))

	self := &ZeusController{
		application:  applicationController,
		orchestrator: orchestrator,
		records:      records,
	}
	self.server = server.New(
		server.WithListener(listen),
		// applications
		server.Get(
			inspectAllApplicationAPIPath,
			applicationController.InspectAllApplication,
			server.WithRequestDecoder(applicationController.DecoderInspectAllApplicationRequest),
		),
		server.Get(
			inspectApplicationAPIPath,
			applicationController.InspectApplication,
			server.WithRequestDecoder(applicationController.DecoderInspectApplicationRequest),
		),
		server.Post(
			createApplicationAPIPath,
			applicationController.CreateApplication,
			server.WithRequestDecoder(applicationController.DecoderCreateApplicationRequest),
		),
		server.Delete(
			deleteApplicationAPIPath,
			applicationController.DeleteApplication,
			server.WithRequestDecoder(applicationController.DecoderDeleteApplicationRequest),
		),
		server.Post(
			enableApplicationAPIPath,
			applicationController.EnableApplication,
			server.WithRequestDecoder(applicationController.DecoderEnableApplicationRequest),
		),
		server.Post(
			disableApplicationAPIPath,
			applicationController.DisableApplication,
			server.WithRequestDecoder(applicationController.DecoderDisableApplicationRequest),
		),
		// Ingress
		server.Get(
			ingressInspectAPIPath,
			self.GetIngressInspect,
			server.WithRequestDecoder(GetIngressInspectRequestDecoder),
		),
		server.Post(
			ingressApplyAPIPath,
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
