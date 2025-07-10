package foo

import (
	"fmt"
	"net"
	"strings"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/miekg/dns"
	"golang.org/x/net/context"
)

func init() {
	plugin.Register("foo", setup)
}

func setup(c *caddy.Controller) error {
	c.Next()          // #1
	if !c.NextArg() { // #2
		return c.ArgErr()
	}

  c.OnStartup(func() error {
    log.Info("OnStartup")
    return nil
  })

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
    m := make(map[string]string)
    m["foo.com"] = "1.1.1.1"
		return Foo{Next: next, Bar: c.Val(), IPMap: m} 
	})

	return nil // #4
}

var log = clog.NewWithPlugin("foo")

type Foo struct {
	Bar  string
	Next plugin.Handler
  IPMap   map[string]string // map[domain] = ip
}

func (h Foo) Name() string { return "foo" }

func (m Foo) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
  fmt.Println("foo")
	q := r.Question[0]
	name := strings.TrimSuffix(q.Name, ".")

	if q.Qtype == dns.TypeA {
		if ip, ok := m.IPMap[name]; ok {
			msg := new(dns.Msg)
			msg.SetReply(r)
			msg.Authoritative = true

			a := &dns.A{
				Hdr: dns.RR_Header{
					Name:   q.Name,
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    300,
				},
				A: net.ParseIP(ip),
			}
			msg.Answer = append(msg.Answer, a)
			w.WriteMsg(msg)
			return dns.RcodeSuccess, nil
		}
	}

	// Not in map? Pass to the next plugin (likely 'forward')
	return plugin.NextOrFailure(m.Name(), m.Next, ctx, w, r)
}
