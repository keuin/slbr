package bilibili

import (
	"context"
	"fmt"
	"net"
)

type IpNetType string

var (
	IPv6Net IpNetType = "ipv6"
	IPv4Net IpNetType = "ipv4"
	IP64    IpNetType = "any"
)

// GetDialNetString returns the string accepted by net.Dialer::DialContext
func (t IpNetType) GetDialNetString() string {
	switch t {
	case IPv4Net:
		return "tcp4"
	case IPv6Net:
		return "tcp6"
	case IP64:
		return "tcp"
	}
	return ""
}

func (t IpNetType) String() string {
	return fmt.Sprintf("%s(%s)", string(t), t.GetDialNetString())
}

type netContext = func(context.Context, string, string) (net.Conn, error)

type netProbe struct {
	list []IpNetType
	i    int
}

func newNetProbe(protocols []IpNetType) netProbe {
	var netList []IpNetType
	netList = append(netList, protocols...)
	return netProbe{
		list: netList,
		i:    0,
	}
}

func (p *netProbe) NextNetworkType(dialer net.Dialer) (netContext, IpNetType) {
	if p.i >= len(p.list) {
		return nil, IP64
	}
	network := p.list[p.i]
	p.i++
	return func(ctx context.Context, _, addr string) (net.Conn, error) {
		return dialer.DialContext(ctx, network.GetDialNetString(), addr)
	}, network
}
