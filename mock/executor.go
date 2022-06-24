package mock

import (
	"context"

	"github.com/mvn-trinhnguyen2-dn/flux"
	"github.com/mvn-trinhnguyen2-dn/flux/execute"
	"github.com/mvn-trinhnguyen2-dn/flux/memory"
	"github.com/mvn-trinhnguyen2-dn/flux/metadata"
	"github.com/mvn-trinhnguyen2-dn/flux/plan"
)

var _ execute.Executor = (*Executor)(nil)

var NoMetadata <-chan metadata.Metadata

// Executor is a mock implementation of an execute.Executor.
type Executor struct {
	ExecuteFn func(ctx context.Context, p *plan.Spec, a memory.Allocator) (map[string]flux.Result, <-chan metadata.Metadata, error)
}

// NewExecutor returns a mock Executor where its methods will return zero values.
func NewExecutor() *Executor {
	return &Executor{
		ExecuteFn: func(context.Context, *plan.Spec, memory.Allocator) (map[string]flux.Result, <-chan metadata.Metadata, error) {
			return nil, NoMetadata, nil
		},
	}
}

func (e *Executor) Execute(ctx context.Context, p *plan.Spec, a memory.Allocator) (map[string]flux.Result, <-chan metadata.Metadata, error) {
	return e.ExecuteFn(ctx, p, a)
}

func init() {
	noMetaCh := make(chan metadata.Metadata)
	close(noMetaCh)
	NoMetadata = noMetaCh
}
