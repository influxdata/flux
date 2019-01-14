package universe_test

import (
	"math"
	"testing"

	"github.com/apache/arrow/go/arrow/array"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/querytest"
	"github.com/influxdata/flux/stdlib/universe"
)

func TestStddevOperation_Marshaling(t *testing.T) {
	data := []byte(`{"id":"stddev","kind":"stddev"}`)
	op := &flux.Operation{
		ID:   "stddev",
		Spec: &universe.StddevOpSpec{},
	}

	querytest.OperationMarshalingTestHelper(t, data, op)
}

func TestStddev_Process(t *testing.T) {
	testCases := []struct {
		name string
		data func() *array.Float64
		want interface{}
	}{
		{
			name: "zero",
			data: func() *array.Float64 {
				return arrow.NewFloat([]float64{1, 1, 1}, nil)
			},
			want: 0.0,
		},
		{
			name: "nonzero",
			data: func() *array.Float64 {
				return arrow.NewFloat([]float64{1, 2, 3}, nil)
			},
			want: 1.0,
		},
		{
			name: "NaN",
			data: func() *array.Float64 {
				return arrow.NewFloat([]float64{1}, nil)
			},
			want: math.NaN(),
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
				b.Append(1)
				b.AppendNull()
				b.Append(2)
				b.AppendNull()
				b.Append(3)
				return b.NewFloat64Array()
			},
			want: 1.0,
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
			executetest.AggFuncTestHelper(
				t,
				new(universe.StddevAgg),
				tc.data(),
				tc.want,
			)
		})
	}
}

func BenchmarkStddev(b *testing.B) {
	data := arrow.NewFloat(NormalData, &memory.Allocator{})
	executetest.AggFuncBenchmarkHelper(
		b,
		new(universe.StddevAgg),
		data,
		2.998926113076968,
	)
}
