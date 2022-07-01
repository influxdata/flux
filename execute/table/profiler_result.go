package table

import (
	"github.com/influxdata/flux"
)

type ProfilerResult struct {
	tables Iterator
}

func NewProfilerResult(tables ...flux.Table) ProfilerResult {
	return ProfilerResult{tables}
}

func (r *ProfilerResult) Name() string {
	return "_profiler"
}

func (r *ProfilerResult) Tables() flux.TableIterator {
	return r.tables
}
