package universe_test

import (
	"math"
	"testing"

	"github.com/apache/arrow/go/arrow/array"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/stdlib/universe"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/querytest"
)

func TestMeanOperation_Marshaling(t *testing.T) {
	data := []byte(`{"id":"mean","kind":"mean"}`)
	op := &flux.Operation{
		ID:   "mean",
		Spec: &universe.MeanOpSpec{},
	}

	querytest.OperationMarshalingTestHelper(t, data, op)
}

func TestMean_Process(t *testing.T) {
	testCases := []struct {
		name string
		data func() *array.Float64
		want float64
	}{
		{
			name: "zero",
			data: func() *array.Float64 {
				return arrow.NewFloat([]float64{0, 0, 0}, nil)
			},
			want: 0.0,
		},
		{
			name: "nonzero",
			data: func() *array.Float64 {
				return arrow.NewFloat([]float64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, nil)
			},
			want: 4.5,
		},
		{
			name: "NaN",
			data: func() *array.Float64 {
				return arrow.NewFloat(nil, nil)
			},
			want: math.NaN(),
		},
		{
			name: "with nulls",
			data: func() *array.Float64 {
				b := arrow.NewFloatBuilder(nil)
				defer b.Release()
				b.AppendValues([]float64{0, 1, 2, 3}, nil)
				b.AppendNull()
				b.AppendValues([]float64{5, 6}, nil)
				b.AppendNull()
				b.AppendValues([]float64{8, 9}, nil)
				return b.NewFloat64Array()
			},
			want: 4.25,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			data := tc.data()
			defer data.Release()

			executetest.AggFuncTestHelper(
				t,
				new(universe.MeanAgg),
				data,
				tc.want,
			)
		})
	}
}

func BenchmarkMean(b *testing.B) {
	data := arrow.NewFloat(NormalData, &memory.Allocator{})
	executetest.AggFuncBenchmarkHelper(
		b,
		new(universe.MeanAgg),
		data,
		10.00081696729983,
	)
}
