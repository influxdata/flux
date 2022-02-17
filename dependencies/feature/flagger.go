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

type Dependency struct {
	Flagger Flagger
}

func (d Dependency) Inject(ctx context.Context) context.Context {
	return Inject(ctx, d.Flagger)
}

// Flags returns all feature flags.
func Flags() []Flag {
	return feature.Flags()
}

// ByKey returns the Flag corresponding to the given key.
func ByKey(k string) (Flag, bool) {
	return feature.ByKey(k)
}

type Metrics = feature.Metrics

// SetMetrics sets the metric store for feature flags.
func SetMetrics(m Metrics) {
	feature.SetMetrics(m)
}
