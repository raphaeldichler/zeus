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
	Version                               string `json:"version" yaml:"version"`
	zeusapiserver.IngressApplyRequestBody `json:"ingress" yaml:"ingress"`
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

			apply := toObject[IngressApplyRequest](
				io.NopCloser(bytes.NewReader(content)),
			)

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
			client := clientProvider.client

		},
	}

	ingress.AddCommand(inspectCmd)
}

func (c *client) ingressApply(apply *IngressApplyRequest) string {
	r, err := http.NewRequest(
		"POST",
		unixURL(zeusapiserver.InspectApplicationAPIPath(c.application)),
		nil,
	)
	assert.ErrNil(err)

	resp, err := c.http.Do(r)
	failOnError(err, "Request failed: %v", err)

	switch resp.StatusCode {
	case http.StatusOK:
		return "Created"
	case http.StatusBadRequest:
		return toError(resp)
	default:
		assert.Unreachable("cover all cases of status code")
	}

	return ""
}
