// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package zeusctl

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

var (
	ingress = &cobra.Command{
		Use:   "ingress",
		Short: "Ingress management commands",
	}
	filePath string
)

type ClientProvider struct {
	// Client returns the client. But the client is not initialized until the run command is called by Cobra
	Client *Client
}

func IngressCommands(rootCmd *cobra.Command, clientProvider *ClientProvider) {
	applyIngress(clientProvider)
	inspectIngress(clientProvider)
	rootCmd.AddCommand(ingress)
}

func applyIngress(clientProvider *ClientProvider) {
	applyCmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply ingress configuration",
		Run: func(cmd *cobra.Command, args []string) {
			if filePath == "" {
				fmt.Println("Error: --file/-f is required")
				cmd.Usage()
				os.Exit(1)
			}
			fmt.Printf("Applying ingress from file: %s\n", filePath)
			fmt.Println("Applying ingress", clientProvider.Client.application)
		},
	}

	applyCmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to ingress file")
	applyCmd.MarkFlagRequired("file")

	ingress.AddCommand(applyCmd)
}

func inspectIngress(clientProvider *ClientProvider) {
	inspectCmd := &cobra.Command{
		Use:   "inspect",
		Short: "Inspect ingress configuration",
		Run: func(cmd *cobra.Command, args []string) {
			client := clientProvider.Client.http

			fmt.Println("Inspecting ingress", clientProvider.Client.application)

			req, _ := http.NewRequest("GET", "http://unix/hello", nil)
			resp, err := client.Do(req)
			if err != nil {
				fmt.Printf("[%s] Request failed: %v", "inspect", err)
				return
			}
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)
			fmt.Printf("[%s] Response: %s", "inspect", body)

		},
	}

	ingress.AddCommand(inspectCmd)
}
