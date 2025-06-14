// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package zeusapiserver

import (
	"context"

	"github.com/raphaeldichler/zeus/internal/ingress"
	"github.com/raphaeldichler/zeus/internal/record"
	"github.com/raphaeldichler/zeus/internal/runtime"
)

type service func(record *record.ApplicationRecord)

var services []service = []service{
	func(record *record.ApplicationRecord) {
		ingress.Sync(record)
	},
}

type orchestrator struct {
	records *RecordCollection
	signal  chan struct{}
	cancel  context.CancelFunc
}

func newOrchestrator(
	records *RecordCollection,
) *orchestrator {
	ctx, cancel := context.WithCancel(context.Background())

	o := &orchestrator{
		records: records,
		signal:  make(chan struct{}),
		cancel:  cancel,
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
	record := o.records.getEnabledApplication()
	err := disableNonApplicationContainer(record.Metadata.Application)
	if err != nil {
		return
	}

	for _, svc := range services {
		svc(record)
	}

	// recorver

}

// Disables all containers and networks that are not part of the application
func disableNonApplicationContainer(application string) error {
	containers, err := runtime.SelectAllNonApplicationContainers(application)
	if err != nil {
		return err
	}
	for _, cont := range containers {
		if err := cont.Shutdown(); err != nil {
			return err
		}
	}

	networks, err := runtime.SelectAllNonApplicationNetworks(application)
	if err != nil {
		return err
	}
	for _, nw := range networks {
		if err := nw.Cleanup(); err != nil {
			return err
		}
	}

	return nil
}
