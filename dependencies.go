package flux

import (
	"context"
	"net"
	"syscall"
	"time"

	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/dependencies/filesystem"
	"github.com/influxdata/flux/dependencies/http"
	"github.com/influxdata/flux/dependencies/secret"
	"github.com/influxdata/flux/dependencies/url"
	"github.com/influxdata/flux/dependency"
	"github.com/influxdata/flux/internal/errors"
)

var _ Dependencies = (*Deps)(nil)

type Dependency = dependency.Interface

type key int

const dependenciesKey key = iota

type Dependencies interface {
	Dependency
	HTTPClient() (http.Client, error)
	PrivateHTTPClient() (http.Client, error)
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
	HTTPClient        http.Client
	FilesystemService filesystem.Service
	SecretService     secret.Service
	URLValidator      url.Validator
}

func (d Deps) HTTPClient() (http.Client, error) {
	if d.Deps.HTTPClient != nil {
		return d.Deps.HTTPClient, nil
	}
	return nil, errors.New(codes.Unimplemented, "http client uninitialized in dependencies")
}

func (d Deps) PrivateHTTPClient() (http.Client, error) {
	c, err := d.HTTPClient()
	if err != nil {
		return nil, err
	}
	return http.NewPrivateClient(c), nil
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
	ctx = context.WithValue(ctx, dependenciesKey, d)
	if d.Deps.FilesystemService != nil {
		ctx = filesystem.Inject(ctx, d.Deps.FilesystemService)
	}
	return ctx
}

func GetDependencies(ctx context.Context) Dependencies {
	deps := ctx.Value(dependenciesKey)
	if deps == nil {
		return NewEmptyDependencies()
	}
	return deps.(Dependencies)
}

// NewDefaultDependencies produces a set of dependencies.
// Not all dependencies have valid defaults and will not be set.
func NewDefaultDependencies() Deps {
	validator := url.PassValidator{}
	return Deps{
		Deps: WrappedDeps{
			HTTPClient: http.NewLimitedDefaultClient(validator),
			// Default to having no filesystem, no secrets, and no url validation (always pass).
			FilesystemService: nil,
			SecretService:     secret.EmptySecretService{},
			URLValidator:      validator,
		},
	}
}

// NewEmptyDependencies produces an empty set of dependencies.
// Accessing any dependency will result in an error.
func NewEmptyDependencies() Deps {
	return Deps{}
}

// GetDialer will return a net.Dialer using the injected dependencies
// within the context.Context.
func GetDialer(ctx context.Context) (*net.Dialer, error) {
	deps := GetDependencies(ctx)
	url, err := deps.URLValidator()
	if err != nil {
		return nil, err
	}

	// Control is called after DNS lookup, but before the
	// network connection is initiated.
	control := func(network, address string, c syscall.RawConn) error {
		host, _, err := net.SplitHostPort(address)
		if err != nil {
			return err
		}

		ip := net.ParseIP(host)
		return url.ValidateIP(ip)
	}

	return &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		Control:   control,
	}, nil
}
