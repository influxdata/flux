package execute

import (
	"fmt"
	"strings"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/values"
	"github.com/opentracing/opentracing-go"
)

type Profiler interface {
	Name() string
	GetResult(q flux.Query, alloc *memory.Allocator) (flux.Table, error)
}

type CreateProfilerFunc func() Profiler

var AllProfilers = make(map[string]CreateProfilerFunc)

func RegisterProfilerFactories(cpfs ...CreateProfilerFunc) {
	for _, cpf := range cpfs {
		p := cpf()
		AllProfilers[p.Name()] = cpf
	}
}

func init() {
	RegisterProfilerFactories(
		createQueryProfiler,
		createTransformationProfiler,
	)
}

type TransformationProfilingResult struct {
	Name     string
	Duration time.Duration
	HitCount int64
}

func (r *TransformationProfilingResult) Combine(o *TransformationProfilingResult) error {
	if r.Name != o.Name {
		return errors.Newf(codes.Internal, "Cannot combine a TransformationProfilingResult for %s with another result for %s", r.Name, o.Name)
	}
	r.Duration = time.Duration(r.Duration.Nanoseconds() + o.Duration.Nanoseconds())
	r.HitCount++
	return nil
}

type TransformationProfilingSpan struct {
	opentracing.Span
	profiler *TransformationProfiler
	Name     string
	start    time.Time
	Duration time.Duration
}

func (t *TransformationProfilingSpan) finish() {
	finish := time.Now()
	t.Duration = finish.Sub(t.start)
	if t.profiler != nil && t.profiler.ch != nil {
		t.profiler.ch <- TransformationProfilingResult{
			Name:     t.Name,
			Duration: t.Duration,
		}
	}
}

func (t *TransformationProfilingSpan) Finish() {
	t.finish()
	t.Span.Finish()
}

func (t *TransformationProfilingSpan) FinishWithOptions(opts opentracing.FinishOptions) {
	t.finish()
	t.Span.FinishWithOptions(opts)
}

type TransformationProfiler struct {
	results map[string]TransformationProfilingResult
	ch      chan TransformationProfilingResult
}

func createTransformationProfiler() Profiler {
	p := &TransformationProfiler{
		results: make(map[string]TransformationProfilingResult),
		ch:      make(chan TransformationProfilingResult),
	}
	go func(p *TransformationProfiler) {
		for result := range p.ch {
			if existingResult, exists := p.results[result.Name]; exists {
				existingResult.Combine(&result)
			} else {
				p.results[result.Name] = result
			}
		}
	}(p)
	return p
}

func (t TransformationProfiler) Name() string {
	return "transformation"
}

func (t TransformationProfiler) GetResult(q flux.Query, alloc *memory.Allocator) (flux.Table, error) {
	if t.ch != nil {
		close(t.ch)
		t.ch = nil
	}
	groupKey := NewGroupKey(
		[]flux.ColMeta{
			{
				Label: "_measurement",
				Type:  flux.TString,
			},
		},
		[]values.Value{
			values.NewString("profiler/transformation"),
		},
	)
	b := NewColListTableBuilder(groupKey, alloc)
	colMeta := []flux.ColMeta{
		{
			Label: "_measurement",
			Type:  flux.TString,
		},
		{
			Label: "Name",
			Type:  flux.TString,
		},
		{
			Label: "Duration",
			Type:  flux.TInt,
		},
	}
	for _, col := range colMeta {
		if _, err := b.AddCol(col); err != nil {
			return nil, err
		}
	}
	if t.results != nil && len(t.results) > 0 {
		for _, result := range t.results {
			b.AppendString(0, "profiler/transformation")
			b.AppendString(1, result.Name)
			b.AppendInt(2, result.Duration.Nanoseconds())
		}
	}
	tbl, err := b.Table()
	if err != nil {
		return nil, err
	}
	return tbl, nil
}

type QueryProfiler struct{}

func createQueryProfiler() Profiler {
	return &QueryProfiler{}
}

func (s QueryProfiler) Name() string {
	return "query"
}

func (s QueryProfiler) GetResult(q flux.Query, alloc *memory.Allocator) (flux.Table, error) {
	groupKey := NewGroupKey(
		[]flux.ColMeta{
			{
				Label: "_measurement",
				Type:  flux.TString,
			},
		},
		[]values.Value{
			values.NewString("profiler/query"),
		},
	)
	b := NewColListTableBuilder(groupKey, alloc)
	stats := q.Statistics()
	colMeta := []flux.ColMeta{
		{
			Label: "_measurement",
			Type:  flux.TString,
		},
		{
			Label: "TotalDuration",
			Type:  flux.TInt,
		},
		{
			Label: "CompileDuration",
			Type:  flux.TInt,
		},
		{
			Label: "QueueDuration",
			Type:  flux.TInt,
		},
		{
			Label: "PlanDuration",
			Type:  flux.TInt,
		},
		{
			Label: "RequeueDuration",
			Type:  flux.TInt,
		},
		{
			Label: "ExecuteDuration",
			Type:  flux.TInt,
		},
		{
			Label: "Concurrency",
			Type:  flux.TInt,
		},
		{
			Label: "MaxAllocated",
			Type:  flux.TInt,
		},
		{
			Label: "TotalAllocated",
			Type:  flux.TInt,
		},
		{
			Label: "RuntimeErrors",
			Type:  flux.TString,
		},
	}
	colData := []interface{}{
		"profiler/query",
		stats.TotalDuration.Nanoseconds(),
		stats.CompileDuration.Nanoseconds(),
		stats.QueueDuration.Nanoseconds(),
		stats.PlanDuration.Nanoseconds(),
		stats.RequeueDuration.Nanoseconds(),
		stats.ExecuteDuration.Nanoseconds(),
		int64(stats.Concurrency),
		stats.MaxAllocated,
		stats.TotalAllocated,
		strings.Join(stats.RuntimeErrors, "\n"),
	}
	stats.Metadata.Range(func(key string, value interface{}) bool {
		var ty flux.ColType
		if intValue, ok := value.(int); ok {
			ty = flux.TInt
			colData = append(colData, int64(intValue))
		} else {
			ty = flux.TString
			colData = append(colData, fmt.Sprintf("%v", value))
		}
		colMeta = append(colMeta, flux.ColMeta{
			Label: key,
			Type:  ty,
		})
		return true
	})
	for _, col := range colMeta {
		if _, err := b.AddCol(col); err != nil {
			return nil, err
		}
	}
	for i := 0; i < len(colData); i++ {
		if intValue, ok := colData[i].(int64); ok {
			b.AppendInt(i, intValue)
		} else {
			b.AppendString(i, colData[i].(string))
		}
	}
	tbl, err := b.Table()
	if err != nil {
		return nil, err
	}
	return tbl, nil
}
