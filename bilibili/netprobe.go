package bilibili

import (
	"context"
	"net"
)

type IpNetType string

var (
	IPv6Net IpNetType = "tcp6"
	IPv4Net IpNetType = "tcp4"
	IP64    IpNetType = "tcp"
)

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
		return dialer.DialContext(ctx, string(network), addr)
	}, network
}
