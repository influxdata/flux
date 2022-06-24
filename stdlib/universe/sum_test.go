package universe_test

import (
	"testing"

	"github.com/mvn-trinhnguyen2-dn/flux"
	"github.com/mvn-trinhnguyen2-dn/flux/array"
	"github.com/mvn-trinhnguyen2-dn/flux/arrow"
	"github.com/mvn-trinhnguyen2-dn/flux/execute/executetest"
	"github.com/mvn-trinhnguyen2-dn/flux/memory"
	"github.com/mvn-trinhnguyen2-dn/flux/querytest"
	"github.com/mvn-trinhnguyen2-dn/flux/stdlib/universe"
)

func TestSumOperation_Marshaling(t *testing.T) {
	data := []byte(`{"id":"sum","kind":"sum"}`)
	op := &flux.Operation{
		ID:   "sum",
		Spec: &universe.SumOpSpec{},
	}

	querytest.OperationMarshalingTestHelper(t, data, op)
}

func TestSum_Process(t *testing.T) {
	testCases := []struct {
		name string
		data func() *array.Float
		want interface{}
	}{
		{
			name: "zero",
			data: func() *array.Float {
				return arrow.NewFloat([]float64{0, 0, 0}, nil)
			},
			want: 0.0,
		},
		{
			name: "nonzero",
			data: func() *array.Float {
				return arrow.NewFloat([]float64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, nil)
			},
			want: 45.0,
		},
		{
			name: "empty",
			data: func() *array.Float {
				return arrow.NewFloat(nil, nil)
			},
			want: nil,
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
			want: 34.0,
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
			want: nil,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			data := tc.data()
			defer data.Release()

			executetest.AggFuncTestHelper(
				t,
				new(universe.SumAgg),
				data,
				tc.want,
			)
		})
	}
}

func BenchmarkSum(b *testing.B) {
	data := arrow.NewFloat(NormalData, &memory.ResourceAllocator{})
	executetest.AggFuncBenchmarkHelper(
		b,
		new(universe.SumAgg),
		data,
		9998472.67384332,
	)
}
