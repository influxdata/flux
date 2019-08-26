package dependencies

import (
	"net/http"

	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
)

const InterpreterDepsKey = "interpreter"

type Interface interface {
	HTTPClient() (*http.Client, error)
	SecretService() (SecretService, error)
}

// Dependencies implemnents the Interface.
// Any deps which are nil will produce an explicit error.
type Dependencies struct {
	Deps Deps
}

type Deps struct {
	HTTPClient    *http.Client
	SecretService SecretService
}

func (d Dependencies) HTTPClient() (*http.Client, error) {
	if d.Deps.HTTPClient != nil {
		return d.Deps.HTTPClient, nil
	}
	return nil, errors.New(codes.Unimplemented, "http client uninitialized in dependencies")
}

func (d Dependencies) SecretService() (SecretService, error) {
	if d.Deps.SecretService != nil {
		return d.Deps.SecretService, nil
	}
	return nil, errors.New(codes.Unimplemented, "secret service uninitialized in dependencies")
}

// NewDefaults produces a set of dependencies.
// Not all dependencies have valid defaults and will not be set.
func NewDefaults() Dependencies {
	return Dependencies{
		Deps: Deps{
			HTTPClient:    http.DefaultClient,
			SecretService: nil,
		},
	}
}

// NewEmpty produces an empty set of dependencies.
// Accessing any dependency will result in an error.
func NewEmpty() Interface {
	return Dependencies{}
}
