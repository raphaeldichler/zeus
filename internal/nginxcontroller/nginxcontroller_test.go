// Licensed under the Apache License 2.0. See the LICENSE file for details.
// Copyright 2025 The Zeus Authors.

package nginxcontroller

import (
	"context"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

func buildIngressContainer(application string) string {
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

func runNginxcontroller(t *testing.T, state *record.ApplicationRecord) func() {
	t.Helper()

	state.Metadata.Application = state.Metadata.Application + "-" + id()
	t.Logf("Run nginxcontroller as application %s", state.Metadata.Application)

	network, err := runtime.CreateNewNetwork(state.Metadata.Application)
	assert.ErrNil(err)
	assert.NotNil(network, "must create network")

	state.Metadata.Deployment = record.Development
	state.Ingress.Metadata.Image = image
	state.Ingress.Metadata.CreateTime = time.Now()

	socketPath := HostSocketDirectory(state.Metadata.Application)
	err = os.MkdirAll(socketPath, 0777)
	assert.ErrNil(err)

	container, ok := CreateContainer(state)
	assert.True(ok, "container failed to create")

	return func() {
		container.Shutdown()
		network.Cleanup()
	}
}

func assertHTTPRequest(t *testing.T, url string, expectedContent string) {
	t.Helper()
	resp, err := http.Get(url)
	assert.ErrNil(err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.ErrNil(err)

	if string(body) != expectedContent {
		t.Errorf("Failed to set correct value to the location. got '%s', want '%s'. url: %s", string(body), expectedContent, url)
	}
}

func assertHTTPSGetRequest(
	t *testing.T,
	url string,
	expectedContent string,
	privkey string,
) {
	t.Helper()

	block, _ := pem.Decode([]byte(privkey))
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		t.Fatal("Invalid or unsupported private key format (only RSA PKCS1 supported)")
	}
	privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		t.Fatalf("Failed to parse private key: %v", err)
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	client := &http.Client{
		Transport: &http.Transport{TLSClientConfig: tlsConfig},
	}

	resp, err := client.Get(url)
	if err != nil {
		t.Fatalf("HTTPS request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	if !strings.Contains(string(body), expectedContent) {
		t.Errorf("Response does not contain expected content.\nExpected: %q\nActual: %q", expectedContent, string(body))
	}

	connState := resp.TLS
	if connState == nil || len(connState.PeerCertificates) == 0 {
		t.Fatal("No server certificate found")
	}
	serverCert := connState.PeerCertificates[0]

	serverPubKey, ok := serverCert.PublicKey.(*rsa.PublicKey)
	if !ok {
		t.Fatal("Server certificate does not use RSA public key")
	}
	if serverPubKey.N.Cmp(privKey.PublicKey.N) != 0 || serverPubKey.E != privKey.PublicKey.E {
		t.Error("Server certificate does not match the provided private key")
	}
}

func assertCertificate(t *testing.T, fullchain string, privkey string) {
	certBlock, _ := pem.Decode([]byte(fullchain))
	if certBlock == nil || certBlock.Type != "CERTIFICATE" {
		t.Error("failed to decode certificate PEM")
	}
	certParsed, err := x509.ParseCertificate(certBlock.Bytes)
	assert.ErrNil(err)

	keyBlock, _ := pem.Decode([]byte(privkey))
	if keyBlock == nil {
		t.Error("failed to decode private key PEM")
	}

	var privKey any
	if priv, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes); err == nil {
		privKey = priv
	} else if priv, err := x509.ParsePKCS8PrivateKey(keyBlock.Bytes); err == nil {
		privKey = priv
	} else if priv, err := x509.ParseECPrivateKey(keyBlock.Bytes); err == nil {
		privKey = priv
	} else {
		panic("failed to parse private key with known formats")
	}

	match := false
	switch pub := certParsed.PublicKey.(type) {
	case *rsa.PublicKey:
		if priv, ok := privKey.(*rsa.PrivateKey); ok {
			match = pub.N.Cmp(priv.N) == 0 && pub.E == priv.E
		}
	case *ecdsa.PublicKey:
		if priv, ok := privKey.(*ecdsa.PrivateKey); ok {
			match = pub.X.Cmp(priv.X) == 0 && pub.Y.Cmp(priv.Y) == 0
		}
	case ed25519.PublicKey:
		if priv, ok := privKey.(ed25519.PrivateKey); ok {
			match = pub.Equal(priv.Public())
		}
	default:
		panic("unsupported public key type")
	}

	if !match {

		t.Error("Certificate and private key do not match")
	}
}

func TestNginxcontrollerHTTPServer(t *testing.T) {
	state := &record.ApplicationRecord{}
	state.Metadata.Application = "setting"
	c := runNginxcontroller(t, state)
	defer c()

	client, err := NewClient(state.Metadata.Application)
	ctx := context.Background()
	_, err = client.SetIngressConfig(ctx, &IngressRequest{
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
	assert.ErrNil(err)

	assertHTTPRequest(t, "http://localhost", "Foo")
}

func TestNginxcontrollerHTTPServerPrefixMatching(t *testing.T) {
	state := &record.ApplicationRecord{}
	state.Metadata.Application = "prefix"
	c := runNginxcontroller(t, state)
	defer c()

	client, err := NewClient(state.Metadata.Application)
	ctx := context.Background()
	_, err = client.SetIngressConfig(ctx, &IngressRequest{
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
	assert.ErrNil(err)

	assertHTTPRequest(t, "http://localhost", "Foo")
	assertHTTPRequest(t, "http://localhost/sure-not-defined", "Foo")
	assertHTTPRequest(t, "http://localhost/another-not-defined", "Foo")
}

func TestNginxcontrollerHTTPServerExactMatching(t *testing.T) {
	state := &record.ApplicationRecord{}
	state.Metadata.Application = "exact"
	c := runNginxcontroller(t, state)
	defer c()

	client, err := NewClient(state.Metadata.Application)
	assert.ErrNil(err)
	ctx := context.Background()
	_, err = client.SetIngressConfig(ctx, &IngressRequest{
		Servers: []*Server{
			{
				Domain: "localhost",
				Locations: []*Location{
					newLocation(
						"/exact",
						Matching_Exact,
						"return 200 'exact'",
						"add_header Content-Type text/plain",
					),
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
	assert.ErrNil(err)

	assertHTTPRequest(t, "http://localhost", "Foo")
	assertHTTPRequest(t, "http://localhost/exact", "exact")
	assertHTTPRequest(t, "http://localhost/exacts", "Foo")
	assertHTTPRequest(t, "http://localhost/exact/other", "Foo")
}

func TestNginxcontrollerHTTPServerResetting(t *testing.T) {
	state := &record.ApplicationRecord{}
	state.Metadata.Application = "resetting"
	c := runNginxcontroller(t, state)
	defer c()

	client, err := NewClient(state.Metadata.Application)
	assert.ErrNil(err)
	ctx := context.Background()
	_, err = client.SetIngressConfig(ctx, &IngressRequest{
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
	assert.ErrNil(err)
	assertHTTPRequest(t, "http://localhost", "Foo")

	_, err = client.SetIngressConfig(ctx, &IngressRequest{
		Servers: []*Server{
			{
				Domain: "localhost",
				Locations: []*Location{
					newLocation(
						"/",
						Matching_Prefix,
						"return 200 'Bra'",
						"add_header Content-Type text/plain",
					),
				},
			},
		},
	})
	assert.ErrNil(err)
	assertHTTPRequest(t, "http://localhost", "Bra")
}

func TestNginxcontrollerHTTPServerMultipleLocation(t *testing.T) {
	state := &record.ApplicationRecord{}
	state.Metadata.Application = "locations"
	c := runNginxcontroller(t, state)
	defer c()

	client, err := NewClient(state.Metadata.Application)
	assert.ErrNil(err)
	ctx := context.Background()
	_, err = client.SetIngressConfig(ctx, &IngressRequest{
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
					newLocation(
						"/bra",
						Matching_Prefix,
						"return 200 'Bra'",
						"add_header Content-Type text/plain",
					),
				},
			},
		},
	})
	assert.ErrNil(err)
	assertHTTPRequest(t, "http://localhost", "Foo")
	assertHTTPRequest(t, "http://localhost/bra", "Bra")
}

func TestNginxcontrollerHTTPServerSubdomains(t *testing.T) {
	state := &record.ApplicationRecord{}
	state.Metadata.Application = "locations"
	c := runNginxcontroller(t, state)
	defer c()

	client, err := NewClient(state.Metadata.Application)
	assert.ErrNil(err)
	ctx := context.Background()
	_, err = client.SetIngressConfig(ctx, &IngressRequest{
		Servers: []*Server{
			{
				Domain: "localhost",
				Locations: []*Location{
					newLocation(
						"/",
						Matching_Prefix,
						"return 200 'localhost'",
						"add_header Content-Type text/plain",
					),
				},
			},
			{
				Domain: "app.localhost",
				Locations: []*Location{
					newLocation(
						"/",
						Matching_Prefix,
						"return 200 'app.localhost'",
						"add_header Content-Type text/plain",
					),
				},
			},
		},
	})
	assert.ErrNil(err)
	assertHTTPRequest(t, "http://localhost", "localhost")
	assertHTTPRequest(t, "http://app.localhost", "app.localhost")
}

func TestNginxcontrollerGenerateCertificate(t *testing.T) {
	state := &record.ApplicationRecord{}
	state.Metadata.Application = "locations"
	c := runNginxcontroller(t, state)
	defer c()

	client, err := NewClient(state.Metadata.Application)
	assert.ErrNil(err)
	ctx := context.Background()
	resp, err := client.GenerateCertificates(ctx, &GenerateCertificateRequest{
		Type:             GenerateCertificateType_SelfSigned,
		CertificateEmail: "testing@zeus.com",
		Domain:           "localhost",
	})
	assert.ErrNil(err)

	if resp.Fullchain == "" {
		t.Errorf("no fullchain was returned")
	}
	if resp.Privkey == "" {
		t.Errorf("no privkey was returned")
	}

	assertCertificate(t, resp.Fullchain, resp.Privkey)
}

func TestNginxcontrollerHTTPSServer(t *testing.T) {
	state := &record.ApplicationRecord{}
	state.Metadata.Application = "locations"
	c := runNginxcontroller(t, state)
	defer c()

	client, err := NewClient(state.Metadata.Application)
	assert.ErrNil(err)
	ctx := context.Background()
	resp, err := client.GenerateCertificates(ctx, &GenerateCertificateRequest{
		Type:             GenerateCertificateType_SelfSigned,
		CertificateEmail: "testing@zeus.com",
		Domain:           "localhost",
	})
	assert.ErrNil(err)

	if resp.Fullchain == "" {
		t.Errorf("no fullchain was returned")
	}
	if resp.Privkey == "" {
		t.Errorf("no privkey was returned")
	}

	_, err = client.SetIngressConfig(ctx, &IngressRequest{
		Servers: []*Server{
			{
				Domain: "localhost",
				Tls: &TLS{
					Privkey:   resp.Privkey,
					Fullchain: resp.Fullchain,
				},
				Locations: []*Location{
					newLocation(
						"/",
						Matching_Prefix,
						"return 200 'https-content'",
						"add_header Content-Type text/plain",
					),
				},
			},
		},
	})
	assert.ErrNil(err)

	assertCertificate(t, resp.Fullchain, resp.Privkey)
	assertHTTPSGetRequest(t, "https://localhost", "https-content", resp.Privkey)
}

func TestNginxcontrollerHTTPSAndHTTPServerOnSameDomain(t *testing.T) {
	state := &record.ApplicationRecord{}
	state.Metadata.Application = "locations"
	runNginxcontroller(t, state)

	client, err := NewClient(state.Metadata.Application)
	assert.ErrNil(err)
	ctx := context.Background()
	resp, err := client.GenerateCertificates(ctx, &GenerateCertificateRequest{
		Type:             GenerateCertificateType_SelfSigned,
		CertificateEmail: "testing@zeus.com",
		Domain:           "localhost",
	})
	assert.ErrNil(err)

	if resp.Fullchain == "" {
		t.Errorf("no fullchain was returned")
	}
	if resp.Privkey == "" {
		t.Errorf("no privkey was returned")
	}

	_, err = client.SetIngressConfig(ctx, &IngressRequest{
		Servers: []*Server{
			{
				Domain: "localhost",
				Tls: &TLS{
					Privkey:   resp.Privkey,
					Fullchain: resp.Fullchain,
				},
				Locations: []*Location{
					newLocation(
						"/",
						Matching_Prefix,
						"return 200 'https-content'",
						"add_header Content-Type text/plain",
					),
				},
			},
			{
				Domain: "localhost",
				Locations: []*Location{
					newLocation(
						"/",
						Matching_Prefix,
						"return 200 'http-content'",
						"add_header Content-Type text/plain",
					),
				},
			},
		},
	})
	assert.ErrNil(err)

	assertCertificate(t, resp.Fullchain, resp.Privkey)
	assertHTTPSGetRequest(t, "https://localhost", "https-content", resp.Privkey)
	assertHTTPRequest(t, "http://localhost", "http-content")
}
