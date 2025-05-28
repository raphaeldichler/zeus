// Copyright 2025 The Zeus Authors.
// Licensed under the MIT License. See the LICENSE file for details.

package runtime

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	"github.com/raphaeldichler/zeus/internal/assert"
	"github.com/raphaeldichler/zeus/internal/log"
)

var (
	ErrContainterCannotStart  = errors.New("cannot start the container")
	ErrCommandFailedToExecute = errors.New("command existed with error code")
)

type ContainerConfig struct {
	config        *container.Config
	hostConfig    *container.HostConfig
	networkConfig *network.NetworkingConfig

	// The image which will be used to create, pull, or run the container
	img string
	// Pulls images if its does not exists on the machine
	doPull bool
	// Times to try to start a container before giving up. default 3
	retryStart int
	// Files which are copied into the container before it will be started
	filesToCopyInto []FileContent

	log *log.Logger
}

// Pulls, creates, and starts the container according to the config
func (self *ContainerConfig) startContainer() (*Container, error) {
	if self.doPull {
		if err := pull(self.img); err != nil {
			return nil, err
		}
	}

	self.config.Image = self.img
	containerID, err := create(
		self.config,
		self.hostConfig,
		self.networkConfig,
	)
	if err != nil {
		return nil, err
	}

	container := &Container{
		id:     containerID,
		client: c,
		log:    self.log,
	}

	if err := container.CopyInto(self.filesToCopyInto...); err != nil {
		return nil, err
	}

	if err := start(containerID, self.retryStart); err != nil {
		return nil, err
	}

	return container, nil
}

func defaultContainerConfig(logger *log.Logger) *ContainerConfig {
	return &ContainerConfig{
		config: &container.Config{},
		hostConfig: &container.HostConfig{
			AutoRemove: true,
		},
		networkConfig:   &network.NetworkingConfig{},
		img:             "",
		doPull:          false,
		retryStart:      3,
		log:             logger,
		filesToCopyInto: []FileContent{},
	}
}

type ContainerOption func(cfg *ContainerConfig)

func WithImage(img string) ContainerOption {
	return func(cfg *ContainerConfig) {
		cfg.img = img
	}
}

func WithPulling() ContainerOption {
	return func(cfg *ContainerConfig) {
		cfg.doPull = true
	}
}

func WithCmd(cmd ...string) ContainerOption {
	return func(cfg *ContainerConfig) {
		cfg.config.Cmd = cmd
	}
}

func WithExposeTcpPort(hostPort string, containerPort string) ContainerOption {
	assert.StartsNotWithString(hostPort, "tcp/", "we will append tcp/ if needed")
	assert.StartsNotWithString(containerPort, "tcp/", "we will append tcp/ if needed")

	return func(cfg *ContainerConfig) {
		if cfg.hostConfig.PortBindings == nil {
			cfg.hostConfig.PortBindings = make(nat.PortMap)
		}

		if cfg.config.ExposedPorts == nil {
			cfg.config.ExposedPorts = make(nat.PortSet)
		}

		cfg.config.ExposedPorts[nat.Port(containerPort+"/tcp")] = struct{}{}
		cfg.hostConfig.PortBindings[nat.Port(containerPort+"/tcp")] = []nat.PortBinding{
			{
				HostPort: hostPort,
			},
		}
	}
}

func WithConnectedToNetwork(nt *Network) ContainerOption {
	return func(cfg *ContainerConfig) {
		cfg.networkConfig.EndpointsConfig = map[string]*network.EndpointSettings{
			nt.NetworkName(): {},
		}
	}
}

func WithLabel(key string, value string) ContainerOption {
	return func(cfg *ContainerConfig) {
		if cfg.config.Labels == nil {
			cfg.config.Labels = make(map[string]string)
		}

		cfg.config.Labels[key] = value
	}
}

func WithObjectTypeLabel(objectType string) ContainerOption {
	return WithLabel("zeus.object.type", objectType)
}

func WithCopyIntoBeforeStart(file FileContent) ContainerOption {
	return func(cfg *ContainerConfig) {
		cfg.filesToCopyInto = append(cfg.filesToCopyInto, file)
	}
}

type Container struct {
	id     string
	client *client.Client

	log *log.Logger
}

func NewContainer(
	application string,
	daemon string,
	options ...ContainerOption,
) (*Container, error) {
	assert.NotNil(c, "init of docker-client failed")
	logger := log.New(application, daemon)
	cfg := defaultContainerConfig(logger)

	for _, opt := range options {
		opt(cfg)
	}

	return cfg.startContainer()
}

func (self *Container) DisconnectNetwork(network *Network) error {
	ctx := context.Background()
	return self.client.NetworkDisconnect(ctx, network.NetworkName(), self.id, false)
}

func (self *Container) String() string {
	return self.id
}

func (self *Container) Shutdown() error {
	ctx := context.Background()
	return self.client.ContainerStop(
		ctx,
		self.id,
		container.StopOptions{
			Timeout: nil,
		},
	)
}

func (self *Container) IsRunning() (bool, error) {
	ctx := context.Background()
	resp, err := self.client.ContainerInspect(ctx, self.id)
	if err != nil {
		return false, err
	}

	return resp.State.Running, nil
}

type CmdResult struct {
	exitCode int
	stdout   string
}

func (self *Container) runCommand(cmd ...string) (*CmdResult, error) {
	ctx := context.Background()
	resp, err := self.client.ContainerExecCreate(
		ctx,
		self.id,
		container.ExecOptions{
			Cmd:          cmd,
			AttachStdin:  true,
			AttachStdout: true,
		},
	)
	if err != nil {
		return nil, err
	}

	attach, err := self.client.ContainerExecAttach(
		ctx,
		resp.ID,
		container.ExecAttachOptions{},
	)
	if err != nil {
		return nil, err
	}
	defer attach.Close()

	var outputBuf bytes.Buffer
	_, err = stdcopy.StdCopy(&outputBuf, &outputBuf, attach.Reader)
	if err != nil {
		self.log.Error("Failed to copy stdout for (%s)", self)
		return nil, err
	}

	insp, err := self.client.ContainerExecInspect(ctx, resp.ID)
	if err != nil {
		return nil, err
	}
	self.log.Debug("run command \t\t'%s'; exitCode %d", cmd, insp.ExitCode)

	// todo: add debug logging; see what command are executed
	return &CmdResult{
		exitCode: insp.ExitCode,
		stdout:   outputBuf.String(),
	}, nil
}

func (self *Container) ExitsPath(path string) (bool, error) {
	result, err := self.runCommand("test", "-e", path)
	if err != nil {
		return false, err
	}

	return result.exitCode == 0, nil
}

func (self *Container) ReadFile(path string) (string, error) {
	result, err := self.runCommand("cat", path)
	if err != nil {
		return "", err
	}

	if result.exitCode != 0 {
		return "", errors.New("failed to execute command")
	}

	return result.stdout, nil
}

func (self *Container) RemoveFile(path string) error {
	result, err := self.runCommand("rm", "-f", path)
	if err != nil {
		return err
	}

	if result.exitCode != 0 {
		return ErrCommandFailedToExecute
	}

	return nil
}

func (self *Container) EnsurePathExists(path string) error {
	result, err := self.runCommand("mkdir", "-p", path)
	if err != nil {
		return err
	}

	if result.exitCode != 0 {
		return errors.New("failed to execute command")
	}

	return nil
}

type FileContent interface {
	FilePath() string
	FileContent() []byte
}

type BasicFileContent struct {
	Path    string
	Content []byte
}

func (self *BasicFileContent) FilePath() string {
	return self.Path
}

func (self *BasicFileContent) FileContent() []byte {
	return self.Content
}

func (self *Container) CopyInto(files ...FileContent) error {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	for _, fc := range files {
		content := fc.FileContent()
		hdr := &tar.Header{
			Name:    fc.FilePath(),
			Mode:    0644,
			Size:    int64(len(content)),
			ModTime: time.Now(),
		}

		err := tw.WriteHeader(hdr)
		assert.ErrNil(err)
		_, err = tw.Write(content)
		assert.ErrNil(err)

	}

	err := tw.Close()
	assert.ErrNil(err)

	tarReader := bytes.NewReader(buf.Bytes())
	return self.client.CopyToContainer(
		context.Background(),
		self.id,
		"/",
		tarReader,
		container.CopyToContainerOptions{},
	)
}

func (self *Container) Sighup() error {
	ctx := context.Background()
	if err := self.client.ContainerKill(
		ctx,
		self.id,
		"SIGHUP",
	); err != nil {
		return err
	}

	return nil
}

func (self *Container) AssertPathExists(path string) error {
	self.log.Debug("assert path exists \t'%s'", path)
	exists, err := self.ExitsPath(path)
	if err != nil {
		return err
	}
	assert.True(exists, "location requires path to exists. setting up server should ensure it")

	return nil
}

func (self *Container) AssertIsRunning() error {
	runs, err := self.IsRunning()
	if err != nil {
		return err
	}
	assert.True(runs, "a container must run to copy data into")

	return nil
}
