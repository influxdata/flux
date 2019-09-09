package dependencies

import (
	"net"
	"net/http"
	"time"

	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/dependencies/secret"
	"github.com/influxdata/flux/dependencies/url"
	"github.com/influxdata/flux/internal/errors"
)

const InterpreterDepsKey = "interpreter"

type Interface interface {
	HTTPClient() (*http.Client, error)
	SecretService() (secret.Service, error)
	URLValidator() (url.Validator, error)
}

// Dependencies implements the Interface.
// Any deps which are nil will produce an explicit error.
type Dependencies struct {
	Deps Deps
}

type Deps struct {
	HTTPClient    *http.Client
	SecretService secret.Service
	URLValidator  url.Validator
}

func (d Dependencies) HTTPClient() (*http.Client, error) {
	if d.Deps.HTTPClient != nil {
		return d.Deps.HTTPClient, nil
	}
	return nil, errors.New(codes.Unimplemented, "http client uninitialized in dependencies")
}

func (d Dependencies) SecretService() (secret.Service, error) {
	if d.Deps.SecretService != nil {
		return d.Deps.SecretService, nil
	}
	return nil, errors.New(codes.Unimplemented, "secret service uninitialized in dependencies")
}

func (d Dependencies) URLValidator() (url.Validator, error) {
	if d.Deps.URLValidator != nil {
		return d.Deps.URLValidator, nil
	}
	return nil, errors.New(codes.Unimplemented, "url validator uninitialized in dependencies")
}

// newDefaultTransport creates a new transport with sane defaults.
func newDefaultTransport() *http.Transport {
	// These defaults are copied from http.DefaultTransport.
	return &http.Transport{
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
	}
}

// NewDefaults produces a set of dependencies.
// Not all dependencies have valid defaults and will not be set.
func NewDefaults() Dependencies {
	return Dependencies{
		Deps: Deps{
			HTTPClient:    &http.Client{Transport: newDefaultTransport()},
			SecretService: nil,
			URLValidator:  url.PassValidator{},
		},
	}
}

// NewEmpty produces an empty set of dependencies.
// Accessing any dependency will result in an error.
func NewEmpty() Interface {
	return Dependencies{}
}
