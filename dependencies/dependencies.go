package dependencies

import (
	"net"
	"net/http"
	"runtime"
	"time"

	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
)

const InterpreterDepsKey = "interpreter"

type Interface interface {
	HTTPClient() (*http.Client, error)
}

type DefaultDependencies struct {
	httpclient *http.Client
}

func (d DefaultDependencies) HTTPClient() (*http.Client, error) {
	if d.httpclient != nil {
		return d.httpclient, nil
	}
	return nil, errors.New(codes.Invalid, "http client uninitialized in dependencies")
}

// The values defined below come from the default implementation of Transport.
// It establishes network connections as needed and caches them for reuse by subsequent calls.
// It uses HTTP proxies as directed by the $HTTP_PROXY and $NO_PROXY (or $http_proxy and $no_proxy) environment variables.
func NewDefaultDependencies() Interface {
	return &DefaultDependencies{&http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment, // function to return a proxy for a given Request
			DialContext: (&net.Dialer{ // specifies the dial function for creating unencrypted TCP connections.
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:          100,                       // max number of idle connections across all hosts
			IdleConnTimeout:       90 * time.Second,          // max amount of time an idle connection will remain idle before closing itself
			TLSHandshakeTimeout:   10 * time.Second,          // max amount of time waiting to wait for a TLS handshake
			ExpectContinueTimeout: 1 * time.Second,           // amount of  time to wait for a server's first response headers after fully writing the request headers
			MaxIdleConnsPerHost:   runtime.GOMAXPROCS(0) + 1, // max idle connections to keep per-host
		},
	},
	}
}
