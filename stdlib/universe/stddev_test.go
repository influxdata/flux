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
	data := []byte(`{"id":"stddev","kind":"stddev","spec":{"mode":"sample"}}`)
	op := &flux.Operation{
		ID:   "stddev",
		Spec: &universe.StddevOpSpec{Mode: "sample"},
	}

	querytest.OperationMarshalingTestHelper(t, data, op)
}

func TestStddev_Process(t *testing.T) {
	testCases := []struct {
		name        string
		data        func() *array.Float64
		wantForMode map[string]interface{}
	}{
		{
			name: "zero",
			data: func() *array.Float64 {
				return arrow.NewFloat([]float64{1, 1, 1}, nil)
			},
			wantForMode: map[string]interface{}{
				"sample":     0.0,
				"population": 0.0,
			},
		},
		{
			name: "nonzero",
			data: func() *array.Float64 {
				return arrow.NewFloat([]float64{1, 2, 3}, nil)
			},
			wantForMode: map[string]interface{}{
				"sample":     1.0,
				"population": 0.816496580927726,
			},
		},
		{
			name: "NaN",
			data: func() *array.Float64 {
				return arrow.NewFloat([]float64{1}, nil)
			},
			wantForMode: map[string]interface{}{
				"sample":     math.NaN(),
				"population": 0.0,
			},
		},
		{
			name: "empty",
			data: func() *array.Float64 {
				return arrow.NewFloat(nil, nil)
			},
			wantForMode: map[string]interface{}{
				"sample":     nil,
				"population": nil,
			},
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
			wantForMode: map[string]interface{}{
				"sample":     1.0,
				"population": 0.816496580927726,
			},
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
			wantForMode: map[string]interface{}{
				"sample":     nil,
				"population": nil,
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		for mode, want := range tc.wantForMode {
			t.Run(tc.name+"_"+mode, func(t *testing.T) {
				executetest.AggFuncTestHelper(
					t,
					&universe.StddevAgg{Mode: mode},
					tc.data(),
					want,
				)
			})
		}
	}
}

func BenchmarkStddev(b *testing.B) {
	data := arrow.NewFloat(NormalData, &memory.Allocator{})
	executetest.AggFuncBenchmarkHelper(
		b,
		&universe.StddevAgg{Mode: "sample"},
		data,
		2.998926113076968,
	)
}
