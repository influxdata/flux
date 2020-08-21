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
	GetResult(q flux.Query, alloc *memory.Allocator) (flux.Table, error)
}

var AllProfilers map[string]Profiler = make(map[string]Profiler)

func RegisterProfilers(ps ...Profiler) {
	for _, p := range ps {
		AllProfilers[p.Name()] = p
	}
}

type FluxStatisticsProfiler struct{}

func init() {
	RegisterProfilers(FluxStatisticsProfiler{})
}

func (s FluxStatisticsProfiler) Name() string {
	return "FluxStatistics"
}

func (s FluxStatisticsProfiler) GetResult(q flux.Query, alloc *memory.Allocator) (flux.Table, error) {
	groupKey := NewGroupKey(
		[]flux.ColMeta{
			{
				Label: "_measurement",
				Type:  flux.TString,
			},
		},
		[]values.Value{
			values.NewString("profiler/FluxStatistics"),
		},
	)
	b := NewColListTableBuilder(groupKey, alloc)
	stats := q.Statistics()
	colMeta := []flux.ColMeta{
		{
			Label: "_measurement",
			Type: flux.TString,
		},
		{
			Label: "TotalDuration",
			Type: flux.TInt,
		},
		{
			Label: "CompileDuration",
			Type: flux.TInt,
		},
		{
			Label: "QueueDuration",
			Type: flux.TInt,
		},
		{
			Label: "PlanDuration",
			Type: flux.TInt,
		},
		{
			Label: "RequeueDuration",
			Type: flux.TInt,
		},
		{
			Label: "ExecuteDuration",
			Type: flux.TInt,
		},
		{
			Label: "Concurrency",
			Type: flux.TInt,
		},
		{
			Label: "MaxAllocated",
			Type: flux.TInt,
		},
		{
			Label: "TotalAllocated",
			Type: flux.TInt,
		},
		{
			Label: "RuntimeErrors",
			Type: flux.TString,
		},
	}
	colData := []interface{} {
		"profiler/FluxStatistics",
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
			Type: ty,
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
