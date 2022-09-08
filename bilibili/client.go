/*
This file defines the Bilibili client struct.
This struct is a facade of all necessary Bilibili HTTP API wrappers.
*/
package bilibili

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
)

const (
	// kUserAgent: the default user-agent header to use when communicating with bilibili.
	kUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.0.0 Safari/537.36"
)

type Bilibili struct {
	userAgent string
	http      *http.Client
	loggerCommon
	ctx      context.Context
	netTypes []IpNetType
}

func NewBilibiliWithContext(ctx context.Context, netTypes []IpNetType) Bilibili {
	logger := loggerCommon{
		debug: log.New(os.Stderr, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile),
		info:  log.New(os.Stderr, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		warn:  log.New(os.Stderr, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile),
		error: log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
	}

	var nets []IpNetType
	nets = append(nets, netTypes...)
	if len(nets) == 0 {
		nets = append(nets, IP64)
	}

	var dialer net.Dialer
	np := newNetProbe(nets)
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DialTLSContext = nil
	transport.DialContext, _ = np.NextNetworkType(dialer)

	return Bilibili{
		loggerCommon: logger,
		userAgent:    kUserAgent,
		http:         http.DefaultClient,
		ctx:          ctx,
		netTypes:     nets,
	}
}

func NewBilibiliWithNetType(netTypes []IpNetType) Bilibili {
	ctx := context.Background()
	return NewBilibiliWithContext(ctx, netTypes)
}

func NewBilibili() Bilibili {
	return NewBilibiliWithNetType(nil)
}
