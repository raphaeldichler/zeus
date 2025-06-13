// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package ingress

import (
	"fmt"
	"math/rand/v2"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/raphaeldichler/zeus/internal/assert"
	"github.com/raphaeldichler/zeus/internal/nginxcontroller"
	"github.com/raphaeldichler/zeus/internal/record"
	"github.com/raphaeldichler/zeus/internal/runtime"
)

var image = ""

func id() string {
	id := rand.IntN(1000000)
	return fmt.Sprintf("%d", id)
}

func buildIngressContainer(application string) string {
	root, err := findProjectRoot()
	assert.ErrNil(err)
	imageRef := application + "-" + id()
	ingressContainerPath := filepath.Join(root, "cmd", "nginxcontroller", "Dockerfile")

	fmt.Printf("[INFO] [%s] Building Docker image: %s\n", application, imageRef)
	cmd := exec.Command("docker", "build", "--file", ingressContainerPath, "--tag", imageRef, root)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] [%s] Docker build failed: %v\n", application, err)
		fmt.Fprintf(os.Stderr, "[DOCKER OUTPUT] [%s]\n%s\n", application, string(out))
		os.Exit(1)
	}
	fmt.Printf("[INFO] [%s] Docker image build succeeded\n", application)

	return imageRef
}

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find project root (go.mod not found)")
		}
		dir = parent
	}
}

func TestMain(m *testing.M) {
	image = buildIngressContainer("nginx")
	exitCode := m.Run()

	if os.Getenv("TEST_DOCKER_CLEANUP") == "1" {
		fmt.Printf("[INFO] Cleaning up Docker image: %s\n", image)
		cmd := exec.Command("docker", "image", "rm", "--force", image)
		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] Failed to remove Docker image: %v\n", err)
			fmt.Fprintf(os.Stderr, "[DOCKER OUTPUT]\n%s\n", string(out))
			exitCode = 1
		} else {
			fmt.Println("[INFO] Docker image cleanup succeeded")
		}
	}

	os.Exit(exitCode)
}

func TestIngressSync(t *testing.T) {
	state := &record.ApplicationRecord{}

	state.Metadata.Application = "sync" + "-" + id()
	state.Metadata.Deployment = record.Development
	state.Ingress.Metadata.Image = image
	state.Ingress.Metadata.CreateTime = time.Now()

	state.Ingress.Servers = []*record.ServerRecord{
		{
			Host: "localhost",
			IPv6: false,
			Tls:  nil,
			HTTP: record.HttpRecord{
				Paths: []record.PathRecord{
					{
						Path:     "/",
						Matching: "prefix",
						Service:  "",
					},
				},
			},
		},
	}

	network, err := runtime.CreateNewNetwork(state.Metadata.Application)
	assert.ErrNil(err)
	assert.NotNil(network, "must create network")

	socketPath := nginxcontroller.HostSocketDirectory(state.Metadata.Application)
	err = os.MkdirAll(socketPath, 0777)
	assert.ErrNil(err)

	Sync(state)
}
