// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package main

import (
	"fmt"
	"os"

	"github.com/raphaeldichler/zeus/internal/zeusapiserver"
)


func main() {
	server, err := zeusapiserver.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize the server, got '%s'\n", err.Error())
		os.Exit(1)
	}

	if err := server.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed run server, got '%s'\n", err.Error())
		os.Exit(1)
	}
}
