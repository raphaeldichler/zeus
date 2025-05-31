// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package ingress

import (
	"maps"
	"slices"

	"github.com/raphaeldichler/zeus/internal/assert"
)

// NginxCache provides a simple in-memory cache for file contents, keyed by file path.
type NginxCache struct {
  locationCache map[LocationIdentifier]LocationConfig
  serverCache map[ServerIdentifier]ServerConfig
}

// NewNginxCache initializes and returns a new, empty NginxCache.
func NewNginxCache() *NginxCache {
	return &NginxCache{
    locationCache: make(map[LocationIdentifier]LocationConfig),
    serverCache: make(map[ServerIdentifier]ServerConfig),
	}
}

func (self *NginxCache) isLocationCached(requestCfg *LocationConfig) bool {
  cfg, ok := self.locationCache[requestCfg.LocationIdentifier]
  if !ok {
    return false
  }

  return requestCfg.Equal(&cfg) 
}

func (self *NginxCache) isLocationIdentifierCached(locationID LocationIdentifier) bool {
  _, ok := self.locationCache[locationID]
  return ok
}

func (self *NginxCache) setLocation(requestCfg *LocationConfig) {
  self.locationCache[requestCfg.LocationIdentifier] = *requestCfg
}

func (self *NginxCache) unsetLocation(locationID LocationIdentifier) {
	delete(self.locationCache, locationID)
}

func (self *NginxCache) locationKeys() []LocationIdentifier {
  return slices.Collect(maps.Keys(self.locationCache))
}

func (self *NginxCache) isServerCached(requestCfg *ServerConfig) bool {
  cfg, ok := self.serverCache[requestCfg.ServerIdentifier]
  if !ok {
    return false
  }

  return requestCfg.Equal(&cfg) 
}

func (self *NginxCache) isServerIdentifierCached(serverID ServerIdentifier) bool {
  _, ok := self.serverCache[serverID]
  return ok
}

func (self *NginxCache) setServer(cfg *ServerConfig) {
  _, ok := self.serverCache[cfg.ServerIdentifier] 
  assert.False(ok, "overwriting is not allowed")
  self.serverCache[cfg.ServerIdentifier] = *cfg
}

func (self *NginxCache) unsetServer(locationID ServerIdentifier) {
	delete(self.serverCache, locationID)
}

func (self *NginxCache) serverKeys() []ServerIdentifier {
  return slices.Collect(maps.Keys(self.serverCache))
}

// transaction executes a set of cache operations atomically.
// A copy of the current cache state is passed to the transaction function.
// If the function returns an error, all changes are discarded.
// If the function returns nil, the changes are committed to the main cache.
func (self *NginxCache) transaction(tx func(*NginxCache) error) error {
	transactionCache := &NginxCache{
    locationCache: maps.Clone(self.locationCache),
    serverCache: maps.Clone(self.serverCache),
	}

	if err := tx(transactionCache); err != nil {
		return err
	}

  self.locationCache = transactionCache.locationCache
  self.serverCache = transactionCache.serverCache
	return nil
}
