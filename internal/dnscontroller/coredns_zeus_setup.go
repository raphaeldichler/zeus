// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package dnscontroller

import (
	"context"
	"net"
	"strings"
	sync "sync"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/miekg/dns"
)

const (
	pluginName = "zeus"
	timeToLive = 600
)

var corednsLog = clog.NewWithPlugin(pluginName)

func init() {
	plugin.Register(pluginName, setup)
}

func setup(c *caddy.Controller) error {
	c.Next()

	p := &ZeusDns{}
	server, err := New(p)
	if err != nil {
		return err
	}

	c.OnStartup(func() error {
		go func() {
			if err := server.Run(); err != nil {
				corednsLog.Fatal(err)
			}
		}()

		return nil
	})

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		p.next = next
		return p
	})

	return nil
}

type ZeusDns struct {
	next  plugin.Handler
	mu    sync.Mutex
	ipMap map[string]string
}

func (z *ZeusDns) setIpMap(m map[string]string) {
	z.mu.Lock()
	defer z.mu.Unlock()
	z.ipMap = m
}

func (z *ZeusDns) Name() string { return pluginName }

func (z *ZeusDns) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	z.mu.Lock()
	defer z.mu.Unlock()

	q := r.Question[0]
	name := strings.TrimSuffix(q.Name, ".")
	if q.Qtype == dns.TypeA {
		if ip, ok := z.ipMap[name]; ok {
			msg := new(dns.Msg)
			msg.SetReply(r)
			msg.Authoritative = true

			a := &dns.A{
				Hdr: dns.RR_Header{
					Name:   q.Name,
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    timeToLive,
				},
				A: net.ParseIP(ip),
			}
			msg.Answer = append(msg.Answer, a)
			w.WriteMsg(msg)
			return dns.RcodeSuccess, nil
		}
	}

	return plugin.NextOrFailure(z.Name(), z.next, ctx, w, r)
}
