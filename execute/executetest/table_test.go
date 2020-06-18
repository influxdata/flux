package executetest

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
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

func TestNormalizeTables(t *testing.T) {
	got := []*Table{
		{
			ColMeta: []flux.ColMeta{
				{
					Label: "otherMeasurement",
					Type:  flux.TString,
				},
				{
					Label: "aaaaaafieldtest",
					Type:  flux.TString,
				},
				{
					Label: "aaaaaameasurement",
					Type:  flux.TString,
				},
				{
					Label: "anothertest",
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
			KeyCols: []string{"aaaaaameasurement", "aaaaaafieldtest", "anothertest"},
			Data: [][]interface{}{
				{"12", "f0", "m0", "test", execute.Time(0), 2.0},
				{"13", "f0", "m0", "test", execute.Time(10), 2.5},
				{"11", "f0", "m0", "test", execute.Time(20), 3.0},
				{"4", "f0", "m0", "test", execute.Time(30), 3.5},
			},
		},
		{
			ColMeta: []flux.ColMeta{
				{
					Label: "atestcolumn",
					Type:  flux.TString,
				},
				{
					Label: "anothertest",
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
			KeyCols:   []string{"atestcolumn", "anothertest"},
			KeyValues: []interface{}{"f0", "m0"},
		},
	}

	want := []*Table{
		{
			ColMeta: []flux.ColMeta{
				{
					Label: execute.DefaultTimeColLabel,
					Type:  flux.TTime,
				},
				{
					Label: execute.DefaultValueColLabel,
					Type:  flux.TFloat,
				},
				{
					Label: "anothertest",
					Type:  flux.TString,
				},
				{
					Label: "atestcolumn",
					Type:  flux.TString,
				},
			},
			KeyCols:   []string{"atestcolumn", "anothertest"},
			KeyValues: []interface{}{"f0", "m0"},
		},
		{
			ColMeta: []flux.ColMeta{
				{
					Label: execute.DefaultTimeColLabel,
					Type:  flux.TTime,
				},
				{
					Label: execute.DefaultValueColLabel,
					Type:  flux.TFloat,
				},
				{
					Label: "aaaaaafieldtest",
					Type:  flux.TString,
				},
				{
					Label: "aaaaaameasurement",
					Type:  flux.TString,
				},
				{
					Label: "anothertest",
					Type:  flux.TString,
				},
				{
					Label: "otherMeasurement",
					Type:  flux.TString,
				},
			},
			KeyCols: []string{"aaaaaameasurement", "aaaaaafieldtest", "anothertest"},
			Data: [][]interface{}{
				{execute.Time(0), 2.0, "f0", "m0", "test", "12"},
				{execute.Time(10), 2.5, "f0", "m0", "test", "13"},
				{execute.Time(20), 3.0, "f0", "m0", "test", "11"},
				{execute.Time(30), 3.5, "f0", "m0", "test", "4"},
			},
		},
	}
	sortByGroupKey(got)
	for i, gotTable := range got {
		if want[i].Key().String() != gotTable.Key().String() {
			t.Fatalf("tables were not in expected order: %v", cmp.Diff(want[i].Key().String(), gotTable.Key().String()))
		}
	}

	for i, gotTable := range got {
		if !reflect.DeepEqual(want[i].ColMeta, gotTable.ColMeta) {
			t.Fatalf("column metadata was not in expected order: %v", cmp.Diff(want[i].ColMeta, gotTable.ColMeta))
		}
		if !reflect.DeepEqual(want[i].Data, gotTable.Data) {
			t.Fatalf("column data was not in expected order: %v", cmp.Diff(want[i].Data, gotTable.Data))
		}
	}
}

func TestSortColumns(t *testing.T) {
	got := Table{
		ColMeta: []flux.ColMeta{
			{
				Label: "zeeBestColumn",
				Type:  flux.TString,
			},
			{
				Label: "testColumn",
				Type:  flux.TString,
			},
			{
				Label: "aTestColumn",
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
		KeyCols:   []string{"aTestColumn", "zeeBestColumn"},
		KeyValues: []interface{}{"m0", "f0"},
	}

	want := []flux.ColMeta{
		{
			Label: execute.DefaultTimeColLabel,
			Type:  flux.TTime,
		},
		{
			Label: execute.DefaultValueColLabel,
			Type:  flux.TFloat,
		},
		{
			Label: "aTestColumn",
			Type:  flux.TString,
		},
		{
			Label: "testColumn",
			Type:  flux.TString,
		},
		{
			Label: "zeeBestColumn",
			Type:  flux.TString,
		},
	}
	sortColumns(&got)
	for i, val := range got.ColMeta {
		if want[i].Label != val.Label {
			t.Fatalf("columns were not sorted correctly: %v", cmp.Diff(want, got.ColMeta))
		}
	}
}
