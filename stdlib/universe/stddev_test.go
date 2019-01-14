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
		data []float64
		want interface{}
	}{
		{
			name: "zero",
			data: []float64{1, 1, 1},
			want: 0.0,
		},
		{
			name: "nonzero",
			data: []float64{1, 2, 3},
			want: 1.0,
		},
		{
			name: "NaN",
			data: []float64{1},
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
				new(universe.StddevAgg),
				arrow.NewFloat(tc.data, nil),
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
