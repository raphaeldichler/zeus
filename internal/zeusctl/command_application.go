// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package zeusctl

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/raphaeldichler/zeus/internal/util/assert"
	"github.com/raphaeldichler/zeus/internal/zeusapiserver"
	"github.com/spf13/cobra"
)

/*
zeus config set application=hades
zeus config set localhost.enabled=true

zeus application create --name=hades --type=production
zeus application inspect
--all/-a (default )
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

func applicationCommands(rootCmd *cobra.Command, clientProvider *contextProvider) {
	createApplication(clientProvider)
	inspectApplication(clientProvider)
	deleteApplication(clientProvider)
	enableApplication(clientProvider)
	disableApplication(clientProvider)
	rootCmd.AddCommand(application)
}

func createApplication(clientProvider *contextProvider) {
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create application",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			applicationName = args[0]

			applicationType = strings.ToLower(applicationType)
			if applicationType == "" {
				failCommand(cmd, "--type/-t is required")
			}

			applicationTypeOption := newChoosableOption("production", "development")
			if !applicationTypeOption.verify(applicationType) {
				failCommand(cmd, "--type/-t must be 'production' or 'development'")
			}

			client := clientProvider.client.http
			assert.NotNil(client, "client must not be nil")

			req, err := http.NewRequest(
				"POST",
				unixURL(zeusapiserver.CreateApplicationAPIPath()),
				zeusapiserver.NewCreateApplicationRequestAsJsonBody(
					applicationName,
					applicationType,
				),
			)
			assert.ErrNil(err)
			resp, err := client.Do(req)
			failOnError(err, "Request failed: %v", err)
			defer resp.Body.Close()

			switch resp.StatusCode {
			case http.StatusCreated:
				fmt.Printf("Successfully created application: %s\n", applicationName)
				return
			case http.StatusBadRequest:
				fmt.Printf("Failed to create application: %s\n%s", applicationName, FormatJSON(resp.Body))
				return
			default:
				assert.Unreachable("cover all cases of status code")
			}
		},
	}

	createCmd.Flags().StringVarP(
		&applicationType, "type", "t", "development", "Application type: development or production",
	)

	application.AddCommand(createCmd)
}

func inspectApplication(clientProvider *contextProvider) {
	inspectCmd := &cobra.Command{
		Use:   "inspect [applications]",
		Short: "Inspect application",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client := clientProvider.client
			assert.NotNil(client, "client must not be nil")

			switch len(args) {
			case 0:
				fmt.Println(
					client.applicationInspectAll(),
				)
			case 1:
				fmt.Println(
					client.applicationInspect(args[0]),
				)
			default:
				assert.Unreachable("cover all cases of number of arguments")
			}
		},
	}

	application.AddCommand(inspectCmd)
}

func deleteApplication(clientProvider *contextProvider) {
	deleteCmd := &cobra.Command{
		Use:   "delete [applications]",
		Short: "Delete application",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			applicationName = args[0]
			assert.NotEmptyString(applicationName, "application name must not be empty")

			client := clientProvider.client
			assert.NotNil(client, "client must not be nil")

			fmt.Println(
				client.applicationDeleted(applicationName),
			)
		},
	}

	application.AddCommand(deleteCmd)
}

func enableApplication(clientProvider *contextProvider) {
	enableCmd := &cobra.Command{
		Use:   "enable [applications]",
		Short: "Enable application",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			applicationName = args[0]
			assert.NotEmptyString(applicationName, "application name must not be empty")

			client := clientProvider.client
			assert.NotNil(client, "client must not be nil")

			fmt.Println(
				client.applicationEnabled(applicationName),
			)
		},
	}

	application.AddCommand(enableCmd)
}

func disableApplication(clientProvider *contextProvider) {
	disableCmd := &cobra.Command{
		Use:   "disable [applications]",
		Short: "Disable application",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			applicationName = args[0]
			assert.NotEmptyString(applicationName, "application name must not be empty")

			client := clientProvider.client
			assert.NotNil(client, "client must not be nil")

			fmt.Println(
				client.applicationDisabled(applicationName),
			)
		},
	}

	application.AddCommand(disableCmd)
}
