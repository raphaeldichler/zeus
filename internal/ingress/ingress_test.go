// Licensed under the Apache License 2.0. See the LICENSE file for details.
// Copyright 2025 The Zeus Authors.

package ingress

import (
	"fmt"
	"math/rand/v2"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func ID() string {
	id := rand.IntN(1000000)
	return fmt.Sprintf("%d", id)
}

func buildIngressContainer(application string) string {
	id := rand.IntN(1000000)
	root, err := findProjectRoot()
	image := fmt.Sprintf("%s:%d", application, id)
	ingressContainerPath := filepath.Join(root, "cmd", "nginxcontroller", "Dockerfile")

	fmt.Printf("[INFO] [%s] Building Docker image: %s\n", application, image)
	cmd := exec.Command("docker", "build", "--file", ingressContainerPath, "--tag", image, root)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] [%s] Docker build failed: %v\n", application, err)
		fmt.Fprintf(os.Stderr, "[DOCKER OUTPUT] [%s]\n%s\n", application, string(out))
		os.Exit(1)
	}
	fmt.Printf("[INFO] [%s] Docker image build succeeded\n", application)

	return image
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
	exitCode := m.Run()

	image := ""
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
