/*
This file defines the Bilibili client struct.
This struct is a facade of all necessary Bilibili HTTP API wrappers.
*/
package bilibili

import (
	"context"
	"github.com/keuin/slbr/logging"
	"github.com/keuin/slbr/types"
	"net"
	"net/http"
	"net/http/cookiejar"
)

const (
	// userAgent: the default user-agent header to use when communicating with bilibili.
	userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.0.0 Safari/537.36"
)

type Bilibili struct {
	userAgent string
	http      *http.Client
	ctx       context.Context
	netTypes  []types.IpNetType
	logger    logging.Logger
}

func NewBilibiliWithContext(ctx context.Context, netTypes []types.IpNetType, logger logging.Logger) *Bilibili {
	var nets []types.IpNetType
	nets = append(nets, netTypes...)
	if len(nets) == 0 {
		nets = append(nets, types.IP64)
	}

	var dialer net.Dialer
	np := newNetProbe(nets)
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DialTLSContext = nil
	transport.DialContext, _ = np.NextNetworkType(dialer)

	cookieJar, _ := cookiejar.New(nil)
	httpClient := &http.Client{
		Jar: cookieJar,
	}

	return &Bilibili{
		logger:    logger,
		userAgent: userAgent,
		http:      httpClient,
		ctx:       ctx,
		netTypes:  nets,
	}
}

func NewBilibiliWithNetType(netTypes []types.IpNetType, logger logging.Logger) *Bilibili {
	ctx := context.Background()
	return NewBilibiliWithContext(ctx, netTypes, logger)
}

func NewBilibili(logger logging.Logger) *Bilibili {
	return NewBilibiliWithNetType(nil, logger)
}
