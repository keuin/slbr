package bilibili

import (
	"encoding/json"
	"github.com/keuin/slbr/types"
	"io"
	"net"
	"net/http"
	"strings"
)

// newRequest create an HTTP request with per-instance User-Agent set.
func (b *Bilibili) newRequest(
	method string,
	url string,
	body io.Reader,
) (req *http.Request, err error) {
	req, err = http.NewRequestWithContext(b.ctx, method, url, body)
	if err != nil {
		b.logger.Error("Cannot create HTTP request instance: %v. Method: %v, URL: %v", err, method, url)
		return
	}
	req.Header.Set("User-Agent", b.userAgent)
	return
}

// newRequest create an HTTP GET request with an empty body and per-instance User-Agent set.
func (b *Bilibili) newGet(url string) (req *http.Request, err error) {
	return b.newRequest("GET", url, strings.NewReader(""))
}

// callGetRaw make a GET request and returns the raw response body.
func callGetRaw(b *Bilibili, url string) (resp *http.Response, respBody []byte, err error) {
	req, err := b.newGet(url)
	if err != nil {
		b.logger.Error("Cannot create HTTP request instance on API %v: %v", url, err)
		return
	}

	r, err := b.Do(req)
	if err != nil {
		b.logger.Error("HTTP Request failed on API %v: %v", url, err)
		return
	}
	defer func() { _ = r.Body.Close() }()

	err = validateHttpStatus(r)
	if err != nil {
		b.logger.Error("%v", err)
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		b.logger.Error("Error when reading HTTP response on API %v: %v", url, err)
		return
	}

	return r, data, err
}

// callGet make a GET request and parse response as a JSON document with given model.
func callGet[T types.BaseResponse[V], V any](b *Bilibili, url string) (resp T, err error) {
	r, data, err := callGetRaw(b, url)
	if err != nil {
		return
	}

	err = json.Unmarshal(data, &resp)
	if err != nil {
		b.logger.Error("Invalid JSON body of HTTP response on API %v: %v. Text: \"%v\"",
			url, err, string(data))
		return
	}

	b.logger.Debug("HTTP %v, len: %v bytes, url: %v", r.StatusCode, len(data), url)
	return
}

func (b *Bilibili) Do(req *http.Request) (resp *http.Response, err error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DialTLSContext = nil

	np := newNetProbe(b.netTypes)
	var dialer net.Dialer
	for netCtx, typeName := np.NextNetworkType(dialer); netCtx != nil; netCtx, typeName = np.NextNetworkType(dialer) {
		transport.DialContext = netCtx
		b.http.Transport = transport
		resp, err = b.http.Do(req)
		_, isOpErr := err.(*net.OpError)
		_, isAddrErr := err.(*net.AddrError)
		if err == nil || !isOpErr || !isAddrErr {
			// return the first success request
			b.logger.Info("Request success with network %v.", typeName)
			return
		}
	}
	return
}
