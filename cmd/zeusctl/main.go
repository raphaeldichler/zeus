package main

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/raphaeldichler/zeus/internal/zeusctl"
	"github.com/spf13/cobra"
)

const defaultConfigPath = "~/.zeus/config.yml"

type CommandProvider func(rootCmd *cobra.Command, clientProvider *zeusctl.ClientProvider)

var topLevelCmd = []CommandProvider{
	zeusctl.IngressCommands,
	zeusctl.ApplicationCommands,
}

func expandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		usr, err := user.Current()
		if err != nil {
			return "", err
		}
		path = filepath.Join(usr.HomeDir, path[1:])
	}
	return filepath.Abs(path)
}

func main() {
	var configPath string
	clientProvider := new(zeusctl.ClientProvider)

	rootCmd := &cobra.Command{
		Use:   "zeus",
		Short: "Zeus CLI",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			path := defaultConfigPath
			if zeusConfig := os.Getenv("ZEUS_CONFIG"); zeusConfig != "" {
				path = zeusConfig
			}
			if configPath != "" {
				path = configPath
			}
			path, err := expandPath(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: Could not resolve config path: %v\n", err)
				os.Exit(1)
			}

			config, err := zeusctl.LoadConfig(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: Could not load config: %v\n", err)
				os.Exit(1)
			}

			client, err := config.NewClient()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: Could not create client: %v\n", err)
				os.Exit(1)
			}

			clientProvider.Client = client
		},
	}
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "Path to optional config file")

	for _, provider := range topLevelCmd {
		provider(rootCmd, clientProvider)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
