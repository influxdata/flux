package execute

import (
	"fmt"
	"strings"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/values"
)

type Profiler interface {
	Name() string
	GetResult(q flux.Query, alloc memory.Allocator) (flux.Table, error)
	GetSortedResult(q flux.Query, alloc memory.Allocator, desc bool, sortKeys ...string) (flux.Table, error)
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

type OperatorProfiler struct{}

func createOperatorProfiler() Profiler {
	return &OperatorProfiler{}
}

func (o *OperatorProfiler) Name() string {
	return "operator"
}

func (o *OperatorProfiler) GetResult(q flux.Query, alloc memory.Allocator) (flux.Table, error) {
	stats := q.Statistics()
	b, err := o.getTableBuilder(stats, alloc)
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
func (o *OperatorProfiler) GetSortedResult(q flux.Query, alloc memory.Allocator, desc bool, sortKeys ...string) (flux.Table, error) {
	stats := q.Statistics()
	b, err := o.getTableBuilder(stats, alloc)
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

func (o *OperatorProfiler) getTableBuilder(stats flux.Statistics, alloc memory.Allocator) (*ColListTableBuilder, error) {
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

	for _, profile := range stats.Profiles {
		b.AppendString(0, "profiler/operator")
		b.AppendString(1, profile.NodeType)
		b.AppendString(2, profile.Label)
		b.AppendInt(3, profile.Count)
		b.AppendInt(4, profile.Min)
		b.AppendInt(5, profile.Max)
		b.AppendInt(6, profile.Sum)
		b.AppendFloat(7, profile.Mean)
	}
	return b, nil
}

type QueryProfiler struct{}

func createQueryProfiler() Profiler {
	return &QueryProfiler{}
}

func (s *QueryProfiler) Name() string {
	return "query"
}

func (s *QueryProfiler) GetResult(q flux.Query, alloc memory.Allocator) (flux.Table, error) {
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
func (s *QueryProfiler) GetSortedResult(q flux.Query, alloc memory.Allocator, desc bool, sortKeys ...string) (flux.Table, error) {
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

func (s *QueryProfiler) getTableBuilder(q flux.Query, alloc memory.Allocator) (*ColListTableBuilder, error) {
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
	for key, values := range stats.Metadata {
		var ty flux.ColType
		if intValue, ok := values[0].(int); ok {
			ty = flux.TInt
			colData = append(colData, int64(intValue))
		} else {
			ty = flux.TString
			var data string
			for _, value := range values {
				valueStr := fmt.Sprintf("%v", value)
				data += valueStr + "\n"
			}
			colData = append(colData, data)
		}
		colMeta = append(colMeta, flux.ColMeta{
			Label: key,
			Type:  ty,
		})
	}
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
