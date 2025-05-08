package main

import (
	"net"

	"github.com/labstack/echo/v4"
)


func main() {
	app := echo.New()

  listener, err := net.Listen("unix", "/var/run/zeus-proxy.sock")


}
