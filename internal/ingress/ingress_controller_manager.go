// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package ingress

import "time"

type IngressTlsState int

const (
	// If in this state it descibes that a TLS certificate is wanted but none has been obtained yet.
	IngressTlsObtain IngressTlsState = iota
	// If in this state it descibes that a TLS certificate is obtained, but if the time is after the expired time
	IngressTlsRenew
)

type IngressTls struct {
	State        IngressTlsState
	Expires      time.Time
	PrivkeyPem   []byte
	FullchainPem []byte
}

type IngressPath struct {
	Path            string
	Matching        LocationMatchingType
	ServiceEndpoint string
}

type IngressRule struct {
	Host  string
	Tls   *IngressRule
	Paths []IngressPath
}

type IngressState struct {
	Application string
	IPv6Enabled bool
	Rules       []IngressRule
}

type IngressControllerManager struct {
	controllers map[string]NginxController
}

func (self *IngressControllerManager) getOrInitController(_ string) (*NginxController, error) {

	return nil, nil
}

func (self *IngressControllerManager) Sync(state *IngressState) error {
	controller, err := self.getOrInitController(state.Application)
	if err != nil {
		return err
	}

	// todo: we need caching tests for the nginx server, if we need to rollback in an invalid state
	// todo: we need to handle unsetting paths also. if the state nolonger defines a location or server we need to remove it

	// for caching state: we wrap this into one context. if the context gets no errors it applies it and updates the cache
	// system crash is no problem, because we can recover from the state at any time
	// if during the setting the state crashes, we dont apply the new data, and keep the old one.
	// because we only apply config if we are successful invalid data inside the container is not important

	// important: if the nginx server crashes it will apply the config -> this might be in an invalid state.
	// we dont ensure correct state inside the filesystem of the nginx server.
	// we have our own system which will handle the. on crash the system will than remove the container and trigger a state recovery

	// checking if the current nginx controller has ipv6 enabled or disabled? we dont need to. the controller knows it
	// if set the controller will not apply any state change

	return controller.Transaction(func(ctr *NginxController) error {

		for _, rule := range state.Rules {
			tlsEnabled := rule.Tls != nil
			serverID := ServerIdentifier{
				Domain:     rule.Host,
				TlsEnabled: tlsEnabled,
			}

			if tlsEnabled {
				// if tls is enabled we need to obtain or renew (or do nothing) a new certificates.
				// if we set and apply a server without setting it. we have an invalid state, which fails the server
			}

			serverConfig := NewServerConfig(
				serverID,
			)
			if err := controller.SetHTTPServer(serverConfig); err != nil {
				// problem: what do we do if we have parital setted state. which was not applied?
				// the controller think they are applied, but they are not ...
				// -> we need to handle parital setted cache
				return err
			}

			for _, path := range rule.Paths {
				locationID := LocationIdentifier{
					ServerIdentifier: serverID,
					Path:             path.Path,
					Matching:         path.Matching,
				}

				locationConfig := NewLocationConfig(
					locationID,
					"proxy_pass "+path.ServiceEndpoint,
					"proxy_set_header Host $host",
					"proxy_set_header X-Real-IP $remote_addr",
					"proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for",
					"proxy_set_header X-Forwarded-Proto $scheme",
				)
				if err := controller.SetLocation(locationConfig); err != nil {
					return err
				}
			}
		}

		return nil
	})
}

// todo: writes the response directly into the request.
func (self *IngressControllerManager) Inspect() error {

	return nil
}
