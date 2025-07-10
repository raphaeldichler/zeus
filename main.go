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
			&dnscontroller.DNSSetEntryRequest{
				Domain: "foo.com",
				Type:   dnscontroller.DNSEntryType_Internal,
			},
			&dnscontroller.DNSSetEntryRequest{
				Domain: "bee.love.com",
				Type:   dnscontroller.DNSEntryType_Internal,
			},
		},
	})

	fmt.Println(r)
	fmt.Println(err)
}
