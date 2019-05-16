package executetest

import (
	"testing"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
)

func TestTable(t *testing.T) {
	RunTableTests(t, TableTest{
		CreateTableFn: func() flux.Table {
			return &Table{
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
			}
		},
		CreateEmptyTableFn: func() flux.Table {
			return &Table{
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
			}
		},
		IsDone: func(tbl flux.Table) bool {
			return len(tbl.(*Table).Data) == 0 || tbl.(*Table).IsDone
		},
	})
}
