package http

import (
	"net"
	"net/http"
	"time"
)

type Client interface {
	Do(*http.Request) (*http.Response, error)
}

// NewDefaultTransport creates a new transport with sane defaults.
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
