package main

import (
	"fmt"

	"github.com/raphaeldichler/zeus/internal/ingress"
	"github.com/raphaeldichler/zeus/internal/runtime"
)

func main() {
	nw, err := runtime.NewNetwork("poseidon", "network")
	if err != nil {
		fmt.Println(err)
		return
	}

	c, err := ingress.DefaultContainer("poseidon", nw)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(c)

	c.DisconnectNetwork(nw)

}

/*
import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"log"

	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
)

type MyUser struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

func (u *MyUser) GetEmail() string {
	return u.Email
}

func (u *MyUser) GetRegistration() *registration.Resource {
	return u.Registration
}

func (u *MyUser) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

type MyHTTPProvider struct{}

func (p *MyHTTPProvider) Present(domain, token, keyAuth string) error {
	// Construct challenge path and content
	// ACME HTTP-01 challenge expects: http://<domain>/.well-known/acme-challenge/<token>
  fmt.Println("Present()")

  fmt.Println(domain)
  fmt.Println(token)
  fmt.Println(keyAuth)

	return nil
}

func (p *MyHTTPProvider) CleanUp(domain, token, keyAuth string) error {
  fmt.Println("CleanUp()")

  fmt.Println(domain)
  fmt.Println(token)
  fmt.Println(keyAuth)

	return nil
}


func main() {
  privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatal(err)
	}

  fmt.Println(privateKey.PublicKey)

	myUser := &MyUser{
		Email: "raphael@dichler.com",
		key:   privateKey,
	}

	config := lego.NewConfig(myUser)
	config.CADirURL = lego.LEDirectoryStaging
	client, err := lego.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}

  reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		log.Fatal(err)
	}
	myUser.Registration = reg

  err = client.Challenge.SetHTTP01Provider(&MyHTTPProvider{})
	if err != nil {
		log.Fatal(err)
	}


	// Request certificate
	request := certificate.ObtainRequest{
		Domains: []string{"myessaymentor.com"},
		Bundle:  true,
	}
	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
		log.Fatal(err)
	}

	// Use certificates.Certificate and certificates.PrivateKey as needed
	log.Printf("Certificate:\n%s", certificates.Certificate)
	log.Printf("Private Key:\n%s", certificates.PrivateKey)

  server := ingress.NginxServerConfig{
    ServerName: "myessaymentor.com",
    TlsEnabled: true,
    Ipv6: true,
    Entries: []string{},
    Locations: []ingress.NginxServerLocationConfig{
      {
        Path: "/",
        Matching: ingress.LocationPrefix,
        Entries: []string{
          "proxy_pass http://backend:8000",
        },
      },
      {
        Path: "/rick",
        Matching: ingress.LocationPrefix,
        Entries: []string{
          "proxy_pass http://rick:5000",
        },
      },

    },
  }

  server.Print()


	// Define tmpfs file location (in-memory)
	tmpfsDir := "/dev/shm"
	fileName := "secret.txt"
	fullPath := filepath.Join(tmpfsDir, fileName)

	// The secret to store
	secretData := "API_KEY=super-secret-value\nDB_PASSWORD=very-secret-password"

	// Create and write the file
	err := os.WriteFile(fullPath, []byte(secretData), 0600)
	if err != nil {
		fmt.Printf("Failed to write secret file: %v\n", err)
		return
	}

	fmt.Printf("Secret file written to: %s\n", fullPath)
	fmt.Println("To copy into your Docker container, run the following command:")
	fmt.Printf("  docker cp %s <container_id>:/path/in/container/%s\n", fullPath, fileName)
}
*/
