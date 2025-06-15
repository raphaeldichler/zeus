// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package zeusctl

import (
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/cobra"
)

var (
	ingress = &cobra.Command{
		Use:   "ingress",
		Short: "Ingress management commands",
	}
	filePath string
)

func ingressCommands(rootCmd *cobra.Command, clientProvider *clientProvider) {
	applyIngress(clientProvider)
	inspectIngress(clientProvider)
	rootCmd.AddCommand(ingress)
}

func applyIngress(clientProvider *clientProvider) {
	applyCmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply ingress configuration",
		Run: func(cmd *cobra.Command, args []string) {
			if filePath == "" {
				failCommand(cmd, "--file/-f is required")
			}
			fmt.Printf("Applying ingress from file: %s\n", filePath)
			fmt.Println("Applying ingress", clientProvider.client.application)
		},
	}

	applyCmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to ingress file")
	applyCmd.MarkFlagRequired("file")

	ingress.AddCommand(applyCmd)
}

func inspectIngress(clientProvider *clientProvider) {
	inspectCmd := &cobra.Command{
		Use:   "inspect",
		Short: "Inspect ingress configuration",
		Run: func(cmd *cobra.Command, args []string) {
			client := clientProvider.client.http

			fmt.Println("Inspecting ingress", clientProvider.client.application)

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
