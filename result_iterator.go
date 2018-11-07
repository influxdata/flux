package flux

import (
	"context"
	"sort"
)

// ResultIterator allows iterating through all results synchronously.
// A ResultIterator is not thread-safe and all of the methods are expected to be
// called within the same goroutine. A ResultIterator may implement Statisticser.
type ResultIterator interface {
	// More indicates if there are more results.
	More() bool

	// Next returns the next result.
	// If More is false, Next panics.
	Next() Result

	// Release discards the remaining results and frees the currently used resources.
	// It must always be called to free resources. It can be called even if there are
	// more results. It is safe to call Release multiple times.
	Release()

	// Err reports the first error encountered.
	// Err will not report anything unless More has returned false,
	// or the query has been cancelled.
	Err() error
}

// queryResultIterator implements a ResultIterator while consuming a Query
type queryResultIterator struct {
	ctx     context.Context
	query   Query
	done    bool
	results ResultIterator
}

func NewResultIteratorFromQuery(ctx context.Context, q Query) ResultIterator {
	return &queryResultIterator{
		ctx:   ctx,
		query: q,
	}
}

func (r *queryResultIterator) More() bool {
	if r.done {
		return false
	}

	if r.results == nil {
		select {
		case <-r.ctx.Done():
			return false
		case results, ok := <-r.query.Ready():
			if !ok {
				return false
			}
			r.results = NewMapResultIterator(results)
		}
	}
	return r.results.More()
}

func (r *queryResultIterator) Next() Result {
	return r.results.Next()
}

func (r *queryResultIterator) Release() {
	r.query.Done()
	r.done = true
	if r.results != nil {
		r.results.Release()
	}
}

func (r *queryResultIterator) Err() error {
	return r.query.Err()
}

func (r *queryResultIterator) Statistics() Statistics {
	return r.query.Statistics()
}

type mapResultIterator struct {
	results map[string]Result
	order   []string
}

func NewMapResultIterator(results map[string]Result) ResultIterator {
	order := make([]string, 0, len(results))
	for k := range results {
		order = append(order, k)
	}
	sort.Strings(order)
	return &mapResultIterator{
		results: results,
		order:   order,
	}
}

func (r *mapResultIterator) More() bool {
	return len(r.order) > 0
}

func (r *mapResultIterator) Next() Result {
	next := r.order[0]
	r.order = r.order[1:]
	return r.results[next]
}

func (r *mapResultIterator) Release() {
	r.results = nil
	r.order = nil
}

func (r *mapResultIterator) Err() error {
	return nil
}

type sliceResultIterator struct {
	results []Result
}

func NewSliceResultIterator(results []Result) ResultIterator {
	return &sliceResultIterator{
		results: results,
	}
}

func (r *sliceResultIterator) More() bool {
	return len(r.results) > 0
}

func (r *sliceResultIterator) Next() Result {
	next := r.results[0]
	r.results = r.results[1:]
	return next
}

func (r *sliceResultIterator) Release() {
	r.results = nil
}

func (r *sliceResultIterator) Err() error {
	return nil
}
