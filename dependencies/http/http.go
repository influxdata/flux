package http

import (
	"io"
	"net"
	"net/http"
	"time"
)

// maxResponseBody is the maximum response body we will read before just discarding
// the rest. This allows sockets to be reused.
const maxResponseBody = 100 * 1024 * 1024 // 100 MB

type Client interface {
	Do(*http.Request) (*http.Response, error)
}

func LimitHTTPBody(client http.Client, size int64) *http.Client {
	// The client is already a struct so it was already copied
	// which makes this safe.
	if client.Transport == nil {
		client.Transport = http.DefaultTransport
	}
	client.Transport = roundTripLimiter{RoundTripper: client.Transport, size: size}
	return &client
}

type limitedReadCloser struct {
	io.Reader
	io.Closer
}

func limitReadCloser(rc io.ReadCloser, size int64) limitedReadCloser {
	return limitedReadCloser{
		Reader: io.LimitReader(rc, size),
		Closer: rc,
	}
}

type roundTripLimiter struct {
	http.RoundTripper
	size int64
}

func (l roundTripLimiter) RoundTrip(r *http.Request) (*http.Response, error) {
	response, err := l.RoundTripper.RoundTrip(r)
	if err != nil {
		return nil, err
	}
	response.Body = limitReadCloser(response.Body, l.size)
	return response, nil
}

// NewDefaultClient creates a client with sane defaults.
func NewDefaultClient() *http.Client {
	// These defaults are copied from http.DefaultTransport.
	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				// DualStack is deprecated
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       10 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			// Fields below are NOT part of Go's defaults
			MaxIdleConnsPerHost: 100,
		},
	}
}

// NewLimitedDefaultClient creates a client with a limit on the response body size.
func NewLimitedDefaultClient() *http.Client {
	cli := NewDefaultClient()
	return LimitHTTPBody(*cli, maxResponseBody)
}
