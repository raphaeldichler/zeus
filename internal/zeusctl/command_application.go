// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package zeusctl

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

/*
zeus application create --name=hades --type=production
zeus application inspect
zeus application inspect poseidon
zeus application delete poseiodn
zeus application enable|disable poseiodn
*/

var (
	application = &cobra.Command{
		Use:   "application",
		Short: "Application management commands",
	}
	applicationName string = ""
	applicationType string = ""
)

func ApplicationCommands(rootCmd *cobra.Command, clientProvider *ClientProvider) {
	createApplication(clientProvider)
	inspectApplication(clientProvider)
	deleteApplication(clientProvider)
	enableApplication(clientProvider)
	disableApplication(clientProvider)
	rootCmd.AddCommand(application)
}

func createApplication(clientProvider *ClientProvider) {
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create application",
		Run: func(cmd *cobra.Command, args []string) {
			if applicationName == "" {
				fmt.Println("Error: --name/-n is required")
				cmd.Usage()
				os.Exit(1)
			}

			applicationType = strings.ToLower(applicationType)
			switch applicationType {
			case "production", "development":
				fmt.Println("Running in", applicationType, "mode")

			default:
				fmt.Fprintf(os.Stderr, "Error: --type/-t must be 'production' or 'development'\n")
				cmd.Usage()
				os.Exit(1)
			}
			if applicationType == "" {
				fmt.Println("Error: --type/-t is required")
				cmd.Usage()
				os.Exit(1)
			}

			fmt.Printf("Creating application: %s\n", applicationName)
			fmt.Println("Creating application", clientProvider.Client.application)
		},
	}

	createCmd.Flags().StringVarP(&applicationName, "name", "n", "", "Name of the application")
	createCmd.Flags().StringVarP(&applicationType, "type", "t", "development", "Environment type (production|development)")

	createCmd.MarkFlagRequired("name")

	application.AddCommand(createCmd)
}

func inspectApplication(clientProvider *ClientProvider) {
	inspectCmd := &cobra.Command{
		Use:   "inspect",
		Short: "Inspect application",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Creating application: %s\n", applicationName)
			fmt.Println("Creating application", clientProvider.Client.application)
		},
	}

	application.AddCommand(inspectCmd)
}

func deleteApplication(clientProvider *ClientProvider) {
	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete application",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Creating application: %s\n", applicationName)
			fmt.Println("Creating application", clientProvider.Client.application)
		},
	}

	application.AddCommand(deleteCmd)
}

func enableApplication(clientProvider *ClientProvider) {
	enableCmd := &cobra.Command{
		Use:   "enable",
		Short: "Enable application",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Creating application: %s\n", applicationName)
			fmt.Println("Creating application", clientProvider.Client.application)
		},
	}

	application.AddCommand(enableCmd)
}

func disableApplication(clientProvider *ClientProvider) {
	disableCmd := &cobra.Command{
		Use:   "disable",
		Short: "Disable application",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Creating application: %s\n", applicationName)
			fmt.Println("Creating application", clientProvider.Client.application)
		},
	}

	application.AddCommand(disableCmd)
}
