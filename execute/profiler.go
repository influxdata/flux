package execute

import (
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
	for _, colName := range []string{"_measurement", "_field", DefaultValueColLabel} {
		if _, err := b.AddCol(flux.ColMeta{
			Label: colName,
			Type:  flux.TString,
		}); err != nil {
			return nil, err
		}
	}
	q.Statistics().Range(func(key string, value string) {
		b.AppendString(0, "profiler/FluxStatistics")
		b.AppendString(1, key)
		b.AppendString(2, value)
	})
	b.Sort([]string{"_field"}, false)
	tbl, err := b.Table()
	if err != nil {
		return nil, err
	}
	return tbl, nil
}
