// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/raphaeldichler/zeus/internal/nginxcontroller"
)

func main() {
  nginxPath := flag.String("nginx-path", "", "Path to the nginx executable")
	flag.Parse()

  if *nginxPath == "" {
    flag.Usage()
		os.Exit(1)
  }
  
  ctr, err := nginxcontroller.NewServer(*nginxPath)
  if err != nil {
    fmt.Fprintf(os.Stderr, "")
		os.Exit(1)
  }

	fmt.Println("Nginx path:", *nginxPath)
  fmt.Println("Hello World")
}
