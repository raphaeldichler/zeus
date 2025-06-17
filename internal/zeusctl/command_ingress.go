// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package zeusctl

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/raphaeldichler/zeus/internal/assert"
	"github.com/raphaeldichler/zeus/internal/zeusapiserver"
	"github.com/spf13/cobra"
)

var (
	ingress = &cobra.Command{
		Use:   "ingress",
		Short: "Ingress management commands",
	}
	filePath string
)

func ingressCommands(rootCmd *cobra.Command, clientProvider *contextProvider) {
	applyIngress(clientProvider)
	inspectIngress(clientProvider)
	rootCmd.AddCommand(ingress)
}

type IngressApplyRequest struct {
	Version  string `json:"version"`
	Metadata struct {
		Name string `json:"name"`
	} `json:"metadata"`
	zeusapiserver.IngressApplyRequestBody `json:"ingress"`
}

func applyIngress(clientProvider *contextProvider) {
	applyCmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply ingress configuration",
		Run: func(cmd *cobra.Command, args []string) {
			client := clientProvider.client
			assert.True(filePath != "", "file path must not be empty")
			content, err := os.ReadFile(filePath)
			failOnError(err, "Could not read file: %v", err)

			reader := io.NopCloser(bytes.NewReader(content))
			apply := toObject[IngressApplyRequest](reader)

			fmt.Println(client.ingressApply(apply))
		},
	}

	applyCmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to ingress file")
	applyCmd.MarkFlagRequired("file")

	ingress.AddCommand(applyCmd)
}

func inspectIngress(clientProvider *contextProvider) {
	inspectCmd := &cobra.Command{
		Use:   "inspect",
		Short: "Inspect ingress configuration",
		Run: func(cmd *cobra.Command, args []string) {
			_ = clientProvider.client
		},
	}

	ingress.AddCommand(inspectCmd)
}

func (c *client) ingressApply(req *IngressApplyRequest) string {
	return ""
}
