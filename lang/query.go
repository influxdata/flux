package lang

import (
	"context"
	"sync"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/dependencies/testing"
	"github.com/influxdata/flux/memory"
	"github.com/opentracing/opentracing-go"
)

// query implements the flux.Query interface.
type query struct {
	ctx     context.Context
	results chan flux.Result
	stats   flux.Statistics
	alloc   *memory.Allocator
	span    opentracing.Span
	cancel  func()
	err     error
	wg      sync.WaitGroup
}

func (q *query) Results() <-chan flux.Result {
	return q.results
}

func (q *query) Done() {
	q.cancel()
	q.wg.Wait()
	q.stats.MaxAllocated = q.alloc.MaxAllocated()
	q.stats.TotalAllocated = q.alloc.TotalAllocated()
	if q.span != nil {
		q.span.Finish()
		q.span = nil
	}

	// Note: it is safe to read and write to q.err because we have explicitly
	// waited on the wait group, therefore only a the current goroutine
	// can access q.err
	if q.err == nil {
		// If the testing framework was configured, verify all expectations.
		q.err = testing.Check(q.ctx)
	}
}

func (q *query) Cancel() {
	q.cancel()
}

func (q *query) Err() error {
	return q.err
}

func (q *query) Statistics() flux.Statistics {
	return q.stats
}

func (q *query) ProfilerResults() (flux.ResultIterator, error) {
	return nil, nil
}
