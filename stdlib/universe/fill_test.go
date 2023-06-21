package universe_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/InfluxCommunity/flux/dependencies/dependenciestest"
	"github.com/InfluxCommunity/flux/dependency"
	"github.com/InfluxCommunity/flux/internal/gen"
	"github.com/InfluxCommunity/flux/internal/operation"
	"github.com/InfluxCommunity/flux/memory"
	"github.com/InfluxCommunity/flux/semantic"
	"github.com/InfluxCommunity/flux/values"

	"github.com/InfluxCommunity/flux/stdlib/influxdata/influxdb"

	"github.com/InfluxCommunity/flux/plan"

	"github.com/InfluxCommunity/flux"
	"github.com/InfluxCommunity/flux/execute"
	"github.com/InfluxCommunity/flux/execute/executetest"
	"github.com/InfluxCommunity/flux/querytest"
	"github.com/InfluxCommunity/flux/stdlib/universe"
)

func TestFill_NewQuery(t *testing.T) {
	tests := []querytest.NewQueryTestCase{
		{
			Name: "from with range and fill",
			Raw:  `from(bucket:"mydb") |> range(start:-4h, stop:-2h) |> fill(column: "c1", value: 1.0)`,
			Want: &operation.Spec{
				Operations: []*operation.Node{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: influxdb.NameOrID{Name: "mydb"},
						},
					},
					{
						ID: "range1",
						Spec: &universe.RangeOpSpec{
							Start: flux.Time{
								Relative:   -4 * time.Hour,
								IsRelative: true,
							},
							Stop: flux.Time{
								Relative:   -2 * time.Hour,
								IsRelative: true,
							},
							TimeColumn:  "_time",
							StartColumn: "_start",
							StopColumn:  "_stop",
						},
					},
					{
						ID: "fill2",
						Spec: &universe.FillOpSpec{
							Column: "c1",
							Type:   "float",
							Value:  "1",
						},
					},
				},
				Edges: []operation.Edge{
					{Parent: "from0", Child: "range1"},
					{Parent: "range1", Child: "fill2"},
				},
			},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			querytest.NewQueryTestHelper(t, tc)
		})
	}
}

func TestFill_Process(t *testing.T) {
	testCases := []struct {
		name string
		spec *universe.FillProcedureSpec
		data func() []flux.Table
		want []*executetest.Table
	}{
		{
			name: "nothing to fill",
			spec: &universe.FillProcedureSpec{
				Column: "_value",
				Value:  values.New(0.0),
			},
			data: func() []flux.Table {
				return []flux.Table{&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), 2.0},
						{execute.Time(2), 1.0},
					},
				}}
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 2.0},
					{execute.Time(2), 1.0},
				},
			}},
		},
		{
			name: "null bool",
			spec: &universe.FillProcedureSpec{
				Column: "_value",
				Value:  values.New(false),
			},
			data: func() []flux.Table {
				return []flux.Table{&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TBool},
					},
					Data: [][]interface{}{
						{execute.Time(1), true},
						{execute.Time(2), nil},
						{execute.Time(3), false},
						{execute.Time(4), nil},
					},
				}}
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TBool},
				},
				Data: [][]interface{}{
					{execute.Time(1), true},
					{execute.Time(2), false},
					{execute.Time(3), false},
					{execute.Time(4), false},
				},
			}},
		},
		{
			name: "missing bool fill col",
			spec: &universe.FillProcedureSpec{
				Column: "_value",
				Value:  values.New(false),
			},
			data: func() []flux.Table {
				return []flux.Table{&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
					},
					Data: [][]interface{}{
						{execute.Time(1)},
						{execute.Time(2)},
						{execute.Time(3)},
						{execute.Time(4)},
					},
				}}
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TBool},
				},
				Data: [][]interface{}{
					{execute.Time(1), false},
					{execute.Time(2), false},
					{execute.Time(3), false},
					{execute.Time(4), false},
				},
			}},
		},
		{
			name: "null int",
			spec: &universe.FillProcedureSpec{
				Column: "_value",
				Value:  values.New(int64(-1)),
			},
			data: func() []flux.Table {
				return []flux.Table{&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TInt},
					},
					Data: [][]interface{}{
						{execute.Time(1), int64(2)},
						{execute.Time(2), nil},
						{execute.Time(3), int64(4)},
						{execute.Time(4), nil},
					},
				}}
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{execute.Time(1), int64(2)},
					{execute.Time(2), int64(-1)},
					{execute.Time(3), int64(4)},
					{execute.Time(4), int64(-1)},
				},
			}},
		},
		{
			name: "missing int fill col",
			spec: &universe.FillProcedureSpec{
				Column: "_value",
				Value:  values.New(int64(-1)),
			},
			data: func() []flux.Table {
				return []flux.Table{&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
					},
					Data: [][]interface{}{
						{execute.Time(1)},
						{execute.Time(2)},
						{execute.Time(3)},
						{execute.Time(4)},
					},
				}}
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{execute.Time(1), int64(-1)},
					{execute.Time(2), int64(-1)},
					{execute.Time(3), int64(-1)},
					{execute.Time(4), int64(-1)},
				},
			}},
		},
		{
			name: "null uint",
			spec: &universe.FillProcedureSpec{
				Column: "_value",
				Value:  values.New(uint64(0)),
			},
			data: func() []flux.Table {
				return []flux.Table{&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TUInt},
					},
					Data: [][]interface{}{
						{execute.Time(1), uint64(2)},
						{execute.Time(2), nil},
						{execute.Time(3), uint64(4)},
						{execute.Time(4), nil},
					},
				}}
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TUInt},
				},
				Data: [][]interface{}{
					{execute.Time(1), uint64(2)},
					{execute.Time(2), uint64(0)},
					{execute.Time(3), uint64(4)},
					{execute.Time(4), uint64(0)},
				},
			}},
		},
		{
			name: "missing uint fill col",
			spec: &universe.FillProcedureSpec{
				Column: "_value",
				Value:  values.New(uint64(0)),
			},
			data: func() []flux.Table {
				return []flux.Table{&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
					},
					Data: [][]interface{}{
						{execute.Time(1)},
						{execute.Time(2)},
						{execute.Time(3)},
						{execute.Time(4)},
					},
				}}
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TUInt},
				},
				Data: [][]interface{}{
					{execute.Time(1), uint64(0)},
					{execute.Time(2), uint64(0)},
					{execute.Time(3), uint64(0)},
					{execute.Time(4), uint64(0)},
				},
			}},
		},
		{
			name: "null float",
			spec: &universe.FillProcedureSpec{
				Column: "_value",
				Value:  values.New(0.0),
			},
			data: func() []flux.Table {
				return []flux.Table{&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), 2.0},
						{execute.Time(2), nil},
						{execute.Time(3), 4.0},
						{execute.Time(4), nil},
					},
				}}
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 2.0},
					{execute.Time(2), 0.0},
					{execute.Time(3), 4.0},
					{execute.Time(4), 0.0},
				},
			}},
		},
		{
			name: "missing float fill col",
			spec: &universe.FillProcedureSpec{
				Column: "_value",
				Value:  values.New(0.0),
			},
			data: func() []flux.Table {
				return []flux.Table{&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
					},
					Data: [][]interface{}{
						{execute.Time(1)},
						{execute.Time(2)},
						{execute.Time(3)},
						{execute.Time(4)},
					},
				}}
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 0.0},
					{execute.Time(2), 0.0},
					{execute.Time(3), 0.0},
					{execute.Time(4), 0.0},
				},
			}},
		},
		{
			name: "null string",
			spec: &universe.FillProcedureSpec{
				Column: "_value",
				Value:  values.New("UNK"),
			},
			data: func() []flux.Table {
				return []flux.Table{&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), "A"},
						{execute.Time(2), nil},
						{execute.Time(3), "B"},
						{execute.Time(4), nil},
					},
				}}
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(1), "A"},
					{execute.Time(2), "UNK"},
					{execute.Time(3), "B"},
					{execute.Time(4), "UNK"},
				},
			}},
		},
		{
			name: "missing string fill col",
			spec: &universe.FillProcedureSpec{
				Column: "_value",
				Value:  values.New("UNK"),
			},
			data: func() []flux.Table {
				return []flux.Table{&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
					},
					Data: [][]interface{}{
						{execute.Time(1)},
						{execute.Time(2)},
						{execute.Time(3)},
						{execute.Time(4)},
					},
				}}
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(1), "UNK"},
					{execute.Time(2), "UNK"},
					{execute.Time(3), "UNK"},
					{execute.Time(4), "UNK"},
				},
			}},
		},
		{
			name: "null time",
			spec: &universe.FillProcedureSpec{
				Column: "_time",
				Value:  values.New(execute.Time(0)),
			},
			data: func() []flux.Table {
				return []flux.Table{&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), "A"},
						{nil, "B"},
						{execute.Time(3), "B"},
						{nil, "C"},
					},
				}}
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(1), "A"},
					{execute.Time(0), "B"},
					{execute.Time(3), "B"},
					{execute.Time(0), "C"},
				},
			}},
		},
		{
			name: "missing time fill col",
			spec: &universe.FillProcedureSpec{
				Column: "_time",
				Value:  values.New(execute.Time(0)),
			},
			data: func() []flux.Table {
				return []flux.Table{&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_value", Type: flux.TString},
					},
					Data: [][]interface{}{
						{"A"},
						{"B"},
						{"B"},
						{"C"},
					},
				}}
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(0), "A"},
					{execute.Time(0), "B"},
					{execute.Time(0), "B"},
					{execute.Time(0), "C"},
				},
			}},
		},
		{
			name: "fill previous",
			spec: &universe.FillProcedureSpec{
				DefaultCost: plan.DefaultCost{},
				Column:      "_value",
				UsePrevious: true,
			},
			data: func() []flux.Table {
				return []flux.Table{&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), "A"},
						{execute.Time(2), nil},
						{execute.Time(3), "B"},
						{execute.Time(4), nil},
					},
				}}
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(1), "A"},
					{execute.Time(2), "A"},
					{execute.Time(3), "B"},
					{execute.Time(4), "B"},
				},
			}},
		},
		{
			name: "fill previous unknown column",
			spec: &universe.FillProcedureSpec{
				DefaultCost: plan.DefaultCost{},
				Column:      "nonexistent",
				UsePrevious: true,
			},
			data: func() []flux.Table {
				return []flux.Table{&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), "A"},
						{execute.Time(2), "B"},
					},
				}}
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(1), "A"},
					{execute.Time(2), "B"},
				},
			}},
		},
		{
			name: "fill previous first nil",
			spec: &universe.FillProcedureSpec{
				DefaultCost: plan.DefaultCost{},
				Column:      "_value",
				UsePrevious: true,
			},
			data: func() []flux.Table {
				return []flux.Table{&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), nil},
						{execute.Time(2), "A"},
						{execute.Time(3), "B"},
						{execute.Time(4), nil},
					},
				}}
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(1), nil},
					{execute.Time(2), "A"},
					{execute.Time(3), "B"},
					{execute.Time(4), "B"},
				},
			}},
		},
		{
			name: "fill previous multiple buffers",
			spec: &universe.FillProcedureSpec{
				DefaultCost: plan.DefaultCost{},
				Column:      "_value",
				UsePrevious: true,
			},
			data: func() []flux.Table {
				return []flux.Table{&executetest.RowWiseTable{
					Table: &executetest.Table{
						ColMeta: []flux.ColMeta{
							{Label: "_time", Type: flux.TTime},
							{Label: "_value", Type: flux.TString},
						},
						Data: [][]interface{}{
							{execute.Time(1), nil},
							{execute.Time(2), "A"},
							{execute.Time(3), nil},
							{execute.Time(4), "B"},
						},
					},
				}}
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(1), nil},
					{execute.Time(2), "A"},
					{execute.Time(3), "A"},
					{execute.Time(4), "B"},
				},
			}},
		},
		{
			name: "fill previous empty table",
			spec: &universe.FillProcedureSpec{
				DefaultCost: plan.DefaultCost{},
				Column:      "_value",
				UsePrevious: true,
			},
			data: func() []flux.Table {
				return []flux.Table{&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TString},
					},
					Data: [][]interface{}{},
				}}
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TString},
				},
				Data: [][]interface{}(nil),
			}},
		},
		{
			name: "null group key",
			spec: &universe.FillProcedureSpec{
				Column: "tag0",
				Value:  values.New(0.0),
			},
			data: func() []flux.Table {
				return []flux.Table{&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "tag0", Type: flux.TFloat},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), nil, 2.0},
						{execute.Time(2), nil, nil},
						{execute.Time(3), nil, 4.0},
						{execute.Time(4), nil, nil},
					},
					GroupKey: execute.NewGroupKey(
						[]flux.ColMeta{
							{Label: "tag0", Type: flux.TFloat},
						},
						[]values.Value{values.NewNull(semantic.BasicFloat)},
					),
				}}
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "tag0", Type: flux.TFloat},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 0.0, 2.0},
					{execute.Time(2), 0.0, nil},
					{execute.Time(3), 0.0, 4.0},
					{execute.Time(4), 0.0, nil},
				},
				KeyCols:   []string{"tag0"},
				KeyValues: []interface{}{0.0},
			}},
		},
		{
			name: "non null group key",
			spec: &universe.FillProcedureSpec{
				Column: "tag0",
				Value:  values.New(0.0),
			},
			data: func() []flux.Table {
				return []flux.Table{&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "tag0", Type: flux.TFloat},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), 1.0, 2.0},
						{execute.Time(2), 1.0, nil},
						{execute.Time(3), 1.0, 4.0},
						{execute.Time(4), 1.0, nil},
					},
					GroupKey: execute.NewGroupKey(
						[]flux.ColMeta{
							{Label: "tag0", Type: flux.TFloat},
						},
						[]values.Value{values.NewFloat(1.0)},
					),
				}}
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "tag0", Type: flux.TFloat},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 1.0, 2.0},
					{execute.Time(2), 1.0, nil},
					{execute.Time(3), 1.0, 4.0},
					{execute.Time(4), 1.0, nil},
				},
				KeyCols:   []string{"tag0"},
				KeyValues: []interface{}{1.0},
			}},
		},
	}
	for _, tc := range testCases {
		tc := tc
		// fill tests
		t.Run(tc.name, func(t *testing.T) {
			executetest.ProcessTestHelper2(
				t,
				tc.data(),
				tc.want,
				nil,
				func(id execute.DatasetID, alloc memory.Allocator) (execute.Transformation, execute.Dataset) {
					ctx, deps := dependency.Inject(context.Background(), dependenciestest.Default())
					defer deps.Finish()
					return universe.NewFillTransformation(ctx, tc.spec, id, alloc)
				},
			)
		})

		// fill narrow transformations tests
		// need to ensure the testcase names are distinct to avoid test results colliding between these two runs.
		t.Run(fmt.Sprintf("%s narrow", tc.name), func(t *testing.T) {
			executetest.ProcessTestHelper2(
				t,
				tc.data(),
				tc.want,
				nil,
				func(id execute.DatasetID, alloc memory.Allocator) (execute.Transformation, execute.Dataset) {
					ctx, deps := dependency.Inject(context.Background(), dependenciestest.Default())
					defer deps.Finish()
					tr, d, err := universe.NewNarrowFillTransformation(ctx, tc.spec, id, alloc)
					if err != nil {
						t.Fatal(err)
					}
					return tr, d
				},
			)
		})
	}
}

func BenchmarkFill_Values(b *testing.B) {
	b.Run("1000000", func(b *testing.B) {
		benchmarkFill(b, 1000000)
	})
}

func benchmarkFill(b *testing.B, n int) {
	b.ReportAllocs()
	spec := &universe.FillProcedureSpec{
		Column: "_value",
		Value:  values.NewFloat(0),
	}
	executetest.ProcessBenchmarkHelper(b,
		func(alloc memory.Allocator) (flux.TableIterator, error) {
			schema := gen.Schema{
				NumPoints: n,
				Alloc:     alloc,
				Tags: []gen.Tag{
					{Name: "_measurement", Cardinality: 1},
					{Name: "_field", Cardinality: 1},
					{Name: "t0", Cardinality: 1},
					{Name: "t1", Cardinality: 1},
					{Name: "t2", Cardinality: 1},
					{Name: "t3", Cardinality: 1},
					{Name: "t4", Cardinality: 1},
					{Name: "t5", Cardinality: 1},
				},
				Nulls: 0.4,
			}
			return gen.Input(context.Background(), schema)
		},
		func(id execute.DatasetID, alloc memory.Allocator) (execute.Transformation, execute.Dataset) {
			return universe.NewFillTransformation(context.Background(), spec, id, alloc)
		},
	)
}
