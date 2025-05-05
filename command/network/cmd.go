package network

import (
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

func NewNetworkCommand(cli *client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "network",
		Short: "Deploy a service",
		Args:  cobra.ExactArgs(0),
	}

  cmd.AddCommand(
    newInspectCommand(),
  )

  return cmd
}
