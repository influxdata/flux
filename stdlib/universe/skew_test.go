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

func TestSkewOperation_Marshaling(t *testing.T) {
	data := []byte(`{"id":"skew","kind":"skew"}`)
	op := &flux.Operation{
		ID:   "skew",
		Spec: &universe.SkewOpSpec{},
	}

	querytest.OperationMarshalingTestHelper(t, data, op)
}

func TestSkew_Process(t *testing.T) {
	testCases := []struct {
		name string
		data func() *array.Float64
		want interface{}
	}{
		{
			name: "zero",
			data: func() *array.Float64 {
				return arrow.NewFloat([]float64{1, 2, 3}, nil)
			},
			want: 0.0,
		},
		{
			name: "nonzero",
			data: func() *array.Float64 {
				return arrow.NewFloat([]float64{2, 2, 3}, nil)
			},
			want: 0.7071067811865475,
		},
		{
			name: "nonzero 2",
			data: func() *array.Float64 {
				return arrow.NewFloat([]float64{2, 2, 3, 4}, nil)
			},
			want: 0.49338220021815854,
		},
		{
			name: "NaN short",
			data: func() *array.Float64 {
				return arrow.NewFloat([]float64{1}, nil)
			},
			want: math.NaN(),
		},
		{
			name: "NaN divide by zero",
			data: func() *array.Float64 {
				return arrow.NewFloat([]float64{1, 1, 1}, nil)
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
				b.Append(2)
				b.AppendNull()
				b.Append(2)
				b.AppendNull()
				b.Append(3)
				return b.NewFloat64Array()
			},
			want: 0.7071067811865475,
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
				new(universe.SkewAgg),
				tc.data(),
				tc.want,
			)
		})
	}
}

func BenchmarkSkew(b *testing.B) {
	data := arrow.NewFloat(NormalData, &memory.Allocator{})
	executetest.AggFuncBenchmarkHelper(
		b,
		new(universe.SkewAgg),
		data,
		-0.0019606823191321435,
	)
}
