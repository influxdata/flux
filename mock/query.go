package mock

import (
	"fmt"
	"sync"

	"github.com/influxdata/flux"
)

// Query provides a customizable query that implements flux.Query.
// Results, as well as errors, statistics, and the cancel function can be set.
type Query struct {
	ResultsCh chan flux.Result
	CancelFn  func()
	err       error
	stats     flux.Statistics

	resSet        bool
	providedStats flux.Statistics

	canceledOnce sync.Once
	Canceled     chan struct{}
}

func (q *Query) Results() <-chan flux.Result {
	return q.ResultsCh
}

func (q *Query) Done() {
	// make stats available
	q.stats = q.providedStats
	q.Cancel()
}

func (q *Query) Cancel() {
	if q.CancelFn != nil {
		q.CancelFn()
	}
	if q.Canceled != nil {
		q.canceledOnce.Do(func() {
			close(q.Canceled)
		})
	}
}

func (q *Query) Err() error {
	return q.err
}

func (q *Query) Statistics() flux.Statistics {
	return q.stats
}

// ProduceResults lets the user provide a function to produce results on the channel returned by `Results`.
// `resultProvider` should check if `canceled` has been closed before sending results. E.g.:
// ```
//	 func (results chan<- flux.Result, canceled <-chan struct{}) {
//		 for _, r := range resultsSlice {
//			 select {
//			 case <-canceled:
//			 	 return
//			 default:
//				 results <- r
//			 }
//		 }
//	 }
// ```
// `resultProvider` is run in a separate goroutine and Results() is closed after function completion.
// ProduceResults can be called only once per Query.
func (q *Query) ProduceResults(resultProvider func(results chan<- flux.Result, canceled <-chan struct{})) {
	if q.resSet {
		panic(fmt.Errorf("cannot set results twice, create a new query instead"))
	}

	q.resSet = true
	q.ResultsCh = make(chan flux.Result)
	q.Canceled = make(chan struct{})
	go func() {
		defer close(q.ResultsCh)
		resultProvider(q.ResultsCh, q.Canceled)
	}()
}

// SetErr sets the error for this query and `Cancel`s it
func (q *Query) SetErr(err error) {
	q.err = err
	q.Cancel()
}

// SetStatistics sets stats for this query. Stats will be available after `Done` is called.
func (q *Query) SetStatistics(stats flux.Statistics) {
	q.providedStats = stats
}
