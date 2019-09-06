package secret

import (
	"context"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
)

func (ess EmptySecretService) LoadSecret(ctx context.Context, k string) (string, error) {
	return "", errors.Newf(codes.NotFound, "secret key %q not found", k)
}

type EmptySecretService struct {
}
