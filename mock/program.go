package mock

import (
	"context"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/memory"
)

// Program is a mock program that can be returned by the mock compiler.
// It will construct a mock query that will then be passed to ExecuteFn.
type Program struct {
	StartFn   func(ec *flux.ExecutionContext) (*Query, error)
	ExecuteFn func(ctx context.Context, q *Query, alloc *memory.Allocator)
}

func (p *Program) Start(ec *flux.ExecutionContext) (flux.Query, error) {
	startFn := p.StartFn
	if startFn == nil {
		var cancel func()
		ctx, cancel := context.WithCancel(ec.Context)
		startFn = func(ec *flux.ExecutionContext) (*Query, error) {
			results := make(chan flux.Result)
			q := &Query{
				ResultsCh: results,
				CancelFn:  cancel,
				Canceled:  make(chan struct{}),
			}
			go func() {
				defer close(results)
				if p.ExecuteFn != nil {
					p.ExecuteFn(ctx, q, ec.Allocator)
				}
			}()
			return q, nil
		}
	}
	return startFn(ec)
}
