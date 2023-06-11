package executetest

import (
	"testing"

	"github.com/InfluxCommunity/flux"
	"github.com/InfluxCommunity/flux/array"
	"github.com/InfluxCommunity/flux/arrow"
	"github.com/InfluxCommunity/flux/execute"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// AggFuncTestHelper splits the data in half, runs Do over each split and compares
// the Value to want.
func AggFuncTestHelper(t *testing.T, agg execute.SimpleAggregate, data *array.Float, want interface{}) {
	t.Helper()

	// Call Do twice, since this is possible according to the interface.
	h := data.Len() / 2
	vf := agg.NewFloatAgg()

	d := arrow.FloatSlice(data, 0, h)
	vf.DoFloat(d)
	d.Release()
	if h < data.Len() {
		d := arrow.FloatSlice(data, h, data.Len())
		vf.DoFloat(d)
		d.Release()
	}

	var got interface{}
	if !vf.IsNull() {
		switch vf.Type() {
		case flux.TBool:
			got = vf.(execute.BoolValueFunc).ValueBool()
		case flux.TInt:
			got = vf.(execute.IntValueFunc).ValueInt()
		case flux.TUInt:
			got = vf.(execute.UIntValueFunc).ValueUInt()
		case flux.TFloat:
			got = vf.(execute.FloatValueFunc).ValueFloat()
		case flux.TString:
			got = vf.(execute.StringValueFunc).ValueString()
		}
	}

	if !cmp.Equal(want, got, cmpopts.EquateNaNs()) {
		t.Errorf("unexpected value -want/+got\n%s", cmp.Diff(want, got))
	}
}

// AggFuncBenchmarkHelper benchmarks the aggregate function over data and compares to wantValue
func AggFuncBenchmarkHelper(b *testing.B, agg execute.SimpleAggregate, data *array.Float, want interface{}) {
	b.Helper()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		vf := agg.NewFloatAgg()
		vf.DoFloat(data)
		var got interface{}
		switch vf.Type() {
		case flux.TBool:
			got = vf.(execute.BoolValueFunc).ValueBool()
		case flux.TInt:
			got = vf.(execute.IntValueFunc).ValueInt()
		case flux.TUInt:
			got = vf.(execute.UIntValueFunc).ValueUInt()
		case flux.TFloat:
			got = vf.(execute.FloatValueFunc).ValueFloat()
		case flux.TString:
			got = vf.(execute.StringValueFunc).ValueString()
		}
		if !cmp.Equal(want, got) {
			b.Errorf("unexpected value -want/+got\n%s", cmp.Diff(want, got))
		}
	}
}
