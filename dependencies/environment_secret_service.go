package dependencies

import (
	"context"
	"os"
)

func (ess EnvironmentSecretService) LoadSecret(ctx context.Context, k string) (string, error) {
	return os.Getenv(k), nil
}

type EnvironmentSecretService struct {
}
