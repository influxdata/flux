package functions_test

import (
	"testing"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/functions"
	"github.com/influxdata/flux/querytest"
)

func TestSumOperation_Marshaling(t *testing.T) {
	data := []byte(`{"id":"sum","kind":"sum"}`)
	op := &flux.Operation{
		ID:   "sum",
		Spec: &functions.SumOpSpec{},
	}

	querytest.OperationMarshalingTestHelper(t, data, op)
}

func TestSum_Process(t *testing.T) {
	executetest.AggFuncTestHelper(t,
		new(functions.SumAgg),
		[]float64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		float64(45),
	)
}

func BenchmarkSum(b *testing.B) {
	executetest.AggFuncBenchmarkHelper(
		b,
		new(functions.SumAgg),
		NormalData,
		10000816.96729983,
	)
}

