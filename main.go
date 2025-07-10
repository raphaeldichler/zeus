package main

import (
	"github.com/coredns/coredns/coremain"
	_ "github.com/raphaeldichler/zeus/internal/dnscontroller"
)

func main() {
	coremain.Run()
}
