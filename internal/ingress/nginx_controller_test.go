// Copyright 2025 The Zeus Authors.
// Licensed under the MIT License. See the LICENSE file for details.

package ingress

import (
	"fmt"
	"io"
	"net/http"
	"testing"

	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"

	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"time"

	"github.com/raphaeldichler/zeus/internal/assert"
	"github.com/raphaeldichler/zeus/internal/runtime"
)

var (
	client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
)

func generateSelfSignedCert(domain string) (certPEM []byte, keyPEM []byte) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	assert.ErrNil(err)

	serialNumber, err := rand.Int(rand.Reader, big.NewInt(1<<62))
	assert.ErrNil(err)

	template := x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               pkix.Name{CommonName: "localhost"},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{domain},
	}
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	assert.ErrNil(err)
	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

	privBytes, err := x509.MarshalECPrivateKey(priv)
	assert.ErrNil(err)
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes})

	return certPEM, keyPEM
}

func setupTestingNginxController(t *testing.T) (controller *NginxController, cleanup func(), err error) {
	application := "testing"

	nw, err := runtime.NewNetwork(application, "network")
	if err != nil {
		t.Errorf("failed to setup network. got %q", err)
		return nil, nil, err
	}

	container, err := DefaultContainer(application, nw)
	if err != nil {
		t.Errorf("failed to setup conainer. got %q", err)
		if err := nw.Cleanup(); err != nil {
			t.Errorf("failed to cleanup network. got %q", err)
		}
		return nil, nil, err
	}
	container.AssertIsRunning()

	return NewNginxController(container, application),
		func() {
			if err := container.Shutdown(); err != nil {
				t.Errorf("failed to shutdown container, got %q", err)
			}
			if err := nw.Cleanup(); err != nil {
				t.Errorf("failed to cleanup network, got %q", err)
			}
		},
		nil
}

func assertHttpOrHttpsOk(
	t *testing.T,
	url string,
	expectedBody string,
) {
	resp, err := client.Get(url)
	if err != nil {
		t.Fatalf("failed to make HTTP request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200 OK, got %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	body := string(bodyBytes)
	if body != expectedBody {
		t.Errorf("response body does match, got %q, expected %q", body, expectedBody)
	}
}

func assertHttpOrHttpsNotFound(
	t *testing.T,
	url string,
) {
	resp, err := client.Get(url)
	if err != nil {
		t.Fatalf("failed to make HTTP request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 404 NOT_FOUND, got %d", resp.StatusCode)
	}
}

func assertHttpOrHttpsFails(
	t *testing.T,
	url string,
) {
	resp, err := client.Get(url)
	if err == nil {
		defer resp.Body.Close()
		t.Errorf("expected HTTPS request to fail, but got status %v", resp.StatusCode)
	}
}

func assertHttpsCertificate(
	t *testing.T,
	url string,
	certPEM []byte,
	shouldMatch bool,
) {
	assert.StartsWith(url, "https://", "assertHttpsCertificate only asserts for HTTPS calls")

	resp, err := client.Get(url)
	if err != nil {
		t.Fatalf("failed to GET from nginx: %v", err)
	}
	defer resp.Body.Close()

	connState := resp.TLS
	if connState == nil || len(connState.PeerCertificates) == 0 {
		t.Fatal("no TLS connection state or peer certificates found")
	}
	serverCert := connState.PeerCertificates[0]

	block, _ := pem.Decode(certPEM)
	if block == nil {
		t.Fatal("failed to decode original cert PEM")
	}

	expectedCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("failed to parse original certificate: %v", err)
	}

	if !expectedCert.Equal(serverCert) == shouldMatch {
		t.Error("certificate receiveid dont match")
	}
}

func TestNginxControllerSetHttpLocation(t *testing.T) {
	controller, cleanup, err := setupTestingNginxController(t)
	if err != nil {
		t.Fatalf("setup failed, got %q", err)
	}
	defer cleanup()

	domain := "localhost"
	serverId := ServerIdentifier{Domain: domain, TlsEnabled: false}
	server := NewServerConfig(serverId, false)
	location := NewLocationConfig(
		serverId,
		"/",
		LocationPrefix,
		"return 200 'foo'",
		"add_header Content-Type text/plain",
	)

	if err := controller.SetHTTPServer(server); err != nil {
		t.Fatalf("failed to set server, got %v", err)
	}
	if err := controller.SetLocation(location); err != nil {
		t.Fatalf("failed to set location, got %v", err)
	}
	if err := controller.ApplyConfig(); err != nil {
		t.Fatalf("failed to apply config, got %v", err)
	}

	expectedResponse := "foo"
	assertHttpOrHttpsOk(t, "http://localhost", expectedResponse)
	assertHttpOrHttpsOk(t, "http://localhost/hello-world", expectedResponse)
	assertHttpOrHttpsOk(t, "http://localhost/hello/world", expectedResponse)
	assertHttpOrHttpsFails(t, "https://localhost")
	assertHttpOrHttpsFails(t, "https://localhost/hello-world")
}

func TestNginxControllerVerifyHttpsCertificate(t *testing.T) {
	domain := "localhost"

	controller, cleanup, err := setupTestingNginxController(t)
	if err != nil {
		t.Fatalf("setup failed, got %q", err)
	}
	defer cleanup()

	serverId := ServerIdentifier{Domain: domain, TlsEnabled: true}
	certPEM, keyPEM := generateSelfSignedCert(domain)
	cert := NewCertificate(domain, keyPEM, certPEM)
	server := NewServerConfig(serverId, false)
	location := NewLocationConfig(
		serverId,
		"/",
		LocationPrefix,
		"return 200 'foo'",
		"add_header Content-Type text/plain",
	)

	if err := controller.SetCertificate(cert); err != nil {
		t.Fatalf("failed to set certificate: %v", err)
	}
	if err := controller.SetHTTPServer(server); err != nil {
		t.Fatalf("failed to set server: %v", err)
	}
	if err := controller.SetLocation(location); err != nil {
		t.Fatalf("failed to set location: %v", err)
	}
	if err := controller.ApplyConfig(); err != nil {
		t.Fatalf("failed to apply config, got %v", err)
	}

	assertHttpsCertificate(t, "https://localhost", certPEM, true)
	assertHttpsCertificate(t, "https://google.com", certPEM, false)
}

func TestNginxControllerChangeCertificate(t *testing.T) {
	domain := "localhost"

	controller, cleanup, err := setupTestingNginxController(t)
	if err != nil {
		t.Fatalf("setup failed, got %q", err)
	}
	defer cleanup()

	serverId := ServerIdentifier{Domain: domain, TlsEnabled: true}
	certPEM1, keyPEM1 := generateSelfSignedCert(domain)
	cert := NewCertificate(domain, keyPEM1, certPEM1)
	server := NewServerConfig(serverId, false)
	location := NewLocationConfig(
		serverId,
		"/",
		LocationPrefix,
		"return 200 'foo'",
		"add_header Content-Type text/plain",
	)

	if err := controller.SetCertificate(cert); err != nil {
		t.Fatalf("failed to set certificate: %v", err)
	}
	if err := controller.SetHTTPServer(server); err != nil {
		t.Fatalf("failed to set server: %v", err)
	}
	if err := controller.SetLocation(location); err != nil {
		t.Fatalf("failed to set location: %v", err)
	}
	if err := controller.ApplyConfig(); err != nil {
		t.Fatalf("failed to apply config, got %v", err)
	}

	assertHttpsCertificate(t, "https://localhost", certPEM1, true)
	assertHttpsCertificate(t, "https://google.com", certPEM1, false)

	certPEM2, keyPEM2 := generateSelfSignedCert(domain)
	cert2 := NewCertificate(domain, keyPEM2, certPEM2)

	if err := controller.SetCertificate(cert2); err != nil {
		t.Fatalf("failed to set certificate: %v", err)
	}
	if err := controller.ApplyConfig(); err != nil {
		t.Fatalf("failed to apply config, got %v", err)
	}

	assertHttpsCertificate(t, "https://localhost", certPEM1, false)
	assertHttpsCertificate(t, "https://localhost", certPEM2, true)
}

func TestNginxControllerSetAndOverwrittingLocations(t *testing.T) {
	domain := "localhost"
	tests := []struct {
		name       string
		urlPrefix  string
		tlsEnabled bool
	}{
		{name: "HTTP", tlsEnabled: false, urlPrefix: "http://localhost"},
		{name: "HTTPS", tlsEnabled: true, urlPrefix: "https://localhost"},
	}

	for _, tt := range tests {
		assert.EndsNotWith(tt.urlPrefix, '/', "url cannot end with '/'")
		certPEM, keyPEM := generateSelfSignedCert(domain)
		cert := NewCertificate(domain, keyPEM, certPEM)

		t.Run(tt.name, func(t *testing.T) {
			controller, cleanup, err := setupTestingNginxController(t)
			if err != nil {
				t.Fatalf("setup failed, got %q", err)
			}
			defer cleanup()

			serverId := ServerIdentifier{Domain: domain, TlsEnabled: tt.tlsEnabled}
			server := NewServerConfig(serverId, false)
			location1 := NewLocationConfig(
				serverId,
				"/test-1",
				LocationExact,
				"return 200 'foo'",
				"add_header Content-Type text/plain",
			)

			if tt.tlsEnabled {
				if err := controller.SetCertificate(cert); err != nil {
					t.Fatalf("failed to set certificate, got %v", err)
				}
			}
			if err := controller.SetHTTPServer(server); err != nil {
				t.Fatalf("failed to set server, got %v", err)
			}
			if err := controller.SetLocation(location1); err != nil {
				t.Fatalf("failed to set location1, got %v", err)
			}
			if err := controller.ApplyConfig(); err != nil {
				t.Fatalf("failed to apply config, got %v", err)
			}

			assertHttpOrHttpsNotFound(t, tt.urlPrefix)
			assertHttpOrHttpsNotFound(t, tt.urlPrefix+"/this-one-should-not-exists")

			url1 := tt.urlPrefix + "/test-1"
			assertHttpOrHttpsOk(t, url1, "foo")

			// add new location, old and new one should be reachable
			location2 := NewLocationConfig(
				serverId,
				"/test-2",
				LocationExact,
				"return 200 'bra'",
				"add_header Content-Type text/plain",
			)
			if err := controller.SetLocation(location2); err != nil {
				t.Fatalf("failed to set location2, got %v", err)
			}
			if err := controller.ApplyConfig(); err != nil {
				t.Fatalf("failed to apply config, got %v", err)
			}

			url2 := tt.urlPrefix + "/test-2"
			assertHttpOrHttpsOk(t, url1, "foo")
			assertHttpOrHttpsOk(t, url2, "bra")

			// overwriting location
			location3 := NewLocationConfig(
				serverId,
				"/test-1",
				LocationExact,
				"return 200 'bra'",
				"add_header Content-Type text/plain",
			)
			if err := controller.SetLocation(location3); err != nil {
				t.Fatalf("failed to set location3, got %v", err)
			}
			if err := controller.ApplyConfig(); err != nil {
				t.Fatalf("failed to apply config, got %v", err)
			}

			assertHttpOrHttpsOk(t, url1, "bra")
			assertHttpOrHttpsOk(t, url2, "bra")
		})
	}

}

func TestNginxControllerUnsetLocation(t *testing.T) {
	domain := "localhost"
	tests := []struct {
		name       string
		urlPrefix  string
		tlsEnabled bool
	}{
		{name: "HTTP", tlsEnabled: false, urlPrefix: "http://localhost"},
		{name: "HTTPS", tlsEnabled: true, urlPrefix: "https://localhost"},
	}

	for _, tt := range tests {
		assert.EndsNotWith(tt.urlPrefix, '/', "url cannot end with '/'")
		certPEM, keyPEM := generateSelfSignedCert(domain)
		cert := NewCertificate(domain, keyPEM, certPEM)

		t.Run(tt.name, func(t *testing.T) {
			controller, cleanup, err := setupTestingNginxController(t)
			if err != nil {
				t.Fatalf("setup failed, got %q", err)
			}
			defer cleanup()

			serverId := ServerIdentifier{Domain: domain, TlsEnabled: tt.tlsEnabled}
			server := NewServerConfig(serverId, false)
			location1 := NewLocationConfig(
				serverId,
				"/test-1",
				LocationExact,
				"return 200 'foo'",
				"add_header Content-Type text/plain",
			)

			if tt.tlsEnabled {
				if err := controller.SetCertificate(cert); err != nil {
					t.Fatalf("failed to set certificate, got %v", err)
				}
			}
			if err := controller.SetHTTPServer(server); err != nil {
				t.Fatalf("failed to set server, got %v", err)
			}
			if err := controller.ApplyConfig(); err != nil {
				t.Fatalf("failed to apply config, got %v", err)
			}

			url := tt.urlPrefix + "/test-1"
			assertHttpOrHttpsNotFound(t, url)
			if err := controller.SetLocation(location1); err != nil {
				t.Fatalf("failed to set location1, got %v", err)
			}
			if err := controller.ApplyConfig(); err != nil {
				t.Fatalf("failed to apply config, got %v", err)
			}

			assertHttpOrHttpsOk(t, url, "foo")

			if err := controller.UnsetLocation(&location1.LocationIdentifier); err != nil {
				t.Fatalf("failed to set location1, got %v", err)
			}
			if err := controller.ApplyConfig(); err != nil {
				t.Fatalf("failed to apply config, got %v", err)
			}

			assertHttpOrHttpsNotFound(t, url)
		})
	}
}

func TestNginxControllerSetSubdomains(t *testing.T) {
	tests := []struct {
		name       string
		tlsEnabled bool
	}{
		{name: "HTTP", tlsEnabled: false},
		{name: "HTTPS", tlsEnabled: true},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			controller, cleanup, err := setupTestingNginxController(t)
			if err != nil {
				t.Fatalf("setup failed, got %q", err)
			}
			defer cleanup()
			domains := []string{"localhost", "app.localhost", "admin.localhost"}
			for _, domain := range domains {
				serverId := ServerIdentifier{Domain: domain, TlsEnabled: tt.tlsEnabled}
				server := NewServerConfig(serverId, false)
				location := NewLocationConfig(
					serverId,
					"/test",
					LocationExact,
					fmt.Sprintf("return 200 '%s'", domain),
					"add_header Content-Type text/plain",
				)

				if tt.tlsEnabled {
					certPEM, keyPEM := generateSelfSignedCert(domain)
					cert := NewCertificate(domain, keyPEM, certPEM)
					if err := controller.SetCertificate(cert); err != nil {
						t.Fatalf("failed to set certificate, got %v", err)
					}
				}

				if err := controller.SetHTTPServer(server); err != nil {
					t.Fatalf("failed to set server, got %v", err)
				}
				if err := controller.SetLocation(location); err != nil {
					t.Fatalf("failed to set location, got %v", err)
				}
			}

			if err := controller.ApplyConfig(); err != nil {
				t.Fatalf("failed to apply config, got %v", err)
			}

			for _, domain := range domains {
				port := "http"
				if tt.tlsEnabled {
					port = "https"
				}

				url := port + "://" + domain + "/test"
				assertHttpOrHttpsOk(t, url, domain)
			}

		})
	}
}
