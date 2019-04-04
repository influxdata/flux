package lang

import (
	"sync"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/memory"
)

// query implements the flux.Query interface.
type query struct {
	results chan flux.Result
	stats   flux.Statistics
	alloc   *memory.Allocator
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
