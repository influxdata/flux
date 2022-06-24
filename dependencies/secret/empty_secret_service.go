package secret

import (
	"context"
	"github.com/mvn-trinhnguyen2-dn/flux/codes"
	"github.com/mvn-trinhnguyen2-dn/flux/internal/errors"
)

func (ess EmptySecretService) LoadSecret(ctx context.Context, k string) (string, error) {
	return "", errors.Newf(codes.NotFound, "secret key %q not found", k)
}

// Secret service that always reports no secrets exist
type EmptySecretService struct {
}
