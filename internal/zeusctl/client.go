// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package zeusctl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/raphaeldichler/zeus/internal/assert"
	"github.com/raphaeldichler/zeus/internal/zeusapiserver"
	"github.com/raphaeldichler/zeus/internal/zeusctl/formatter"
	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v3"
)

const (
	sshTimeout       = 5 * time.Second
	httpTimeout      = 5 * time.Second
	unixServerSocket = "/run/zeus/zeusd.sock"
)

type contextProvider struct {
	// client returns the client. But the client is not initialized until the run command is called by Cobra
	client *client

	// outputFormatter returns the output formatter. But the output formatter is not initialized until the run command is call
	outputFormatter formatter.Output
}

type ConfigLocalhost struct {
	Enabled bool `yaml:"enabled"`
}

type ConfigRemote struct {
	Port string `yaml:"port"`
	IP   string `yaml:"ip"`
	SSH  struct {
		User string `yaml:"user"`
		Cert string `yaml:"cert"`
	} `yaml:"ssh"`
}

type Config struct {
	Application string          `yaml:"application"`
	Localhost   ConfigLocalhost `yaml:"localhost"`
	Remote      ConfigRemote    `yaml:"remote"`
}

func (c *Config) remoteSSHDialer() (*http.Transport, error) {
	signer, err := ssh.ParsePrivateKey([]byte(c.Remote.SSH.Cert))
	if err != nil {
		return nil, fmt.Errorf("cannot parse SSH key: %w", err)
	}

	config := &ssh.ClientConfig{
		User: c.Remote.SSH.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         sshTimeout,
	}

	sshAddr := fmt.Sprintf("%s:22", c.Remote.IP)
	sshConn, err := ssh.Dial("tcp", sshAddr, config)
	if err != nil {
		return nil, fmt.Errorf("SSH dial failed: %w", err)
	}

	return &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return sshConn.Dial("unix", unixServerSocket)
		},
	}, nil
}

func (c *Config) localUnixDialer() (*http.Transport, error) {
	return &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return net.Dial("unix", unixServerSocket)
		},
	}, nil
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		Localhost: ConfigLocalhost{
			Enabled: false,
		},
		Remote: ConfigRemote{
			Port: "22",
		},
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) newClient() (*client, error) {
	var (
		transporter *http.Transport = nil
		err         error           = nil
	)

	if c.Localhost.Enabled {
		transporter, err = c.localUnixDialer()
		if err != nil {
			return nil, err
		}
	} else {
		transporter, err = c.remoteSSHDialer()
		if err != nil {
			return nil, err
		}
	}

	return &client{
		http: &http.Client{
			Transport: transporter,
			Timeout:   httpTimeout,
		},
		application: c.Application,
	}, nil
}

type client struct {
	http        *http.Client
	application string
}

func unixURL(path string) string {
	assert.StartsWithString(path, "/", "path must start with '/'")
	return fmt.Sprintf("http://unix%s", path)
}

func toObject[T any](r io.ReadCloser) *T {
	defer r.Close()
	obj := new(T)
	err := json.NewDecoder(r).Decode(obj)
	assert.ErrNil(err)
	return obj
}

func toError(resp *http.Response) string {
  bad := toObject[zeusapiserver.BadRequest](resp.Body)
	return fmt.Sprintf("Error from server (%s): %s", http.StatusText(resp.StatusCode), bad.Message)
}

func (c *client) applicationEnabled(application string) string {
	req, err := http.NewRequest(
		"POST",
		unixURL(zeusapiserver.EnableApplicationAPIPath(application)),
		nil,
	)
	assert.ErrNil(err)

	resp, err := c.http.Do(req)
	failOnError(err, "Request failed: %v", err)
  switch resp.StatusCode {
    case http.StatusNoContent:
      return "Enabled"
    case http.StatusBadRequest:
      return toError(resp)
    default:
      assert.Unreachable("cover all cases of status code")
  }

	return "" 
}
