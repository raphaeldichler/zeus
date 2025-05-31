// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package ingress

import (
	"errors"
	"slices"
	"testing"
)

func TestIngressNginxCacheLocation(t *testing.T) {
	cache := NewNginxCache()

  sid := ServerIdentifier {
    Domain: "localhost",
    TlsEnabled: true,
    IPv6: true,
  }
  lid := LocationIdentifier{
    ServerIdentifier: sid,
    Path: "/",
    Matching: LocationExact,
  }
  cfg := NewLocationConfig(lid)

  for range 3 {
    if cache.isLocationCached(cfg) {
      t.Errorf("assert element cached failed. %q should be NOT be cached, but was", cfg)
    }

    cache.setLocation(cfg)
    if !cache.isLocationCached(cfg) {
      t.Errorf("assert element cached failed. %q should be be cached, but was not", cfg)
    }
    if !cache.isLocationIdentifierCached(lid) {
      t.Errorf("assert element cached failed. %q should be be cached, but was not", cfg)
    }

    cache.unsetLocation(lid)
    if cache.isLocationCached(cfg) {
      t.Errorf("assert element cached failed. %q should be NOT be cached, but was", cfg)
    }
    if !cache.isLocationIdentifierCached(lid) {
      t.Errorf("assert element cached failed. %q should be be cached, but was not", cfg)
    }
  }
}

func TestIngressNginxCacheServer(t *testing.T) {
	cache := NewNginxCache()

  sid := ServerIdentifier {
    Domain: "localhost",
    TlsEnabled: true,
    IPv6: true,
  }
  cfg := NewServerConfig(sid)

  for range 3 {
    if cache.isServerCached(cfg) {
      t.Errorf("assert element cached failed. %q should be NOT be cached, but was", cfg)
    }
    if cache.isServerIdentifierCached(sid) {
      t.Errorf("assert element cached failed. %q should be NOT be cached, but was", cfg)
    }

    cache.setServer(cfg)
    if !cache.isServerCached(cfg) {
      t.Errorf("assert element cached failed. %q should be be cached, but was not", cfg)
    }
    if !cache.isServerIdentifierCached(sid) {
      t.Errorf("assert element cached failed. %q should be be cached, but was not", cfg)
    }

    cache.unsetServer(sid)
    if cache.isServerCached(cfg) {
      t.Errorf("assert element cached failed. %q should be NOT be cached, but was", cfg)
    }
    if cache.isServerIdentifierCached(sid) {
      t.Errorf("assert element cached failed. %q should be NOT be cached, but was", cfg)
    }
  }
}


/*
func TestNginxCacheSetAndChangeEntryData(t *testing.T) {
	cache := NewNginxCache()

	path := "/some/path"
	content := []byte("this is some data")
	e := &TestCacheEntry{
		path:    path,
		content: content,
	}
	assertElementIsNotCached(t, cache, e)

	cache.set(e)
	assertElementIsCached(t, cache, e)

	tmp := content[0]
	content[0] = '*'
	assertElementIsNotCached(t, cache, e)

	content[0] = tmp
	assertElementIsCached(t, cache, e)
}



func TestNginxCacheTransactionSucceeds(t *testing.T) {
	c := NewNginxCache()

	path1 := "/some/path"
	content1 := []byte("this is some data")
	e1 := &TestCacheEntry{
		path:    path1,
		content: content1,
	}
	assertElementIsNotCached(t, c, e1)

	c.set(e1)
	assertElementIsCached(t, c, e1)

	path2 := "/some/other"
	content2 := []byte("this is other data")
	e2 := &TestCacheEntry{
		path:    path2,
		content: content2,
	}

	c.transaction(func(cache *NginxCache) error {
		assertElementIsCached(t, cache, e1)
		cache.set(e2)
		assertElementIsCached(t, cache, e2)
		assertElementIsNotCached(t, c, e2)

		return nil
	})

	assertElementIsCached(t, c, e1)
	assertElementIsCached(t, c, e2)
}

func TestNginxCacheTransactionFails(t *testing.T) {
	c := NewNginxCache()

	path1 := "/some/path"
	content1 := []byte("this is some data")
	e1 := &TestCacheEntry{
		path:    path1,
		content: content1,
	}
	assertElementIsNotCached(t, c, e1)

	c.set(e1)
	assertElementIsCached(t, c, e1)

	path2 := "/some/other"
	content2 := []byte("this is other data")
	e2 := &TestCacheEntry{
		path:    path2,
		content: content2,
	}
	assertElementIsNotCached(t, c, e2)

	c.transaction(func(cache *NginxCache) error {
		assertElementIsCached(t, cache, e1)
		cache.set(e2)
		assertElementIsCached(t, cache, e2)
		assertElementIsNotCached(t, c, e2)

		return errors.New("some error")
	})

	assertElementIsCached(t, c, e1)
	assertElementIsNotCached(t, c, e2)
}

func TestNginxCacheTransactionSucceedsOverwrite(t *testing.T) {
	c := NewNginxCache()

	path1 := "/some/path"
	content1 := []byte("this is some data")
	e1 := &TestCacheEntry{
		path:    path1,
		content: content1,
	}
	assertElementIsNotCached(t, c, e1)

	c.set(e1)
	assertElementIsCached(t, c, e1)

	content2 := []byte("this is other data")
	e2 := &TestCacheEntry{
		path:    path1,
		content: content2,
	}
	assertElementIsNotCached(t, c, e2)

	c.transaction(func(cache *NginxCache) error {
		assertElementIsCached(t, cache, e1)
		// overwrite e1 with e2
		cache.set(e2)
		assertElementIsCached(t, cache, e2)
		assertElementIsNotCached(t, c, e2)

		return nil
	})

	assertElementIsNotCached(t, c, e1)
	assertElementIsCached(t, c, e2)
}

func TestNginxCacheTransactionFailsOverwrite(t *testing.T) {
	c := NewNginxCache()

	path1 := "/some/path"
	content1 := []byte("this is some data")
	e1 := &TestCacheEntry{
		path:    path1,
		content: content1,
	}
	assertElementIsNotCached(t, c, e1)

	c.set(e1)
	assertElementIsCached(t, c, e1)

	content2 := []byte("this is other data")
	e2 := &TestCacheEntry{
		path:    path1,
		content: content2,
	}
	assertElementIsNotCached(t, c, e2)

	c.transaction(func(cache *NginxCache) error {
		assertElementIsCached(t, cache, e1)
		// overwrite e1 with e2
		cache.set(e2)
		assertElementIsCached(t, cache, e2)
		assertElementIsNotCached(t, c, e2)

		return errors.New("some error")
	})

	assertElementIsCached(t, c, e1)
	assertElementIsNotCached(t, c, e2)
}

func TestIngressNginxCacheKeys(t *testing.T) {
  c := NewNginxCache()

	path1 := "/1"
	content1 := []byte("this is some data")
	e1 := &TestCacheEntry{
		path:    path1,
		content: content1,
	}

	path2 := "/2"
	content2 := []byte("this is some data")
	e2 := &TestCacheEntry{
		path:    path2,
		content: content2,
	}

  c.set(e1)
  c.set(e2)

  keys := c.keys()
  if len(keys) != 2 {
      t.Errorf("failed to get all keys from cache. expected only two keys in '%q'", keys)
  }
  if !slices.Contains(keys, path1) {
    t.Errorf("failed to get all keys from cache. expected key '%q' to be part of '%q'", path1, keys)
  }
  if !slices.Contains(keys, path2) {
    t.Errorf("failed to get all keys from cache. expected key '%q' to be part of '%q'", path2, keys)
  }

  c.unset(path1)
  keys = c.keys()
  if len(keys) != 1 {
      t.Errorf("failed to get all keys from cache. expected only one keys in '%q'", keys)
  }
  if !slices.Contains(keys, path2) {
    t.Errorf("failed to get all keys from cache. expected key '%q' to be part of '%q'", path2, keys)
  }

  c.unset(path2)

  keys = c.keys()
  if len(keys) != 0 {
      t.Errorf("failed to get all keys from cache. expected only zero keys in '%q'", keys)
  }
} */
