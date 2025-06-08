// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package zeusapiserver

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/raphaeldichler/zeus/internal/assert"
	"github.com/raphaeldichler/zeus/internal/record"
	"go.etcd.io/bbolt"
	bboltErr "go.etcd.io/bbolt/errors"
)

const ZeusDataStorePath = "/run/zeus/store"

var (
	ErrStopIteration      = errors.New("stop bbolt for-each iteration")
	ErrApplicationEnabled = errors.New("stop bbolt for-each iteration")
)

type application string

func (a application) valid() bool {
	app := string(a)

	if strings.TrimSpace(app) != app {
		return false
	}

	if len(app) == 0 {
		return false
	}

	if len(app) >= bbolt.MaxKeySize {
		return false
	}

	if !isLetter(app) {
		return false
	}

	return true
}

type RecordCollection struct {
	db *bbolt.DB
	mu sync.Mutex
}

func OpenAndCreateRecordCollection() (*RecordCollection, error) {
	db, err := bbolt.Open(ZeusDataStorePath, 0600, nil)
	if err != nil {
		return nil, err
	}

	return &RecordCollection{
		db: db,
	}, nil
}

func (self *RecordCollection) cleanup() error {
	return self.db.Close()
}

// Only returns an error if the application cannot be found.
func (self *RecordCollection) delete(app application) error {
	self.mu.Lock()
	defer self.mu.Unlock()

	return self.db.Update(func(tx *bbolt.Tx) error {
		return tx.DeleteBucket([]byte(app))
	})
}

// Only returns an error if the application already exists.
func (self *RecordCollection) add(app application, deploymentType record.DeploymentType) error {
	self.mu.Lock()
	defer self.mu.Unlock()

	err := self.db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucket([]byte(app))
		if err != nil {
			return bboltErr.ErrBucketExists
		}

		appRecord := record.New(string(app), deploymentType)
		blob := appRecord.ToGob()
		assert.True(len(blob) < bbolt.MaxValueSize, "blob must stay under 2GB")
		err = b.Put(RecordKey, blob)
		assert.ErrNil(err)

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// Only returns an error if the application does not exists.
func (self *RecordCollection) get(app application) (*record.ApplicationRecord, error) {
	self.mu.Lock()
	defer self.mu.Unlock()

	var appRecord *record.ApplicationRecord
	err := self.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(app))
		if b == nil {
			return bboltErr.ErrBucketNotFound
		}

		recordBytes := b.Get(RecordKey)
		assert.NotNil(b, "application must have a record entry")

		appRecord = record.FromGob(recordBytes)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return appRecord, nil
}

// Error nil if app is enabled. Error == ErrApplicationEnabled one applicaiton already enabled
// Error == ErrBucketNotFound no applicaiton with this name
func (self *RecordCollection) enableIfNonElse(app application) error {
	self.mu.Lock()
	defer self.mu.Unlock()

	err := self.db.Update(func(tx *bbolt.Tx) error {
		err := tx.ForEach(func(name []byte, b *bbolt.Bucket) error {
			fmt.Println("foreach ", string(name))
			recordBytes := b.Get(RecordKey)
			assert.NotNil(b, "application must have a record entry")

			appRecord := record.FromGob(recordBytes)
			if appRecord.Metadata.Enabled {
				fmt.Println("stop")
				return ErrStopIteration
			}

			return nil
		})
		if errors.Is(err, ErrStopIteration) {
			return ErrApplicationEnabled
		}

		b := tx.Bucket([]byte(app))
		if b == nil {
			return bboltErr.ErrBucketNotFound
		}

		recordBytes := b.Get(RecordKey)
		assert.NotNil(b, "application must have a record entry")
		appRecord := record.FromGob(recordBytes)
		appRecord.Metadata.Enabled = true

		blob := appRecord.ToGob()
		assert.True(len(blob) < bbolt.MaxValueSize, "blob must stay under 2GB")
		err = b.Put(RecordKey, blob)
		assert.ErrNil(err)

		return nil
	})

	switch {
	case errors.Is(err, ErrApplicationEnabled), errors.Is(err, bboltErr.ErrBucketNotFound):
		return err

	case errors.Is(err, nil):
		return nil

	default:
		assert.Unreachable("cover all cases")
	}

	return nil
}

func (self *RecordCollection) all() []*record.ApplicationRecord {
	self.mu.Lock()
	defer self.mu.Unlock()

	var records []*record.ApplicationRecord = nil
	err := self.db.View(func(tx *bbolt.Tx) error {
		return tx.ForEach(func(name []byte, b *bbolt.Bucket) error {
			recordBytes := b.Get(RecordKey)
			assert.NotNil(b, "application must have a record entry")

			appRecord := record.FromGob(recordBytes)
			records = append(records, appRecord)

			return nil
		})
	})
	assert.ErrNil(err)

	return records
}

// Runs a transaction which first reads the record than performance action on it and after that its stored again.
// Its ensured that druing this transaction no other thread can interact with the data.
//
// Only returns an ErrBucketNotFound error if the defined app doesnt exists or the error which f returns.
func (self *RecordCollection) tx(app application, f func(rec *record.ApplicationRecord) error) error {
	self.mu.Lock()
	defer self.mu.Unlock()

	var appRecord *record.ApplicationRecord
	err := self.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(app))
		if b == nil {
			return bboltErr.ErrBucketNotFound
		}

		recordBytes := b.Get(RecordKey)
		assert.NotNil(b, "application must have a record entry")

		appRecord = record.FromGob(recordBytes)
		return nil
	})
	if err != nil {
		return err
	}

	if err := f(appRecord); err != nil {
		return err
	}

	err = self.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(app))
		assert.NotNil(b, "the application must exist")

		blob := appRecord.ToGob()
		assert.True(len(blob) < bbolt.MaxValueSize, "blob must stay under 2GB")
		err := b.Put(RecordKey, blob)
		assert.ErrNil(err)

		return nil
	})
	assert.ErrNil(err)

	return nil
}
