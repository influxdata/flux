package execute

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/flux"
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
	Type string
	// Those labels are actually their operation name. See flux/internal/spec.buildSpec.
	// Some examples are:
	// merged_fromRemote_range1_filter2_filter3_filter4, window5, window8, generated_yield, etc.
	Label string
	Start time.Time
	Stop  time.Time
}

type OperatorProfilingSpan struct {
	opentracing.Span
	profiler *OperatorProfiler
	Result   OperatorProfilingResult
}

func (t *OperatorProfilingSpan) finish() time.Time {
	t.Result.Stop = time.Now()
	if t.profiler != nil && t.profiler.ch != nil {
		t.profiler.ch <- t.Result
	}
	return t.Result.Stop
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
	results []OperatorProfilingResult
	// Receive the profiling results from the spans.
	ch chan OperatorProfilingResult
	mu sync.Mutex
}

func createOperatorProfiler() Profiler {
	p := &OperatorProfiler{
		results: make([]OperatorProfilingResult, 0),
		ch:      make(chan OperatorProfilingResult),
	}
	go func(p *OperatorProfiler) {
		for result := range p.ch {
			p.mu.Lock()
			p.results = append(p.results, result)
			p.mu.Unlock()
		}
	}(p)
	return p
}

func (o *OperatorProfiler) Name() string {
	return "operator"
}

func (o *OperatorProfiler) GetResult(q flux.Query, alloc *memory.Allocator) (flux.Table, error) {
	if o.ch != nil {
		close(o.ch)
		o.ch = nil
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
			Label: "Begin",
			Type:  flux.TTime,
		},
		{
			Label: "End",
			Type:  flux.TTime,
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
	}
	for _, col := range colMeta {
		if _, err := b.AddCol(col); err != nil {
			return nil, err
		}
	}
	o.mu.Lock()
	if o.results != nil && len(o.results) > 0 {
		for _, result := range o.results {
			b.AppendString(0, "profiler/operator")
			b.AppendTime(1, values.Time(result.Start.UnixNano()))
			b.AppendTime(2, values.Time(result.Stop.UnixNano()))
			b.AppendString(3, result.Type)
			b.AppendString(4, result.Label)
			b.AppendInt(5, result.Stop.Sub(result.Start).Nanoseconds())
		}
	}
	o.mu.Unlock()
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
func StartSpanFromContext(ctx context.Context, operationName string, label string) (context.Context, opentracing.Span) {
	var span opentracing.Span
	start := time.Now()
	if flux.IsQueryTracingEnabled(ctx) {
		span, ctx = opentracing.StartSpanFromContext(ctx, operationName, opentracing.StartTime(start))
	}
	if tfp, ok := ctx.Value(OperatorProfilerContextKey).(*OperatorProfiler); ok {
		span = &OperatorProfilingSpan{
			Span:     span,
			profiler: tfp,
			Result: OperatorProfilingResult{
				Type:  operationName,
				Label: label,
				Start: start,
			},
		}
	}
	return ctx, span
}

type QueryProfiler struct{}

func createQueryProfiler() Profiler {
	return &QueryProfiler{}
}

func (s *QueryProfiler) Name() string {
	return "query"
}

func (s *QueryProfiler) GetResult(q flux.Query, alloc *memory.Allocator) (flux.Table, error) {
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
