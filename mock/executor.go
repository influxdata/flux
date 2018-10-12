package mock

import (
	"context"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	plan "github.com/influxdata/flux/planner"
)

var _ execute.Executor = (*Executor)(nil)

// Executor is a mock implementation of an execute.Executor.
type Executor struct {
	ExecuteFn func(ctx context.Context, p *plan.PlanSpec, a *execute.Allocator) (map[string]flux.Result, error)
}

// NewExecutor returns a mock Executor where its methods will return zero values.
func NewExecutor() *Executor {
	return &Executor{
		ExecuteFn: func(context.Context, *plan.PlanSpec, *execute.Allocator) (map[string]flux.Result, error) {
			return nil, nil
		},
	}
}

func (e *Executor) Execute(ctx context.Context, p *plan.PlanSpec, a *execute.Allocator) (map[string]flux.Result, error) {
	return e.ExecuteFn(ctx, p, a)
}
