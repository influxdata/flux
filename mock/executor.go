package mock

import (
	"context"

	"github.com/InfluxCommunity/flux"
	"github.com/InfluxCommunity/flux/execute"
	"github.com/InfluxCommunity/flux/memory"
	"github.com/InfluxCommunity/flux/plan"
)

var _ execute.Executor = (*Executor)(nil)

var EmptyStatistics <-chan flux.Statistics

// Executor is a mock implementation of an execute.Executor.
type Executor struct {
	ExecuteFn func(ctx context.Context, p *plan.Spec, a memory.Allocator) (map[string]flux.Result, <-chan flux.Statistics, error)
}

// NewExecutor returns a mock Executor where its methods will return zero values.
func NewExecutor() *Executor {
	return &Executor{
		ExecuteFn: func(context.Context, *plan.Spec, memory.Allocator) (map[string]flux.Result, <-chan flux.Statistics, error) {
			return nil, EmptyStatistics, nil
		},
	}
}

func (e *Executor) Execute(ctx context.Context, p *plan.Spec, a memory.Allocator) (map[string]flux.Result, <-chan flux.Statistics, error) {
	return e.ExecuteFn(ctx, p, a)
}

func init() {
	emptyStatsCh := make(chan flux.Statistics)
	close(emptyStatsCh)
	EmptyStatistics = emptyStatsCh
}
