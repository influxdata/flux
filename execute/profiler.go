package execute

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/values"
	"github.com/opentracing/opentracing-go"
)

type Profiler interface {
	Name() string
	GetResult(q flux.Query, alloc *memory.Allocator) (flux.Table, error)
	GetSortedResult(q flux.Query, alloc *memory.Allocator, desc bool, sortKeys ...string) (flux.Table, error)
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

func (t *OperatorProfilingSpan) finish(finishTime time.Time) time.Time {
	t.Result.Stop = finishTime
	if t.profiler != nil && t.profiler.chIn != nil {
		t.profiler.chIn <- t.Result
	}
	return t.Result.Stop
}

func (t *OperatorProfilingSpan) Finish() {
	finishTime := t.finish(time.Now())
	if t.Span != nil {
		t.Span.FinishWithOptions(opentracing.FinishOptions{
			FinishTime: finishTime,
		})
	}
}

func (t *OperatorProfilingSpan) FinishWithOptions(opts opentracing.FinishOptions) {
	finishTime := t.finish(opts.FinishTime)
	opts.FinishTime = finishTime
	if t.Span != nil {
		t.Span.FinishWithOptions(opts)
	}
}

const OperatorProfilerContextKey = "operator-profiler"

type operatorProfilingResultAggregate struct {
	operationType string
	label         string
	resultCount   int64
	resultMin     int64
	resultMax     int64
	resultSum     int64
	resultMean    float64
}

type operatorProfilerLabelGroup = map[string]*operatorProfilingResultAggregate
type operatorProfilerTypeGroup = map[string]operatorProfilerLabelGroup

type OperatorProfiler struct {
	// Receive the profiling results from the spans.
	chIn  chan OperatorProfilingResult
	chOut chan operatorProfilingResultAggregate
}

func createOperatorProfiler() Profiler {
	p := &OperatorProfiler{
		chIn:  make(chan OperatorProfilingResult),
		chOut: make(chan operatorProfilingResultAggregate),
	}
	go func(p *OperatorProfiler) {
		aggs := make(operatorProfilerTypeGroup)
		for result := range p.chIn {
			_, ok := aggs[result.Type]
			if !ok {
				aggs[result.Type] = make(operatorProfilerLabelGroup)
			}
			_, ok = aggs[result.Type][result.Label]
			if !ok {
				aggs[result.Type][result.Label] = &operatorProfilingResultAggregate{}
			}
			a := aggs[result.Type][result.Label]

			// Aggregate the results
			a.resultCount++
			duration := result.Stop.Sub(result.Start).Nanoseconds()
			if duration > a.resultMax {
				a.resultMax = duration
			}
			if duration < a.resultMin || a.resultMin == 0 {
				a.resultMin = duration
			}
			a.resultSum += duration
		}

		// Write the aggregated results to chOut, where they'll be
		// converted into rows and appended to the final table
		for typ, labels := range aggs {
			for label, agg := range labels {
				agg.resultMean = float64(agg.resultSum) / float64(agg.resultCount)
				agg.operationType = typ
				agg.label = label
				p.chOut <- *agg
			}
		}
		close(p.chOut)
	}(p)

	return p
}

func (o *OperatorProfiler) Name() string {
	return "operator"
}

func (o *OperatorProfiler) closeIncomingChannel() {
	if o.chIn != nil {
		close(o.chIn)
		o.chIn = nil
	}
}

func (o *OperatorProfiler) GetResult(q flux.Query, alloc *memory.Allocator) (flux.Table, error) {
	o.closeIncomingChannel()
	b, err := o.getTableBuilder(alloc)
	if err != nil {
		return nil, err
	}
	tbl, err := b.Table()
	if err != nil {
		return nil, err
	}
	return tbl, nil
}

// GetSortedResult is identical to GetResult, except it calls Sort()
// on the ColListTableBuilder to make testing easier.
// sortKeys and desc are passed directly into the Sort() call
func (o *OperatorProfiler) GetSortedResult(q flux.Query, alloc *memory.Allocator, desc bool, sortKeys ...string) (flux.Table, error) {
	o.closeIncomingChannel()
	b, err := o.getTableBuilder(alloc)
	if err != nil {
		return nil, err
	}
	b.Sort(sortKeys, desc)
	tbl, err := b.Table()
	if err != nil {
		return nil, err
	}
	return tbl, nil
}

func (o *OperatorProfiler) getTableBuilder(alloc *memory.Allocator) (*ColListTableBuilder, error) {
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
			Label: "Count",
			Type:  flux.TInt,
		},
		{
			Label: "MinDuration",
			Type:  flux.TInt,
		},
		{
			Label: "MaxDuration",
			Type:  flux.TInt,
		},
		{
			Label: "DurationSum",
			Type:  flux.TInt,
		},
		{
			Label: "MeanDuration",
			Type:  flux.TFloat,
		},
	}
	for _, col := range colMeta {
		if _, err := b.AddCol(col); err != nil {
			return nil, err
		}
	}

	for agg := range o.chOut {
		b.AppendString(0, "profiler/operator")
		b.AppendString(1, agg.operationType)
		b.AppendString(2, agg.label)
		b.AppendInt(3, agg.resultCount)
		b.AppendInt(4, agg.resultMin)
		b.AppendInt(5, agg.resultMax)
		b.AppendInt(6, agg.resultSum)
		b.AppendFloat(7, agg.resultMean)
	}
	return b, nil
}

// Create a tracing span.
// Depending on whether the Jaeger tracing and/or the operator profiling are enabled,
// the Span produced by this function can be very different.
// It could be a no-op span, a Jaeger span, a no-op span wrapped by a profiling span, or
// a Jaeger span wrapped by a profiling span.
func StartSpanFromContext(ctx context.Context, operationName string, label string, opts ...opentracing.StartSpanOption) (context.Context, opentracing.Span) {
	var span opentracing.Span
	var start time.Time
	for _, opt := range opts {
		if st, ok := opt.(opentracing.StartTime); ok {
			start = time.Time(st)
			break
		}
	}
	if start.IsZero() {
		start = time.Now()
		opts = append(opts, opentracing.StartTime(start))
	}
	if flux.IsQueryTracingEnabled(ctx) {
		span, ctx = opentracing.StartSpanFromContext(ctx, operationName, opts...)
	}

	if HaveExecutionDependencies(ctx) {
		deps := GetExecutionDependencies(ctx)
		if deps.ExecutionOptions.OperatorProfiler != nil {
			tfp := deps.ExecutionOptions.OperatorProfiler
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
	b, err := s.getTableBuilder(q, alloc)
	if err != nil {
		return nil, err
	}
	tbl, err := b.Table()
	if err != nil {
		return nil, err
	}
	return tbl, nil
}

// GetSortedResult is identical to GetResult, except it calls Sort()
// on the ColListTableBuilder to make testing easier.
// sortKeys and desc are passed directly into the Sort() call
func (s *QueryProfiler) GetSortedResult(q flux.Query, alloc *memory.Allocator, desc bool, sortKeys ...string) (flux.Table, error) {
	b, err := s.getTableBuilder(q, alloc)
	if err != nil {
		return nil, err
	}
	b.Sort(sortKeys, desc)
	tbl, err := b.Table()
	if err != nil {
		return nil, err
	}
	return tbl, nil
}

func (s *QueryProfiler) getTableBuilder(q flux.Query, alloc *memory.Allocator) (*ColListTableBuilder, error) {
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
	return b, nil
}
