package feature

import (
	"context"
)

type key int

const flaggerKey key = iota

// Flagger reports flag values.
type Flagger interface {
	// FlagValue returns the value for a given flag.
	FlagValue(ctx context.Context, flag Flag) interface{}
}

// MutableFlagger reports flag values and allows for values to be modified.
type MutableFlagger interface {
	Flagger
	// SetFlagValue overrides any value for a flag and persists the new value.
	SetFlagValue(context.Context, Flag, interface{})
}

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

// mutableFlagger implements Flagger from a Flagger.
type mutableFlagger struct {
	Flagger
	overrides map[string]interface{}
}

func (f mutableFlagger) FlagValue(ctx context.Context, flag Flag) interface{} {
	if v, ok := f.overrides[flag.Key()]; ok {
		return v
	}
	return f.Flagger.FlagValue(ctx, flag)
}

func (f mutableFlagger) SetFlagValue(ctx context.Context, flag Flag, value interface{}) {
	f.overrides[flag.Key()] = value
}

func NewMutableFlagger(flagger Flagger) MutableFlagger {
	return mutableFlagger{
		Flagger:   flagger,
		overrides: make(map[string]interface{}),
	}
}
