package main

import (
	"context"
	"fmt"

	"github.com/raphaeldichler/zeus/internal/dnscontroller"
)

func main() {
	client := dnscontroller.NewClient()
	defer client.Close()

	ctx := context.Background()
	r, err := client.SetDNSEntry(ctx, &dnscontroller.DNSSetRequest{
		NetworkHash: "1234",
		Entries: []*dnscontroller.DNSSetEntryRequest{
			{
				Domain: "foo.com",
				Type:   dnscontroller.DNSEntryType_Internal,
			},
			{
				Domain: "bra.com",
				Type:   dnscontroller.DNSEntryType_Internal,
			},
			{
				Domain: "got.com",
				Type:   dnscontroller.DNSEntryType_External,
			},
		},
	})

	fmt.Println(err)
  for _, e := range r.DNSEntries {
	  fmt.Println(e.Domain)
	  fmt.Println(e.IP)
    fmt.Println("---")
  }

}
