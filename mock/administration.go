package mock

import (
	"context"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/memory"
)

type Administration struct{}

func (a *Administration) Context() context.Context {
	return nil
}

func (a *Administration) ResolveTime(qt flux.Time) execute.Time {
	return execute.Now()
}

func (a *Administration) StreamContext() execute.StreamContext {
	return nil
}

func (a *Administration) Allocator() *memory.Allocator {
	return &memory.Allocator{}
}

func (a *Administration) Parents() []execute.DatasetID {
	return nil
}
