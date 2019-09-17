package universe_test

import (
	"testing"
	"time"

	"github.com/apache/arrow/go/arrow/array"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/querytest"
	"github.com/influxdata/flux/stdlib/influxdata/influxdb"
	"github.com/influxdata/flux/stdlib/universe"
)

func TestQuantile_NewQuery(t *testing.T) {
	tests := []querytest.NewQueryTestCase{
		{
			Name: "tdigest",
			Raw:  `from(bucket:"testdb") |> range(start: -1h) |> quantile(q: 0.99, method: "estimate_tdigest", compression: 1000.0)`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: "testdb",
						},
					},
					{
						ID: "range1",
						Spec: &universe.RangeOpSpec{
							Start: flux.Time{
								Relative:   -1 * time.Hour,
								IsRelative: true,
							},
							Stop: flux.Time{
								IsRelative: true,
							},
							TimeColumn:  "_time",
							StartColumn: "_start",
							StopColumn:  "_stop",
						},
					},
					{
						ID: "quantile2",
						Spec: &universe.QuantileOpSpec{
							Quantile:        0.99,
							Compression:     1000,
							Method:          "estimate_tdigest",
							AggregateConfig: execute.DefaultAggregateConfig,
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "range1"},
					{Parent: "range1", Child: "quantile2"},
				},
			},
		},
		{
			Name: "exact_mean",
			Raw:  `from(bucket:"testdb") |> range(start: -1h) |> quantile(q: 0.99, method: "exact_mean")`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: "testdb",
						},
					},
					{
						ID: "range1",
						Spec: &universe.RangeOpSpec{
							Start: flux.Time{
								Relative:   -1 * time.Hour,
								IsRelative: true,
							},
							Stop: flux.Time{
								IsRelative: true,
							},
							TimeColumn:  "_time",
							StartColumn: "_start",
							StopColumn:  "_stop",
						},
					},
					{
						ID: "quantile2",
						Spec: &universe.QuantileOpSpec{
							Quantile:        0.99,
							Method:          "exact_mean",
							AggregateConfig: execute.DefaultAggregateConfig,
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "range1"},
					{Parent: "range1", Child: "quantile2"},
				},
			},
		},
		{
			Name: "exact_selector",
			Raw:  `from(bucket:"testdb") |> range(start: -1h) |> quantile(q: 0.99, method: "exact_selector")`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: "testdb",
						},
					},
					{
						ID: "range1",
						Spec: &universe.RangeOpSpec{
							Start: flux.Time{
								Relative:   -1 * time.Hour,
								IsRelative: true,
							},
							Stop: flux.Time{
								IsRelative: true,
							},
							TimeColumn:  "_time",
							StartColumn: "_start",
							StopColumn:  "_stop",
						},
					},
					{
						ID: "quantile2",
						Spec: &universe.QuantileOpSpec{
							Quantile:       0.99,
							Method:         "exact_selector",
							SelectorConfig: execute.DefaultSelectorConfig,
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "range1"},
					{Parent: "range1", Child: "quantile2"},
				},
			},
		},
		{
			Name: "custom col",
			Raw:  `from(bucket:"testdb") |> range(start: -1h) |> quantile(q: 0.99, method: "exact_selector", column: "foo")`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: "testdb",
						},
					},
					{
						ID: "range1",
						Spec: &universe.RangeOpSpec{
							Start: flux.Time{
								Relative:   -1 * time.Hour,
								IsRelative: true,
							},
							Stop: flux.Time{
								IsRelative: true,
							},
							TimeColumn:  "_time",
							StartColumn: "_start",
							StopColumn:  "_stop",
						},
					},
					{
						ID: "quantile2",
						Spec: &universe.QuantileOpSpec{
							Quantile: 0.99,
							Method:   "exact_selector",
							SelectorConfig: execute.SelectorConfig{
								Column: "foo",
							},
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "range1"},
					{Parent: "range1", Child: "quantile2"},
				},
			},
		},
		{
			Name: "custom column",
			Raw:  `from(bucket:"testdb") |> range(start: -1h) |> quantile(q: 0.99, method: "exact_mean", column: "foo")`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: "testdb",
						},
					},
					{
						ID: "range1",
						Spec: &universe.RangeOpSpec{
							Start: flux.Time{
								Relative:   -1 * time.Hour,
								IsRelative: true,
							},
							Stop: flux.Time{
								IsRelative: true,
							},
							TimeColumn:  "_time",
							StartColumn: "_start",
							StopColumn:  "_stop",
						},
					},
					{
						ID: "quantile2",
						Spec: &universe.QuantileOpSpec{
							Quantile: 0.99,
							Method:   "exact_mean",
							AggregateConfig: execute.AggregateConfig{
								Columns: []string{"foo"},
							},
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "range1"},
					{Parent: "range1", Child: "quantile2"},
				},
			},
		},
		// errors
		{
			Name:    "wrong method",
			Raw:     `from(bucket:"testdb") |> range(start: -1h) |> quantile(q: 0.99, method: "non_existent_method")`,
			WantErr: true,
		},
		{
			Name:    "non-tdigest with compression",
			Raw:     `from(bucket:"testdb") |> range(start: -1h) |> quantile(q: 0.99, method: "exact_mean", compression: 800.0)`,
			WantErr: true,
		},
		{
			Name:    "selector with columns",
			Raw:     `from(bucket:"testdb") |> range(start: -1h) |> quantile(q: 0.99, method: "exact_selector", columns: ["1", "2"])`,
			WantErr: true,
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

func TestQuantileOperation_Marshaling(t *testing.T) {
	data := []byte(`{"id":"quantile","kind":"quantile","spec":{"quantile":0.9}}`)
	op := &flux.Operation{
		ID: "quantile",
		Spec: &universe.QuantileOpSpec{
			Quantile: 0.9,
		},
	}

	querytest.OperationMarshalingTestHelper(t, data, op)
}

func TestQuantile_Process(t *testing.T) {
	testCases := []struct {
		name     string
		data     func() *array.Float64
		quantile float64
		exact    bool
		want     interface{}
	}{
		{
			name: "zero",
			data: func() *array.Float64 {
				return arrow.NewFloat([]float64{0, 0, 0}, nil)
			},
			quantile: 0.5,
			want:     0.0,
		},
		{
			name: "50th",
			data: func() *array.Float64 {
				return arrow.NewFloat([]float64{1, 2, 3, 4, 5, 5, 4, 3, 2, 1}, nil)
			},
			quantile: 0.5,
			want:     3.0,
		},
		{
			name: "75th",
			data: func() *array.Float64 {
				return arrow.NewFloat([]float64{1, 2, 3, 4, 5, 5, 4, 3, 2, 1}, nil)
			},
			quantile: 0.75,
			want:     4.0,
		},
		{
			name: "90th",
			data: func() *array.Float64 {
				return arrow.NewFloat([]float64{1, 2, 3, 4, 5, 5, 4, 3, 2, 1}, nil)
			},
			quantile: 0.9,
			want:     5.0,
		},
		{
			name: "99th",
			data: func() *array.Float64 {
				return arrow.NewFloat([]float64{1, 2, 3, 4, 5, 5, 4, 3, 2, 1}, nil)
			},
			quantile: 0.99,
			want:     5.0,
		},
		{
			name: "exact 50th",
			data: func() *array.Float64 {
				return arrow.NewFloat([]float64{1, 2, 3, 4, 5}, nil)
			},
			quantile: 0.5,
			exact:    true,
			want:     3.0,
		},
		{
			name: "exact 75th",
			data: func() *array.Float64 {
				return arrow.NewFloat([]float64{1, 2, 3, 4, 5}, nil)
			},
			quantile: 0.75,
			exact:    true,
			want:     4.0,
		},
		{
			name: "exact 90th",
			data: func() *array.Float64 {
				return arrow.NewFloat([]float64{1, 2, 3, 4, 5}, nil)
			},
			quantile: 0.9,
			exact:    true,
			want:     4.6,
		},
		{
			name: "exact 99th",
			data: func() *array.Float64 {
				return arrow.NewFloat([]float64{1, 2, 3, 4, 5}, nil)
			},
			quantile: 0.99,
			exact:    true,
			want:     4.96,
		},
		{
			name: "exact 100th",
			data: func() *array.Float64 {
				return arrow.NewFloat([]float64{1, 2, 3, 4, 5}, nil)
			},
			quantile: 1,
			exact:    true,
			want:     5.0,
		},
		{
			name: "exact 50th normal",
			data: func() *array.Float64 {
				return arrow.NewFloat(NormalData, nil)
			},
			quantile: 0.5,
			exact:    true,
			want:     10.000736834856248,
		},
		{
			name: "normal",
			data: func() *array.Float64 {
				return arrow.NewFloat(NormalData, nil)
			},
			quantile: 0.9,
			want:     13.842132136909889,
		},
		{
			name: "empty",
			data: func() *array.Float64 {
				return arrow.NewFloat(nil, nil)
			},
			want: nil,
		},
		{
			name: "with nulls",
			data: func() *array.Float64 {
				b := arrow.NewFloatBuilder(nil)
				defer b.Release()
				b.AppendValues([]float64{1, 3, 3}, nil)
				b.AppendNull()
				b.AppendValues([]float64{5, 5, 4, 3}, nil)
				b.AppendNull()
				b.AppendValues([]float64{1}, nil)
				return b.NewFloat64Array()
			},
			quantile: 0.5,
			want:     3.0,
		},
		{
			name: "only nulls",
			data: func() *array.Float64 {
				b := arrow.NewFloatBuilder(nil)
				defer b.Release()
				b.AppendNull()
				b.AppendNull()
				return b.NewFloat64Array()
			},
			want: nil,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var agg execute.Aggregate
			if tc.exact {
				agg = &universe.ExactQuantileAgg{Quantile: tc.quantile}
			} else {
				agg = &universe.QuantileAgg{
					Quantile:    tc.quantile,
					Compression: 1000,
				}
			}
			executetest.AggFuncTestHelper(
				t,
				agg,
				tc.data(),
				tc.want,
			)
		})
	}
}

func TestQuantileSelector_Process(t *testing.T) {
	testCases := []struct {
		name     string
		quantile float64
		data     []flux.Table
		want     []*executetest.Table
	}{
		{
			name:     "select_10",
			quantile: 0.1,
			data: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"t1"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "t1", Type: flux.TString},
						{Label: "t2", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(0), 1.0, "a", "y"},
						{execute.Time(10), 2.0, "a", "x"},
						{execute.Time(20), 3.0, "a", "y"},
						{execute.Time(30), 4.0, "a", "x"},
						{execute.Time(40), 5.0, "a", "y"},
					},
				}},
			want: []*executetest.Table{
				{
					KeyCols: []string{"t1"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "t1", Type: flux.TString},
						{Label: "t2", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(0), 1.0, "a", "y"},
					},
				},
			},
		},
		{
			name:     "select_20",
			quantile: 0.2,
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"t1"},
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "t1", Type: flux.TString},
					{Label: "t2", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(0), 1.0, "a", "y"},
					{execute.Time(10), 2.0, "a", "x"},
					{execute.Time(20), 3.0, "a", "y"},
					{execute.Time(30), 4.0, "a", "x"},
					{execute.Time(40), 5.0, "a", "y"},
				},
			}},
			want: []*executetest.Table{
				{
					KeyCols: []string{"t1"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "t1", Type: flux.TString},
						{Label: "t2", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(0), 1.0, "a", "y"},
					},
				}},
		},
		{
			name:     "select_40",
			quantile: 0.4,
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"t1"},
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "t1", Type: flux.TString},
					{Label: "t2", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(0), 1.0, "a", "y"},
					{execute.Time(10), 2.0, "a", "x"},
					{execute.Time(20), 3.0, "a", "y"},
					{execute.Time(30), 4.0, "a", "x"},
					{execute.Time(40), 5.0, "a", "y"},
				},
			}},
			want: []*executetest.Table{{
				KeyCols: []string{"t1"},
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "t1", Type: flux.TString},
					{Label: "t2", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(10), 2.0, "a", "x"},
				},
			}},
		},
		{
			name:     "select_50",
			quantile: 0.5,
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"t1"},
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "t1", Type: flux.TString},
					{Label: "t2", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(0), 1.0, "a", "y"},
					{execute.Time(10), 2.0, "a", "x"},
					{execute.Time(20), 3.0, "a", "y"},
					{execute.Time(30), 4.0, "a", "x"},
					{execute.Time(40), 5.0, "a", "y"},
				},
			}},
			want: []*executetest.Table{{
				KeyCols: []string{"t1"},
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "t1", Type: flux.TString},
					{Label: "t2", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(20), 3.0, "a", "y"},
				},
			}},
		},
		{
			name:     "select_80",
			quantile: 0.8,
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"t1"},
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "t1", Type: flux.TString},
					{Label: "t2", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(0), 1.0, "a", "y"},
					{execute.Time(10), 2.0, "a", "x"},
					{execute.Time(20), 3.0, "a", "y"},
					{execute.Time(30), 4.0, "a", "x"},
					{execute.Time(40), 5.0, "a", "y"},
				},
			}},
			want: []*executetest.Table{{
				KeyCols: []string{"t1"},
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "t1", Type: flux.TString},
					{Label: "t2", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(30), 4.0, "a", "x"},
				},
			}},
		},
		{
			name:     "select_90",
			quantile: 0.9,
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"t1"},
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "t1", Type: flux.TString},
					{Label: "t2", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(0), 1.0, "a", "y"},
					{execute.Time(10), 2.0, "a", "x"},
					{execute.Time(20), 3.0, "a", "y"},
					{execute.Time(30), 4.0, "a", "x"},
					{execute.Time(40), 5.0, "a", "y"},
				},
			}},
			want: []*executetest.Table{{
				KeyCols: []string{"t1"},
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "t1", Type: flux.TString},
					{Label: "t2", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(40), 5.0, "a", "y"},
				},
			}},
		},
		{
			name:     "select_100",
			quantile: 1.0,
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"t1"},
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "t1", Type: flux.TString},
					{Label: "t2", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(0), 1.0, "a", "y"},
					{execute.Time(10), 2.0, "a", "x"},
					{execute.Time(20), 3.0, "a", "y"},
					{execute.Time(30), 4.0, "a", "x"},
					{execute.Time(40), 5.0, "a", "y"},
				},
			}},
			want: []*executetest.Table{{
				KeyCols: []string{"t1"},
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "t1", Type: flux.TString},
					{Label: "t2", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(40), 5.0, "a", "y"},
				},
			}},
		},
		{
			name:     "select_50_nulls",
			quantile: 0.5,
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"t1"},
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "t1", Type: flux.TString},
					{Label: "t2", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(0), 1.0, "a", "y"},
					{execute.Time(10), nil, "a", "x"},
					{execute.Time(20), 2.0, "a", "x"},
					{execute.Time(30), 3.0, "a", "y"},
					{execute.Time(40), nil, "a", "y"},
					{execute.Time(50), 4.0, "a", "x"},
					{execute.Time(60), 5.0, "a", "y"},
				},
			}},
			want: []*executetest.Table{{
				KeyCols: []string{"t1"},
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "t1", Type: flux.TString},
					{Label: "t2", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(30), 3.0, "a", "y"},
				},
			}},
		},
		{
			name:     "empty",
			quantile: 0.5,
			data: []flux.Table{&executetest.Table{
				KeyCols:   []string{"t1"},
				KeyValues: []interface{}{"a"},
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "t1", Type: flux.TString},
					{Label: "t2", Type: flux.TString},
				},
				Data: [][]interface{}{},
			}},
			want: []*executetest.Table{{
				KeyCols: []string{"t1"},
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "t1", Type: flux.TString},
					{Label: "t2", Type: flux.TString},
				},
				Data: [][]interface{}{
					{nil, nil, "a", nil},
				},
			}},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			executetest.ProcessTestHelper(
				t,
				tc.data,
				tc.want,
				nil,
				func(d execute.Dataset, c execute.TableBuilderCache) execute.Transformation {
					return universe.NewExactQuantileSelectorTransformation(d, c, &universe.ExactQuantileSelectProcedureSpec{Quantile: tc.quantile}, executetest.UnlimitedAllocator)
				},
			)
		})
	}
}

func BenchmarkQuantile(b *testing.B) {
	data := arrow.NewFloat(NormalData, &memory.Allocator{})
	executetest.AggFuncBenchmarkHelper(
		b,
		&universe.QuantileAgg{
			Quantile:    0.9,
			Compression: 1000,
		},
		data,
		13.842132136909889,
	)
}
