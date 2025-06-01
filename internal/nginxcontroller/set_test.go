// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package nginxcontroller

import (
	"slices"
	"testing"
)

type TestEntry struct {
  Key string
  Other string
}

func (self TestEntry) Equal(other *TestEntry) bool  {
  return self.Key == other.Key
}


func TestNginxControllerSet(t *testing.T) {
  s := NewSet[TestEntry]()

  e1 := &TestEntry{Key: "k1", Other: "o1"}
  e2 := &TestEntry{Key: "k2", Other: "o1"}

  s.set(e1)
  s.set(e2)

  entries := s.entries()
  if !slices.Contains(entries, e1) {
    t.Errorf("set should container '%q', but doesnt.", e1)
  }
  if !slices.Contains(entries, e2) {
    t.Errorf("set should container '%q', but doesnt.", e2)
  }

  e3 := &TestEntry{Key: "k2", Other: "o3"}
  s.set(e3)
  entries = s.entries()
  if len(entries) != 2 {
    t.Errorf("set should only container 2 elemtns, but has '%d.", len(entries))
  }

  get1 := s.get(e1)
  get2 := s.get(e2)

  if e1.Key != get1.Key || e1.Other != get1.Other {
    t.Errorf("set should return element '%q', but got '%q'", e1, get1)
  }
  if e3.Key != get2.Key || e3.Other != get2.Other {
    t.Errorf("set should return element '%q', but got '%q'", e2, get2)
  }
}
