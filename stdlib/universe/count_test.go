package universe_test

import (
	"testing"
	"time"

	"github.com/InfluxCommunity/flux"
	"github.com/InfluxCommunity/flux/array"
	"github.com/InfluxCommunity/flux/arrow"
	"github.com/InfluxCommunity/flux/execute"
	"github.com/InfluxCommunity/flux/execute/executetest"
	"github.com/InfluxCommunity/flux/internal/operation"
	"github.com/InfluxCommunity/flux/memory"
	"github.com/InfluxCommunity/flux/querytest"
	"github.com/InfluxCommunity/flux/stdlib/influxdata/influxdb"
	"github.com/InfluxCommunity/flux/stdlib/universe"
)

func TestCount_NewQuery(t *testing.T) {
	tests := []querytest.NewQueryTestCase{
		{
			Name: "from with range and count",
			Raw:  `from(bucket:"mydb") |> range(start:-4h, stop:-2h) |> count()`,
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
						ID: "count2",
						Spec: &universe.CountOpSpec{
							SimpleAggregateConfig: execute.DefaultSimpleAggregateConfig,
						},
					},
				},
				Edges: []operation.Edge{
					{Parent: "from0", Child: "range1"},
					{Parent: "range1", Child: "count2"},
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

func TestCount_Process(t *testing.T) {
	testCases := []struct {
		name string
		data func() *array.Float
		want int64
	}{
		{
			name: "zero",
			data: func() *array.Float {
				return arrow.NewFloat([]float64{0, 0, 0}, nil)
			},
			want: 3,
		},
		{
			name: "nonzero",
			data: func() *array.Float {
				return arrow.NewFloat([]float64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, nil)
			},
			want: 10,
		},
		{
			name: "empty",
			data: func() *array.Float {
				return arrow.NewFloat(nil, nil)
			},
			want: 0,
		},
		{
			name: "with nulls",
			data: func() *array.Float {
				b := arrow.NewFloatBuilder(nil)
				defer b.Release()
				b.AppendValues([]float64{0, 1, 2, 3}, nil)
				b.AppendNull()
				b.AppendValues([]float64{5, 6}, nil)
				b.AppendNull()
				b.AppendValues([]float64{8, 9}, nil)
				return b.NewFloatArray()
			},
			want: 10,
		},
		{
			name: "only nulls",
			data: func() *array.Float {
				b := arrow.NewFloatBuilder(nil)
				defer b.Release()
				b.AppendNull()
				b.AppendNull()
				return b.NewFloatArray()
			},
			want: 2,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			data := tc.data()
			defer data.Release()

			executetest.AggFuncTestHelper(
				t,
				new(universe.CountAgg),
				data,
				tc.want,
			)
		})
	}
}
func BenchmarkCount(b *testing.B) {
	data := arrow.NewFloat(NormalData, &memory.ResourceAllocator{})
	executetest.AggFuncBenchmarkHelper(
		b,
		new(universe.CountAgg),
		data,
		int64(len(NormalData)),
	)
}
