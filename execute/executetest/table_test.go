package executetest

import (
	"context"
	"testing"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/memory"
)

func TestTable(t *testing.T) {
	RunTableTests(t, TableTest{
		NewFn: func(ctx context.Context, alloc *memory.Allocator) flux.TableIterator {
			return &TableIterator{
				Tables: []*Table{
					{
						ColMeta: []flux.ColMeta{
							{
								Label: "_measurement",
								Type:  flux.TString,
							},
							{
								Label: "_field",
								Type:  flux.TString,
							},
							{
								Label: execute.DefaultTimeColLabel,
								Type:  flux.TTime,
							},
							{
								Label: execute.DefaultValueColLabel,
								Type:  flux.TFloat,
							},
						},
						KeyCols: []string{"_measurement", "_field"},
						Data: [][]interface{}{
							{"m0", "f0", execute.Time(0), 2.0},
							{"m0", "f0", execute.Time(10), 2.5},
							{"m0", "f0", execute.Time(20), 3.0},
							{"m0", "f0", execute.Time(30), 3.5},
						},
						Alloc: alloc,
					},
					{
						ColMeta: []flux.ColMeta{
							{
								Label: "_measurement",
								Type:  flux.TString,
							},
							{
								Label: "_field",
								Type:  flux.TString,
							},
							{
								Label: execute.DefaultTimeColLabel,
								Type:  flux.TTime,
							},
							{
								Label: execute.DefaultValueColLabel,
								Type:  flux.TFloat,
							},
						},
						KeyCols:   []string{"_measurement", "_field"},
						KeyValues: []interface{}{"m0", "f0"},
						Alloc:     alloc,
					},
				},
			}
		},
		IsDone: func(tbl flux.Table) bool {
			return tbl.Empty() || tbl.(*Table).IsDone
		},
	})
}
