package execute

import (
	"context"
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
		createOperatorProfiler,
	)
}

type OperatorProfilingResult struct {
	Type     string
	Label    string
	Duration time.Duration
	HitCount int64
}

func (r *OperatorProfilingResult) Combine(o *OperatorProfilingResult) error {
	if r.Type != o.Type {
		return errors.Newf(codes.Internal, "Cannot combine a OperatorProfilingResult for type %s with another result for type %s", r.Type, o.Type)
	}
	if r.Label != o.Label {
		return errors.Newf(codes.Internal, "Cannot combine a OperatorProfilingResult for %s with another result for %s", r.Label, o.Label)
	}
	r.Duration = time.Duration(r.Duration.Nanoseconds() + o.Duration.Nanoseconds())
	r.HitCount += o.HitCount
	return nil
}

type OperatorProfilingSpan struct {
	opentracing.Span
	profiler *OperatorProfiler
	Type     string
	Label    string
	start    time.Time
	Duration time.Duration
}

func (t *OperatorProfilingSpan) finish() time.Time {
	finishTime := time.Now()
	t.Duration = finishTime.Sub(t.start)
	if t.profiler != nil && t.profiler.ch != nil {
		t.profiler.ch <- OperatorProfilingResult{
			Type:     t.Type,
			Label:    t.Label,
			Duration: t.Duration,
			HitCount: 1,
		}
	}
	return finishTime
}

func (t *OperatorProfilingSpan) Finish() {
	finishTime := t.finish()
	if t.Span != nil {
		t.Span.FinishWithOptions(opentracing.FinishOptions{
			FinishTime: finishTime,
		})
	}
}

func (t *OperatorProfilingSpan) FinishWithOptions(opts opentracing.FinishOptions) {
	finishTime := t.finish()
	opts.FinishTime = finishTime
	if t.Span != nil {
		t.Span.FinishWithOptions(opts)
	}
}

const OperatorProfilerContextKey = "operator-profiler"

type OperatorProfiler struct {
	// Result aggregated by the transformation/data source name.
	// Those names are actually their operation name. See flux/internal/spec.buildSpec.
	// Some examples are:
	// merged_fromRemote_range1_filter2_filter3_filter4, window5, window8, generated_yield, etc.
	results map[string]OperatorProfilingResult
	// Receive the profiling results from the spans.
	ch chan OperatorProfilingResult
}

func createOperatorProfiler() Profiler {
	p := &OperatorProfiler{
		results: make(map[string]OperatorProfilingResult),
		ch:      make(chan OperatorProfilingResult),
	}
	go func(p *OperatorProfiler) {
		for result := range p.ch {
			if existingResult, exists := p.results[result.Label]; exists {
				// Aggregate the results by name
				existingResult.Combine(&result)
			} else {
				p.results[result.Label] = result
			}
		}
	}(p)
	return p
}

func (t OperatorProfiler) Name() string {
	return "operator"
}

func (t OperatorProfiler) GetResult(q flux.Query, alloc *memory.Allocator) (flux.Table, error) {
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
			values.NewString("profiler/operator"),
		},
	)
	b := NewColListTableBuilder(groupKey, alloc)
	colMeta := []flux.ColMeta{
		{
			Label: "_measurement",
			Type:  flux.TString,
		},
		{
			Label: "Type",
			Type:  flux.TString,
		},
		{
			Label: "Label",
			Type:  flux.TString,
		},
		{
			Label: "Duration",
			Type:  flux.TInt,
		},
		{
			Label: "HitCount",
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
			b.AppendString(0, "profiler/operator")
			b.AppendString(1, result.Type)
			b.AppendString(2, result.Label)
			b.AppendInt(3, result.Duration.Nanoseconds())
			b.AppendInt(4, result.HitCount)
		}
	}
	tbl, err := b.Table()
	if err != nil {
		return nil, err
	}
	return tbl, nil
}

// Create a tracing span.
// Depending on whether the Jaeger tracing and/or the operator profiling are enabled,
// the Span produced by this function can be very different.
// It could be a no-op span, a Jaeger span, a no-op span wrapped by a profiling span, or
// a Jaeger span wrapped by a profiling span.
func StartSpanFromContext(ctx context.Context, operationName string, label string) opentracing.Span {
	var span opentracing.Span
	start := time.Now()
	if flux.IsQueryTracingEnabled(ctx) {
		span, _ = opentracing.StartSpanFromContext(ctx, operationName, opentracing.StartTime(start))
	}
	if tfp, ok := ctx.Value(OperatorProfilerContextKey).(*OperatorProfiler); ok {
		span = &OperatorProfilingSpan{
			Span:     span,
			profiler: tfp,
			Type:     operationName,
			Label:    label,
			start:    start,
			Duration: 0,
		}
	}
	return span
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
