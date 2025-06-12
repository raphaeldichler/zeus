// Licensed under the Apache License 2.0. See the LICENSE file for details.
// Copyright 2025 The Zeus Authors.

package nginxcontroller

import (
	"context"
	"fmt"
	"math/rand/v2"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/raphaeldichler/zeus/internal/assert"
	"github.com/raphaeldichler/zeus/internal/record"
	"github.com/raphaeldichler/zeus/internal/runtime"
)

var image = ""

func id() string {
	id := rand.IntN(1000000)
	return fmt.Sprintf("%d", id)
}

func buildIngressContainer(application string) {
	id := rand.IntN(1000000)
	root, err := findProjectRoot()
	assert.ErrNil(err)
	imageRef := fmt.Sprintf("%s:%d", application, id)
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

	image = imageRef
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
	buildIngressContainer("nginx")
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

func TestNginxcontroller(t *testing.T) {
	application := "nginxcontroller-" + id()

	network, err := runtime.CreateNewNetwork(application)
	assert.ErrNil(err)
	assert.NotNil(network, "must create network")
	//defer network.Cleanup()

	state := record.ApplicationRecord{
		Ingress: record.NewIngressRecord(),
	}
	state.Ingress.Metadata.Image = image
	state.Ingress.Metadata.Name = application
	state.Ingress.Metadata.CreateTime = time.Now()

	_, ok := CreateContainer(&state)
	assert.True(ok, "container failed to create")
	//defer container.Shutdown()

	client, err := NewClient(application)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	resp, err := client.SetIngressConfig(ctx, &IngressRequest{
		Servers: []*Server{
			{
				Domain: "localhost",
				Locations: []*Location{
					newLocation(
						"/",
						Matching_Prefix,
						"return 200 'Foo'",
						"add_header Content-Type text/plain",
					),
				},
			},
		},
	})
	fmt.Println(resp, err)
}
