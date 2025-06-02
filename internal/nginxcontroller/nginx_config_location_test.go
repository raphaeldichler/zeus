// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package nginxcontroller

import "testing"


func TestNginxControllerLocationConfigEqualDifferentMatching(t *testing.T) {
  locExact := NewLocationConfig(
    "/one",
    ExactMatching,
  )
  locPrefix := NewLocationConfig(
    "/one",
    PrefixMatching,
  )

  if locExact.Equal(locPrefix) {
    t.Errorf("exact location matches prefix location, but it should not.")
  }
}


func TestNginxControllerLocationConfigEqual(t *testing.T) {
  matchings := []struct {
    name string 
    matching MatchingType
  } {
    {
      name: "exact",
      matching: ExactMatching,
    },
    {
      name: "prefix",
      matching: PrefixMatching,

    },
  }

  for _, tt := range matchings {
    t.Run(tt.name, func(t *testing.T) {
      loc1 := NewLocationConfig(
        "/one",
        tt.matching,
      )
      loc2 := NewLocationConfig(
        "/one",
        tt.matching,
      )

      if !loc1.Equal(loc2) {
        t.Errorf("locations dont match, but should.")
      }
    })
  }
}
