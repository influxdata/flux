package mock

import (
	"context"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/memory"
)

// Administration is a mock implementation of the execute.Administration interface.
// This may be used for tests that require implementation of this interface.
type Administration struct {
	ctx context.Context
	mem *memory.Allocator
}

func AdministrationWithContext(ctx context.Context, mem ...*memory.Allocator) *Administration {
	a := &Administration{ctx: ctx}
	if len(mem) > 0 {
		a.mem = mem[0]
	}
	return a
}

func (a *Administration) Context() context.Context {
	return a.ctx
}

func (a *Administration) ResolveTime(qt flux.Time) execute.Time {
	return execute.Now()
}

func (a *Administration) StreamContext() execute.StreamContext {
	return nil
}

func (a *Administration) Allocator() *memory.Allocator {
	if a.mem != nil {
		return a.mem
	}
	return &memory.Allocator{}
}

func (a *Administration) Parents() []execute.DatasetID {
	return nil
}
