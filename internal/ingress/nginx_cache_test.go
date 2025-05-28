// Copyright 2025 The Zeus Authors.
// Licensed under the MIT License. See the LICENSE file for details.

package ingress

import (
	"testing"
)

type TestCacheEntry struct {
	path    string
	content []byte
}

func (self *TestCacheEntry) FilePath() string {
	return self.path
}

func (self *TestCacheEntry) FileContent() []byte {
	return self.content
}

func assertElementIsCached(t *testing.T, cache *NginxCache, e CacheEntry) {
	if !cache.isCached(e) {
		t.Errorf("assert element cached failed. %q should be cached, but wasnt", e)
	}
}

func assertElementIsNotCached(t *testing.T, cache *NginxCache, e CacheEntry) {
	if cache.isCached(e) {
		t.Errorf("assert element cached failed. %q should be NOT be cached, but was", e)
	}
}

func TestSetAndUnsetEntry(t *testing.T) {
	cache := NewNginxCache()

	path := "/some/path"
	content := []byte("this is some data")
	e := &TestCacheEntry{
		path:    path,
		content: content,
	}
	assertElementIsNotCached(t, &cache, e)

	cache.set(e)
	assertElementIsCached(t, &cache, e)

	cache.unset(path)
	assertElementIsNotCached(t, &cache, e)
}

func TestSetEntryMultipleSameKeys(t *testing.T) {
	cache := NewNginxCache()

	path := "/some/path"
	content := []byte("this is some data")
	e := &TestCacheEntry{
		path:    path,
		content: content,
	}
	assertElementIsNotCached(t, &cache, e)

	cache.set(e)
	assertElementIsCached(t, &cache, e)
	e1 := &TestCacheEntry{
		path:    path,
		content: []byte("some other data"),
	}
	assertElementIsNotCached(t, &cache, e1)

	cache.set(e1)
	assertElementIsCached(t, &cache, e1)
	assertElementIsNotCached(t, &cache, e)
}

func TestSetAndChangeEntryData(t *testing.T) {
	cache := NewNginxCache()

	path := "/some/path"
	content := []byte("this is some data")
	e := &TestCacheEntry{
		path:    path,
		content: content,
	}
	assertElementIsNotCached(t, &cache, e)

	cache.set(e)
	assertElementIsCached(t, &cache, e)

	tmp := content[0]
	content[0] = '*'
	assertElementIsNotCached(t, &cache, e)

	content[0] = tmp
	assertElementIsCached(t, &cache, e)
}
