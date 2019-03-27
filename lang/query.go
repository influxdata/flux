package lang

import (
	"github.com/influxdata/flux"
)

// Query implements the flux.Query interface.
type Query struct {
	ch   chan flux.Result
}

func (q *Query) Results() <-chan flux.Result {
	return q.ch
}

func (q *Query) Done() {
	// consume all remaining elements so channel can be closed
	for ok := true; ok == true; _, ok = <-q.ch {}
}

func (*Query) Cancel() {
}

func (*Query) Err() error {
	return nil
}

func (*Query) Statistics() flux.Statistics {
	return flux.Statistics{}
}
