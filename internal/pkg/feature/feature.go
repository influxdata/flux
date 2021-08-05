package feature

import (
	"context"
)

type key int

const flaggerKey key = iota

// Flagger returns flag values.
type Flagger interface {
	// FlagValue returns the value for a given flag.
	FlagValue(ctx context.Context, flag Flag) interface{}
}

// Inject will inject the Flagger into the context.
func Inject(ctx context.Context, flagger Flagger) context.Context {
	return context.WithValue(ctx, flaggerKey, flagger)
}

// GetFlagger returns the Flagger for this context or the default flagger.
func GetFlagger(ctx context.Context) Flagger {
	flagger := ctx.Value(flaggerKey)
	if flagger == nil {
		return defaultFlagger{}
	}
	return flagger.(Flagger)
}

// defaultFlagger returns a flagger that always returns default values.
type defaultFlagger struct{}

// FlagValue returns the default value for the flag.
// It never returns an error.
func (defaultFlagger) FlagValue(_ context.Context, flag Flag) interface{} {
	return flag.Default()
}
