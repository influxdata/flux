package feature

import (
	"context"

	"github.com/influxdata/flux/internal/feature"
)

type (
	Flagger = feature.Flagger
	Flag    = feature.Flag
)

// Inject injects a Flagger to the context.Context.
func Inject(ctx context.Context, flagger Flagger) context.Context {
	return feature.Inject(ctx, flagger)
}

// Flags returns all feature flags.
func Flags() []Flag {
	return feature.Flags()
}

// ByKey returns the Flag corresponding to the given key.
func ByKey(k string) (Flag, bool) {
	return feature.ByKey(k)
}
