package universe_test

import (
	"testing"

	"github.com/apache/arrow/go/arrow/array"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/querytest"
	"github.com/influxdata/flux/stdlib/universe"
)

func TestMadOperation_Marshaling(t *testing.T) {
	data := []byte(`{"id":"mad","kind":"mad"}`)
	op := &flux.Operation{
		ID:   "mad",
		Spec: &universe.MadOpSpec{},
	}

	querytest.OperationMarshalingTestHelper(t, data, op)
}

func TestMad_Process(t *testing.T) {
	testCases := []struct {
		name string
		data func() *array.Float64
		want interface{}
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
				return arrow.NewFloat([]float64{55.0, 51.0, 12.0}, nil)
			},
			want: 4.0,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			data := tc.data()
			defer data.Release()

			executetest.AggFuncTestHelper(
				t,
				new(universe.MadAgg),
				data,
				tc.want,
			)
		})
	}
}

// func BenchmarkMad(b *testing.B) {
// 	data := arrow.NewFloat(NormalData, &memory.Allocator{})
// 	executetest.AggFuncBenchmarkHelper(
// 		b,
// 		new(universe.MadAgg),
// 		data,
// 		10000816.96729983,
// 	)
// }
