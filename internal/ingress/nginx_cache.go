// Copyright 2025 The Zeus Authors.
// Licensed under the MIT License. See the LICENSE file for details.

package ingress

import (
	"bytes"
	"strings"

	"github.com/raphaeldichler/zeus/internal/assert"
)

type CacheEntry interface {
	FilePath() string
	FileContent() []byte
}

type NginxCache struct {
	cache map[string][]byte
}

func NewNginxCache() NginxCache {
	return NginxCache{
		cache: make(map[string][]byte),
	}
}

func (self *NginxCache) isCached(enty CacheEntry) bool {
	i, ok := self.cache[enty.FilePath()]
	if !ok {
		return false
	}

	return bytes.Compare(enty.FileContent(), i) == 0
}

func (self *NginxCache) isKeyCached(path string) bool {
	_, ok := self.cache[path]
	return ok
}

func (self *NginxCache) set(entry CacheEntry) {
	content := bytes.Clone(entry.FileContent())
	assert.NotNil(content, "cannot cache nil")

	key := strings.Clone(entry.FilePath())

	self.cache[key] = content
}

func (self *NginxCache) unset(path string) {
	delete(self.cache, path)
}
