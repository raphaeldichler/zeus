// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package nginxcontroller

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"os/exec"
	"time"

	"github.com/raphaeldichler/zeus/internal/assert"
	"github.com/raphaeldichler/zeus/internal/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	NginxConfigPath  = "/etc/nginx/nginx.conf"
	SocketPath       = "/run/zeus/nginx.sock"
	SocketMountPath  = "/run/zeus"
	NginxPidFilePath = "/run/nginx.pid"
)

type Controller struct {
	UnimplementedNginxControllerServer
	server   *grpc.Server
	listener net.Listener
	nginx    string
	log      *log.Logger
	config   *IngressRequest
}

func New() (*Controller, error) {
	nginx, err := exec.LookPath("nginx")
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(SocketPath); err == nil {
		if err := os.Remove(SocketPath); err != nil {
			return nil, err
		}
	}

	listen, err := net.Listen("unix", SocketPath)
	if err != nil {
		return nil, err
	}

	if err := os.Chmod(SocketPath, 0666); err != nil {
		return nil, err
	}

	s := grpc.NewServer()
	srv := &Controller{
		server:   s,
		listener: listen,
		nginx:    nginx,
		log:      log.New("nginx", "controller"),
	}
	RegisterNginxControllerServer(s, srv)

	return srv, nil
}

func (self *Controller) Run() error {
	defer self.listener.Close()
	return self.server.Serve(self.listener)
}

func (self *Controller) reloadNginxConfig() error {
	cmd := exec.Command(self.nginx, "-s", "reload")
	out, err := cmd.CombinedOutput()
	self.log.Info("Reload nginx config. Got '%s'", string(out))

	return err
}

func (self *Controller) storeAndApplyConfig(d directory) error {
	err := self.config.storeAsNginxConfig(d)
	if err != nil {
		return err
	}

	if err := self.reloadNginxConfig(); err != nil {
		return err
	}

	time.Sleep(time.Millisecond * 500)
	return nil
}

func (self *Controller) SetIngressConfig(
	ctx context.Context,
	req *IngressRequest,
) (*IngressResponse, error) {
	d, err := openDirectory()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to open a new direcotry: %v", err)
	}
	defer time.AfterFunc(time.Minute, func() { d.close() })

	old := self.config
	self.config = req
	if err := self.storeAndApplyConfig(d); err != nil {
		self.config = old
		return nil, status.Errorf(codes.Internal, "Failed to load new nginx config: %v", err)
	}

	return &IngressResponse{}, nil
}

func (self *Controller) GenerateCertificates(
	ctx context.Context,
	req *GenerateCertificateRequest,
) (*GenerateCertificateResponse, error) {

	switch req.Type {
	case GenerateCertificateType_AuthoritySigned:

		certBundle, err := obtainCertificate(
			self,
			req.Domain,
			req.CertificateEmail,
		)
		if err != nil {
			return nil, status.Errorf(codes.Unknown, "Failed to obtain certificate: %v", err)
		}

		return &GenerateCertificateResponse{
			Fullchain: string(certBundle.FullchainPem),
			Privkey:   string(certBundle.PrivKeyPem),
		}, nil

	case GenerateCertificateType_SelfSigned:
		priv, err := rsa.GenerateKey(rand.Reader, 2048)
		assert.ErrNil(err)
		now := time.Now()
		expires := now.Add(TlsRenewThreshold)

		serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
		assert.ErrNil(err)

		template := x509.Certificate{
			SerialNumber: serial,
			Subject: pkix.Name{
				CommonName: req.Domain,
			},
			NotBefore:             now,
			NotAfter:              expires,
			KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
			ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
			BasicConstraintsValid: true,
			DNSNames:              []string{req.Domain},
		}

		certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
		assert.ErrNil(err)

		certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
		keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

		return &GenerateCertificateResponse{
			Fullchain: string(certPEM),
			Privkey:   string(keyPEM),
		}, nil

	default:
		assert.Unreachable("not all cases are covered")
	}

	return nil, nil
}
