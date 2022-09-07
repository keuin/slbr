package bilibili

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
)

// newRequest create an HTTP request with per-instance User-Agent set.
func (b Bilibili) newRequest(
	method string,
	url string,
	body io.Reader,
) (req *http.Request, err error) {
	req, err = http.NewRequestWithContext(b.ctx, method, url, body)
	if err != nil {
		b.error.Printf("Cannot create HTTP request instance: %v. Method: %v, URL: %v", err, method, url)
		return
	}
	req.Header.Set("User-Agent", b.userAgent)
	return
}

// newRequest create an HTTP GET request with an empty body and per-instance User-Agent set.
func (b Bilibili) newGet(url string) (req *http.Request, err error) {
	return b.newRequest("GET", url, strings.NewReader(""))
}

// callGet make a GET request and parse response as a JSON document with given model.
func callGet[T BaseResponse[V], V any](b Bilibili, url string) (resp T, err error) {
	logger := log.Default()
	req, err := b.newGet(url)
	if err != nil {
		logger.Printf("ERROR: Cannot create HTTP request instance on API %v: %v", url, err)
		return
	}

	r, err := b.http.Do(req)
	if err != nil {
		logger.Printf("ERROR: HTTP Request failed on API %v: %v", url, err)
		return
	}
	defer func() { _ = r.Body.Close() }()

	err = validateHttpStatus(r)
	if err != nil {
		b.error.Printf("%v", err)
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		b.error.Printf("Error when reading HTTP response on API %v: %v", url, err)
		return
	}

	err = json.Unmarshal(data, &resp)
	if err != nil {
		b.error.Printf("Invalid JSON body of HTTP response on API %v: %v. Text: \"%v\"",
			url, err, string(data))
		return
	}

	b.debug.Printf("HTTP %v, len: %v bytes, url: %v", r.StatusCode, len(data), url)
	return
}
