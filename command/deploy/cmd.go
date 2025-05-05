package deploy

import (
	"fmt"

	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

func NewDeployCommand(cli *client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy [file]",
		Short: "Deploy a service",
		Args:  cobra.ExactArgs(1),
    Run:  func(cmd *cobra.Command, args []string) {
        file := args[0]
        fmt.Printf("Deploying file: %s\n", file)

        config, err := loadConfig(file)
        if err != nil {
          panic(err)
        }

        if err := deployService(cli, config); err != nil {
          panic(err)
        }
    },
	}

  return cmd
}
