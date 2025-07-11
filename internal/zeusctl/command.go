// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package zeusctl

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/raphaeldichler/zeus/internal/util/assert"
	"github.com/raphaeldichler/zeus/internal/zeusctl/formatter"
	"github.com/spf13/cobra"
)

const (
	defaultConfigPath         string = "~/.zeus/config.yml"
	enviornmentNameZeusConfig string = "ZEUS_CONFIG"
)

type CommandProvider func(rootCmd *cobra.Command, clientProvider *contextProvider)

var (
	configPath   string
	outputFormat string
)

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

type Command struct {
	root *cobra.Command
}

type choosableOption struct {
	options map[string]bool
}

func newChoosableOption(options ...string) *choosableOption {
	result := &choosableOption{
		options: make(map[string]bool),
	}
	for _, option := range options {
		result.options[option] = true
	}
	return result
}

func (o *choosableOption) verify(option string) (ok bool) {
	if !o.options[option] {
		return false
	}
	return true
}

func NewCommand() *Command {
	clientProvider := new(contextProvider)
	rootCmd := &cobra.Command{
		Use:   "zeus",
		Short: "Zeus CLI",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			outputFormats := newChoosableOption("json", "yaml", "pretty")
			if !outputFormats.verify(outputFormat) {
				failCommand(cmd, "Invalid output format: %v", outputFormat)
			}
			formatter, ok := formatter.StringToFormat[outputFormat]
			assert.True(ok, "formatter must exist")

			path := defaultConfigPath
			if zeusConfig := os.Getenv(enviornmentNameZeusConfig); zeusConfig != "" {
				path = zeusConfig
			}
			if configPath != "" {
				path = configPath
			}
			path, err := expandPath(path)
			failOnError(err, "Could not resolve config path: %v", err)

			config, err := loadConfig(path)
			failOnError(err, "Could not load config: %v", err)

			client, err := config.newClient(formatter)
			failOnError(err, "Could not create client: %v", err)

			clientProvider.client = client
		},
	}

	rootCmd.PersistentFlags().StringVar(
		&configPath,
		"config",
		"",
		fmt.Sprintf(
			"Config file path (default: \"%s\", can also be set via $%s)",
			defaultConfigPath, enviornmentNameZeusConfig,
		),
	)
	rootCmd.PersistentFlags().StringVarP(
		&outputFormat, "output", "o", "pretty", "Output format: json, yaml, or pretty",
	)

	for _, provider := range []CommandProvider{
		ingressCommands,
		applicationCommands,
	} {
		provider(rootCmd, clientProvider)
	}

	return &Command{
		root: rootCmd,
	}
}

func (c *Command) Run() error {
	return c.root.Execute()
}
