package main

import (
	"log"
	"net"
	"os"

	"github.com/docker/docker/client"
	"github.com/labstack/echo/v4"
	"github.com/raphaeldichler/zeus/internal/deploy"
	"github.com/sirupsen/logrus"
)

const (
  SocketPath = "zeus-proxy.sock"
)

func ensureSocket() {
  if err := os.RemoveAll(SocketPath); err != nil {
		log.Fatal(err)
	}

}

func main() {
  ensureSocket()

  logrus.SetFormatter(&logrus.TextFormatter{
    FullTimestamp:   true,
    TimestampFormat: "2006-01-02 15:04:05",
  })

	app := echo.New()
  app.GET("/", func(c echo.Context) error {
		c.Logger().Info("handling root endpoint")
		c.Logger().Info("doing something else in the same request")
		return c.String(200, "Hello from socket!")
	})
  
  listener, err := net.Listen("unix", SocketPath)
  if err != nil {
    log.Fatal(err)
  }
  app.Listener = listener

  /*
  server := new(http.Server)
  if err := app.StartServer(server); err != nil {
    log.Fatal(err)
  }
  */

  cli, err := client.NewClientWithOpts(
    client.WithAPIVersionNegotiation(),
  )
  if err != nil {
    logrus.Fatal(err)
  }


  deploy.Deploy(nil, cli)
}
