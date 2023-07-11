package bilibili

import (
	"context"
	"github.com/keuin/slbr/types"
	"net"
)

type netContext = func(context.Context, string, string) (net.Conn, error)

type netProbe struct {
	list []types.IpNetType
	i    int
}

func newNetProbe(protocols []types.IpNetType) netProbe {
	var netList []types.IpNetType
	netList = append(netList, protocols...)
	return netProbe{
		list: netList,
		i:    0,
	}
}

func (p *netProbe) NextNetworkType(dialer net.Dialer) (netContext, types.IpNetType) {
	if p.i >= len(p.list) {
		return nil, types.IP64
	}
	network := p.list[p.i]
	p.i++
	return func(ctx context.Context, _, addr string) (net.Conn, error) {
		return dialer.DialContext(ctx, network.GetDialNetString(), addr)
	}, network
}
