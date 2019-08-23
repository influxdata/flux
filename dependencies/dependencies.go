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
	SecretService() (SecretService, error)
}

type DefaultDependencies struct {
	httpclient    *http.Client
	secretservice SecretService
}

func (d DefaultDependencies) HTTPClient() (*http.Client, error) {
	if d.httpclient != nil {
		return d.httpclient, nil
	}
	return nil, errors.New(codes.Unimplemented, "http client uninitialized in dependencies")
}

func (d DefaultDependencies) SecretService() (SecretService, error) {
	if d.secretservice != nil {
		return d.secretservice, nil
	}
	return nil, errors.New(codes.Unimplemented, "secret service uninitialized in dependencies")
}

// The values defined below come from the default implementation of Transport.
// It establishes network connections as needed and caches them for reuse by subsequent calls.
// It uses HTTP proxies as directed by the $HTTP_PROXY and $NO_PROXY (or $http_proxy and $no_proxy) environment variables.
func NewDefaultDependencies() Interface {
	return &DefaultDependencies{
		httpclient:    nil,
		secretservice: nil,
	}
}

func NewCLIDependencies() Interface {
	return &DefaultDependencies{
		httpclient: &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
				MaxIdleConnsPerHost:   runtime.GOMAXPROCS(0) + 1,
			},
		},
		secretservice: EnvironmentSecretService{},
	}
}
