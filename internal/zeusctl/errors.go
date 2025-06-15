// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package zeusctl

import (
	"fmt"
	"os"

	"github.com/raphaeldichler/zeus/internal/assert"
	"github.com/spf13/cobra"
)

// Prints the error message, then prints the usage message for the command and exits with a non-zero exit code.
func failOnError(err error, format string, args ...any) {
	assert.StartsNotWithString(format, "Error: ", "Error: prefix should not be used")
	if err != nil {
		fmt.Fprintf(os.Stderr, format, args...)
		os.Exit(1)
	}
}

// Prints the error message, then prints the usage message for the command and exits with a non-zero exit code.
func failCommand(cmd *cobra.Command, format string, args ...any) {
	assert.StartsNotWithString(format, "Error: ", "Error: prefix should not be used")
	fmt.Printf("Error: "+format+"\n", args...)
	cmd.Usage()
	os.Exit(1)
}
