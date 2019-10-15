package flux

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/dependencies/filesystem"
	"github.com/influxdata/flux/dependencies/secret"
	"github.com/influxdata/flux/dependencies/url"
	"github.com/influxdata/flux/internal/errors"
)

// Dependency is an interface that must be implemented by every injectable dependency.
// On Inject, the dependency is injected into the context and the resulting one is returned.
// Every dependency must provide a function to extract it from the context.
type Dependency interface {
	Inject(ctx context.Context) context.Context
}

type key int

const dependenciesKey key = iota

type Dependencies interface {
	Dependency
	HTTPClient() (*http.Client, error)
	FilesystemService() (filesystem.Service, error)
	SecretService() (secret.Service, error)
	URLValidator() (url.Validator, error)
}

// Deps implements Dependencies.
// Any deps which are nil will produce an explicit error.
type Deps struct {
	Deps WrappedDeps
}

type WrappedDeps struct {
	HTTPClient        *http.Client
	FilesystemService filesystem.Service
	SecretService     secret.Service
	URLValidator      url.Validator
}

func (d Deps) HTTPClient() (*http.Client, error) {
	if d.Deps.HTTPClient != nil {
		return d.Deps.HTTPClient, nil
	}
	return nil, errors.New(codes.Unimplemented, "http client uninitialized in dependencies")
}

func (d Deps) FilesystemService() (filesystem.Service, error) {
	if d.Deps.FilesystemService != nil {
		return d.Deps.FilesystemService, nil
	}
	return nil, errors.New(codes.Unimplemented, "filesystem service uninitialized in dependencies")
}

func (d Deps) SecretService() (secret.Service, error) {
	if d.Deps.SecretService != nil {
		return d.Deps.SecretService, nil
	}
	return nil, errors.New(codes.Unimplemented, "secret service uninitialized in dependencies")
}

func (d Deps) URLValidator() (url.Validator, error) {
	if d.Deps.URLValidator != nil {
		return d.Deps.URLValidator, nil
	}
	return nil, errors.New(codes.Unimplemented, "url validator uninitialized in dependencies")
}

func (d Deps) Inject(ctx context.Context) context.Context {
	return context.WithValue(ctx, dependenciesKey, d)
}

func GetDependencies(ctx context.Context) Dependencies {
	deps := ctx.Value(dependenciesKey)
	if deps == nil {
		return NewEmptyDependencies()
	}
	return deps.(Dependencies)
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

// NewDefaultDependencies produces a set of dependencies.
// Not all dependencies have valid defaults and will not be set.
func NewDefaultDependencies() Deps {
	return Deps{
		Deps: WrappedDeps{
			HTTPClient: &http.Client{Transport: newDefaultTransport()},
			// Default to having no filesystem, no secrets, and no url validation (always pass).
			FilesystemService: nil,
			SecretService:     secret.EmptySecretService{},
			URLValidator:      url.PassValidator{},
		},
	}
}

// NewEmptyDependencies produces an empty set of dependencies.
// Accessing any dependency will result in an error.
func NewEmptyDependencies() Dependencies {
	return Deps{}
}
