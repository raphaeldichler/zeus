package network

import (
	"fmt"

	"github.com/spf13/cobra"
)


func newInspectCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inspect",
		Short: "Display detailed information on network",
		Args:  cobra.ExactArgs(0),
    Run:  func(cmd *cobra.Command, args []string) {
        fmt.Printf("Inspect file\n")
    },
	}

  return cmd

}
