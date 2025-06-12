// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package record

import (
	"bytes"
	"encoding/gob"

	"github.com/raphaeldichler/zeus/internal/assert"
	"github.com/raphaeldichler/zeus/internal/log"
)

type DeploymentType int

const (
	Production DeploymentType = iota + 1
	Development
)

func (d DeploymentType) String() string {
	switch d {
	case Production:
		return "production"

	case Development:
		return "development"

	default:
		assert.Unreachable("all cases are covered")
	}

	return ""
}

type RecordKey string

type ApplicationRecord struct {
	Metadata ApplicationMetadata
	Ingress  RecordIngress
	Service  RecordService
}

type ApplicationMetadata struct {
	Application string
	Deployment  DeploymentType
	Enabled     bool
}

func New(app string, deploymentType DeploymentType) *ApplicationRecord {
	return &ApplicationRecord{
		Metadata: ApplicationMetadata{
			Application: app,
			Deployment:  deploymentType,
			Enabled:     false,
		},
	}
}

func (self *ApplicationRecord) Logger(daemon string) *log.Logger {
	return log.New(self.Metadata.Application, daemon)
}

func (self *ApplicationRecord) ToGob() []byte {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(self)
	assert.ErrNil(err)

	return buf.Bytes()
}

func FromGob(data []byte) *ApplicationRecord {
	out := new(ApplicationRecord)
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(out)
	assert.ErrNil(err)

	return out
}

// Persists the application state
func (self *ApplicationRecord) Apply() error {
	return nil
}
