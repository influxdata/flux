package universe_test

import (
	"testing"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/stdlib/universe"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/querytest"
)

func TestSpreadOperation_Marshaling(t *testing.T) {
	data := []byte(`{"id":"spread","kind":"spread"}`)
	op := &flux.Operation{
		ID:   "spread",
		Spec: &universe.SpreadOpSpec{},
	}

	querytest.OperationMarshalingTestHelper(t, data, op)
}

func TestSpread_Process(t *testing.T) {
	agg := new(universe.SpreadAgg)
	executetest.AggFuncTestHelper(t,
		agg,
		arrow.NewFloat([]float64{
			0, 1, 2, 3, 4,
			5, 6, 7, 8, 9,
		}, nil),
		float64(9),
	)
}

func BenchmarkSpread(b *testing.B) {
	data := arrow.NewFloat(NormalData, &memory.Allocator{})
	executetest.AggFuncBenchmarkHelper(
		b,
		new(universe.SpreadAgg),
		data,
		28.227196461851847,
	)
}
