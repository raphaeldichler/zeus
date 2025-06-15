// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package main

import (
	"fmt"
	"os"

	"github.com/raphaeldichler/zeus/internal/zeusctl"
)

func main() {
  cmd := zeusctl.NewCommand()

  if err := cmd.Run(); err != nil {
    fmt.Fprintf(os.Stderr, "Failed to run command, got '%s'\n", err.Error())
    os.Exit(1)
  }
}
