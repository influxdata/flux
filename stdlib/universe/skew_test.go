package universe_test

import (
	"math"
	"testing"

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
		data []float64
		want interface{}
	}{
		{
			name: "zero",
			data: []float64{1, 2, 3},
			want: 0.0,
		},
		{
			name: "nonzero",
			data: []float64{2, 2, 3},
			want: 0.7071067811865475,
		},
		{
			name: "nonzero",
			data: []float64{2, 2, 3, 4},
			want: 0.49338220021815854,
		},
		{
			name: "NaN short",
			data: []float64{1},
			want: math.NaN(),
		},
		{
			name: "NaN divide by zero",
			data: []float64{1, 1, 1},
			want: math.NaN(),
		},
		{
			name: "empty",
			data: []float64{},
			want: nil,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			executetest.AggFuncTestHelper(
				t,
				new(universe.SkewAgg),
				arrow.NewFloat(tc.data, nil),
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
		0.0032200673020400935,
	)
}
