// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package zeusapiserver

import (
	"context"

	"github.com/raphaeldichler/zeus/internal/dnscontroller"
	"github.com/raphaeldichler/zeus/internal/ingress"
	"github.com/raphaeldichler/zeus/internal/record"
	"github.com/raphaeldichler/zeus/internal/runtime"
	"github.com/raphaeldichler/zeus/internal/util/assert"
	log "github.com/raphaeldichler/zeus/internal/util/logger"
)

type (
	// setup is the function which is used to setup the application state before syncronization
	setup func() error

	// service is the function which is used to syncronize the application state
	service func(record *record.ApplicationRecord)
)

// EnviromentManager is responsible for setting up the enviroment on the host system.
// All operations which are done via the enviroment manager are not application
// dependend, but rather service dependend.
type EnviromentManager interface {
	Setup() error
}

var (
	environmentManagers = []EnviromentManager{
		dnscontroller.SocketFileEnvironmentManager,
	}
	services []service = []service{
		ingress.Sync,
	}
	setups []setup = []setup{
		ingress.Setup,
	}
)

type orchestrator struct {
	records *RecordCollection
	signal  chan struct{}
	cancel  context.CancelFunc
	logger  *log.Logger
}

func newOrchestrator(
	records *RecordCollection,
	logger *log.Logger,
) *orchestrator {
	ctx, cancel := context.WithCancel(context.Background())

	o := &orchestrator{
		records: records,
		signal:  make(chan struct{}),
		cancel:  cancel,
		logger:  logger,
	}
	go o.worker(ctx)

	return o
}

func (o *orchestrator) close() {
	o.cancel()
}

func (o *orchestrator) ping() {
	select {
	case o.signal <- struct{}{}:
	default:
	}
}

func (o *orchestrator) worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-o.signal:
			o.orchestrate()
		}
	}
}

func (o *orchestrator) orchestrate() {
	o.logger.Info("Orchestration was invoked")

	var failed = false
	for _, environmentManager := range environmentManagers {
		if err := environmentManager.Setup(); err != nil {
			o.logger.Info("")
			failed = true
		}
	}
	if failed {
		o.logger.Info("")
		return
	}

	record := o.records.getEnabledApplication()
	if record == nil {
		o.logger.Info("Filter enabled applications: no record found")
		return
	}

	err := o.disableNonApplicationContainer(record.Metadata.Application)
	if err != nil {
		return
	}

	for _, setup := range setups {
		if err := setup(); err != nil {
			o.logger.Error("Failed to setup application: %v", err)
			return
		}
	}

	nw, err := runtime.TrySelectApplicationNetwork(record.Metadata.Application)
	if err != nil {
		o.logger.Error("Failed to select application network: %v", err)
		return
	}
	if nw == nil {
		o.logger.Info("Start orchestration: no network found. Create new network")
		nw, err := runtime.CreateNewNetwork(record.Metadata.Application)
		if err != nil {
			o.logger.Error("Failed to create new network: %v", err)
			return
		}
		assert.NotNil(nw, "network must not be nil")
	}

	for _, svc := range services {
		svc(record)
	}

	o.records.sync(record)
}

// Disables all containers and networks that are not part of the application
func (o *orchestrator) disableNonApplicationContainer(application string) error {
	o.logger.Info("Disable non application containers")
	containers, err := runtime.SelectAllNonApplicationContainers(application)
	if err != nil {
		return err
	}

	for _, cont := range containers {
		o.logger.Info("Disable container %s", cont)
		if err := cont.Shutdown(); err != nil {
			return err
		}
	}

	o.logger.Info("Disable non application networks %s", application)
	networks, err := runtime.SelectAllNonApplicationNetworks(application)
	if err != nil {
		return err
	}
	for _, nw := range networks {
		o.logger.Info("Disable network %s", nw)
		if err := nw.Cleanup(); err != nil {
			return err
		}
	}

	return nil
}
