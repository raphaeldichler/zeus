// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package ingress

import (
	"net/http"
	"path/filepath"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/go-connections/nat"
	"github.com/raphaeldichler/zeus/internal/assert"
	"github.com/raphaeldichler/zeus/internal/ingress/errtype"
	"github.com/raphaeldichler/zeus/internal/log"
	"github.com/raphaeldichler/zeus/internal/nginxcontroller"
	"github.com/raphaeldichler/zeus/internal/record"
	"github.com/raphaeldichler/zeus/internal/runtime"
)

const (
	// Time until the ingress-daemon will consider the tls certificate to renew.
	// Aka if the time left until the certificate expires is less than this threshold
	TlsRenewThreshold = time.Hour * 24
	// Time until we will renew the certificate
	TlsNewRenewThreshold = time.Hour * 24 * 40
)

type IngressDaemon struct {
	application string

	nginxControllerClient *nginxcontroller.Client
	network               *runtime.Network
	container             *runtime.Container
	acmeProvider          *NginxChallengeProvider

	log *log.Logger
}

// Creates a new IngressDaemon which handles the state of the ingress object.
func NewIngress(application string) *IngressDaemon {
	logger := log.New(
		application, "ingress-daemon",
	)

	return &IngressDaemon{
		nginxControllerClient: nil,
		container:             nil,
		acmeProvider:          nil,
		log:                   logger,
	}
}

func (self *IngressDaemon) Sync(state *record.ApplicationRecord) {
	self.log.Info("Starting IngressControllerManager.Sync() - syncing ingress controllers")
	defer self.log.Info("Completed syncing ingress controllers")
	if !state.Ingress.Enabled() {
		return
	}

	self.ensureNetwork(state)
	self.ensureContainer(state)
	if self.container == nil {
		return
	}
	self.syncTlsCertificates(state)

	response, err := self.nginxControllerClient.SetConfig(self.buildIngressConfigRequest(state))
	if err != nil {
		state.Ingress.Errors.SetIngressError(
			errtype.FailedInteractionWithNginxController(errtype.NginxSend, err),
		)
		return
	}
	defer response.Body.Close()
	assert.True(
		response.StatusCode == http.StatusCreated || response.StatusCode == http.StatusInternalServerError,
		"a request cannot result in a bad request or any other problems",
	)

	if response.StatusCode == http.StatusInternalServerError {
		state.Ingress.Errors.SetIngressError(
			errtype.FailedInteractionWithNginxController(errtype.NginxApply, err),
		)
	}
}

func (self *IngressDaemon) hostNginxControllerSocketPath() string {
	return filepath.Join(self.hostNginxControllerMountPath(), "nginx.sock")
}

func (self *IngressDaemon) hostNginxControllerMountPath() string {
	return filepath.Join("/run/zeus/", self.application, "/ingress/")
}

func (self *IngressDaemon) ensureNetwork(state *record.ApplicationRecord) {
	if self.network == nil {
		self.network = runtime.SelectNetwork(
			[]runtime.Label{
				runtime.ObjectImageLabel("network"),
				runtime.ApplicationNameLabel(self.application),
			},
		)
	}
}

// Ensures that an ingress container is create and running.
//
// If no container exists a new one is created and started.
// On fatal errors no container will be set and the application state will be updated accrodingly.
func (self *IngressDaemon) ensureContainer(state *record.ApplicationRecord) {
	assert.NotNil(self.network, "network must be selected; network is created on applicaiton creation")
	if self.container == nil {
		selectedContainers, err := runtime.SelectContainer(
			[]runtime.Label{
				runtime.ObjectTypeLabel("ingress"),
				runtime.ObjectImageLabel(state.Ingress.Metadata.Image),
				runtime.ApplicationNameLabel(self.application),
			},
		)
		if err != nil {
			state.Ingress.Errors.SetIngressError(
				errtype.FailedInteractionWithDockerDaemon(errtype.DockerSelect, err),
			)
		}

		switch len(selectedContainers) {
		case 0:
			self.container = nil

		case 1:
			self.container = selectedContainers[0]

		default:
			assert.Unreachable(
				"Too many ingress container exists in the current context. Possible external tampering.",
			)
		}
	}

	if self.container == nil {
		c, err := runtime.NewContainer(
			self.application,
			"ingress-nginx",
			runtime.WithImage(state.Ingress.Metadata.Image),
			runtime.WithPulling(),
			runtime.WithExposeTcpPort("80", "80"),
			runtime.WithExposeTcpPort("443", "443"),
			runtime.WithConnectedToNetwork(self.network),
			runtime.WithLabels(
				runtime.ObjectTypeLabel("ingress"),
				runtime.ObjectImageLabel(state.Ingress.Metadata.Image),
				runtime.ApplicationNameLabel(self.application),
			),
			runtime.WithMount(self.hostNginxControllerMountPath(), nginxcontroller.SocketMountPath),
		)
		if err != nil {
			state.Ingress.Errors.SetIngressError(
				errtype.FailedInteractionWithDockerDaemon(errtype.DockerCreate, err),
			)
			return
		}
		self.container = c
	}
	assert.NotNil(self.container, "at this state the container was set correctly")

	if ok := self.validateContainer(state); !ok {
		if err := self.container.Shutdown(); err != nil {
			state.Ingress.Errors.SetIngressError(
				errtype.FailedInteractionWithDockerDaemon(errtype.DockerStop, err),
			)
		}
		self.container = nil
		return
	}

	self.nginxControllerClient = nginxcontroller.NewClient(self.application, self.hostNginxControllerSocketPath())
	self.acmeProvider = NewNginxChallengeProvider(self.nginxControllerClient, self.application)
}

func (self *IngressDaemon) validateContainer(state *record.ApplicationRecord) bool {
	assert.NotNil(self.container, "at this state the container was set correctly")

	inspect, err := self.container.Inspect()
	if err != nil {
		state.Ingress.Errors.SetIngressError(
			errtype.FailedInteractionWithDockerDaemon(errtype.DockerInspect, err),
		)
		return false
	}

	mounts := inspect.Mounts
	if len(mounts) != 1 {
		return false
	}
	socketMount := mounts[0]
	correctSocketMount := container.MountPoint{
		Type:        mount.TypeBind,
		Source:      nginxcontroller.SocketMountPath,
		Destination: self.hostNginxControllerMountPath(),
		Mode:        "",
		RW:          true,
		Propagation: "rprivate",
	}
	if socketMount != correctSocketMount {
		return false
	}

	portBindings := inspect.HostConfig.PortBindings
	if portBindings == nil {
		return false
	}
	tcp443, ok := portBindings[nat.Port("443/tcp")]
	if !ok {
		return false
	}
	correctTcp443 := nat.PortBinding{HostPort: "443", HostIP: ""}
	if len(tcp443) != 1 || tcp443[0] != correctTcp443 {
		return false
	}
	tcp80, ok := portBindings[nat.Port("80/tcp")]
	if !ok {
		return false
	}
	correctTcp80 := nat.PortBinding{HostPort: "80", HostIP: ""}
	if len(tcp80) != 1 || correctTcp80 != tcp80[0] {
		return false
	}

	return true
}

func (self *IngressDaemon) syncTlsCertificates(state *record.ApplicationRecord) {
	assert.NotNil(self.acmeProvider, "using the acme provider in the next steps, it must exists")

	for _, server := range state.Ingress.Servers {
		tls := server.Tls
		if tls == nil {
			continue
		}

		if tls.State == record.TlsRenew && time.Since(tls.Expires) > TlsRenewThreshold {
			continue
		}

		// we dont differatiate between renew and obtain. renew can result in obtaining a new one
		// therefore we just obtain a new one to reduce the possiblilties of failures
		certBundle, err := ObtainCertificate(
			self.acmeProvider,
			server.Host,
			tls.CertificateEmail,
		)
		if err != nil {
			state.Ingress.Errors.SetTlsError(
				errtype.FailedObtainCertificate(server.Host, err),
			)
			continue
		}

		tls.FullchainPem = certBundle.FullchainPem
		tls.PrivkeyPem = certBundle.PrivKeyPem
		tls.State = record.TlsRenew
		tls.Expires = time.Now().Add(TlsNewRenewThreshold)
	}
}

func (self *IngressDaemon) buildIngressConfigRequest(state *record.ApplicationRecord) *nginxcontroller.ApplyRequest {
	req := nginxcontroller.NewApplyRequest()
	for _, server := range state.Ingress.Servers {
		if state.Ingress.Errors.ExistsTlsError(errtype.FailedObtainCertificateQuery(server.Host)) {
			continue
		}

		opts := nginxcontroller.NewServerRequestOptions()
		opts.Add(
			nginxcontroller.WithDomain(server.Host),
			nginxcontroller.WithIPv6Enabled(server.IPv6),
		)

		if tls := server.Tls; tls != nil {
			opts.Add(
				nginxcontroller.WithCertificate(
					string(tls.PrivkeyPem),
					string(tls.FullchainPem),
				),
			)
		}

		for _, paths := range server.HTTP.Paths {
			serviceEndpoint := state.Service.GetEndpoint(paths.Service)

			opts.Add(
				nginxcontroller.WithLocation(
					paths.Path,
					paths.Matching,
					serviceEndpoint,
				),
			)
		}

		req.AddServer(opts.Options...)
	}

	return req
}
