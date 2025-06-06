// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package ingress

import (
	"net/http"
	"path/filepath"
	"time"

	"github.com/raphaeldichler/zeus/internal/assert"
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

func (self *IngressDaemon) hostNginxControllerSocket() string {
	return filepath.Join("/run/zeus/", self.application, "/ingress/nginx.sock")
}

// Ensures that an ingress container is create and running.
//
// If no container exists a new one is created and started.
// On fatal errors no container will be set and the application state will be updated accrodingly.
func (self *IngressDaemon) ensureContainer(state *record.ApplicationRecord) {
	if self.network == nil {
		self.network = runtime.SelectNetwork(
			[]runtime.Label{
				{Value: "zeus.object.type", Key: "network"},
				{Value: "zeus.application.name", Key: self.application},
			},
		)
	}

	assert.NotNil(self.network, "network must be selected; network is created on applicaiton creation")
	if self.container == nil {
		// we dont have registered a container
		self.container = runtime.SelectContainer(
			[]runtime.Label{
				{Value: "zeus.object.type", Key: "ingress"},
				{Value: "zeus.object.image", Key: state.Ingress.Metadata.Image},
				{Value: "zeus.application.name", Key: self.application},
			},
		)
	}

	if self.container == nil {
		// we have not found the correct container, we create one
		// with that we ensure ingress auto update -> if we update the version of the

		c, err := runtime.NewContainer(
			self.application,
			"ingress-nginx",
			runtime.WithImage("raphaeldichler/zeus-nginx:1.0.0"),
			runtime.WithPulling(),
			runtime.WithExposeTcpPort("80", "80"),
			runtime.WithExposeTcpPort("443", "443"),
			runtime.WithConnectedToNetwork(self.network),
			runtime.WithLabel("zeus.object.type", "ingress"),
			runtime.WithLabel("zeus.object.image", "zeus-nginx:v1.0.0"),
			runtime.WithLabel("zeus.application.name", self.application),
			runtime.WithMount(self.hostNginxControllerSocket(), nginxcontroller.SocketPath),
		)
		if err != nil {
			state.Ingress.Errors.SetServerError(
				"FailedCreatingIngressContainer",
				"*",
				err.Error(),
			)
			return
		}
		self.container = c
	}
	assert.NotNil(self.container, "at this state the container was set correctly")

	self.nginxControllerClient = nginxcontroller.NewClient(self.application, self.hostNginxControllerSocket())
	self.acmeProvider = NewNginxChallengeProvider(self.nginxControllerClient, self.application)
}

func (self *IngressDaemon) Sync(state *record.ApplicationRecord) {
	self.log.Info("Starting IngressControllerManager.Sync() - syncing ingress controllers")
	if !state.Ingress.Enabled() {
		self.log.Info("Nothing todo. Completed syncing ingress controllers")
		return
	}

	self.ensureContainer(state)
	self.syncTlsCertificates(state)

	req := nginxcontroller.NewApplyRequest()
	for _, server := range state.Ingress.Servers {
		tls := server.Tls
		if tls != nil && tls.State == record.TlsObtain {
			// we skip it, tls sync failed - server
			// dependency error - tls certificate failed so this fails
			continue
		}

		opts := nginxcontroller.NewServerRequestOptions()
		if server.IPv6 {
			opts.Add(nginxcontroller.WithIPv6())
		}

		if tls != nil {
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

	response, err := self.nginxControllerClient.SetConfig(req)
	if err != nil {
		state.Ingress.Errors.SetServerError(
			"FailedSendingIngressConfig",
			"*",
			err.Error(),
		)
		return
	}
	defer response.Body.Close()
	assert.True(
		response.StatusCode == http.StatusCreated || response.StatusCode == http.StatusInternalServerError,
		"a request cannot result in a bad request or any other problems",
	)

	if response.StatusCode == http.StatusInternalServerError {
		state.Ingress.Errors.SetServerError(
			"FailedApplyingIngressConfig",
			"*",
			err.Error(),
		)
	}

	self.log.Info("Completed syncing ingress controllers")
}

func (self *IngressDaemon) syncTlsCertificates(state *record.ApplicationRecord) {
	assert.NotNil(self.acmeProvider, "using the acme provider in the next steps, it must exists")

	for _, server := range state.Ingress.Servers {
		tls := server.Tls
		if tls == nil {
			// no certificate is required for this server
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
				"FailedObtainCertificate",
				server.Host,
				err.Error(),
			)
			continue
		}

		tls.FullchainPem = certBundle.FullchainPem
		tls.PrivkeyPem = certBundle.PrivKeyPem
		tls.State = record.TlsRenew
		tls.Expires = time.Now().Add(TlsNewRenewThreshold)
	}
}
