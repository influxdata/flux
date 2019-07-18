package universe_test

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/querytest"
	"github.com/influxdata/flux/stdlib/universe"
	"testing"
)

func TestCMOOperation_Marshaling(t *testing.T) {
	data := []byte(`{"id":"chandeMomentumOscillator","kind":"chandeMomentumOscillator","spec":{"n":1}}`)
	op := &flux.Operation{
		ID: "chandeMomentumOscillator",
		Spec: &universe.ChandeMomentumOscillatorOpSpec{
			N: 1,
		},
	}
	querytest.OperationMarshalingTestHelper(t, data, op)
}

func TestCMO_PassThrough(t *testing.T) {
	executetest.TransformationPassThroughTestHelper(t, func(d execute.Dataset, c execute.TableBuilderCache) execute.Transformation {
		s := universe.NewChandeMomentumOscillatorTransformation(
			d,
			c,
			&universe.ChandeMomentumOscillatorProcedureSpec{},
		)
		return s
	})
}

func TestCMO_Process(t *testing.T) {
	testCases := []struct {
		name string
		spec *universe.ChandeMomentumOscillatorProcedureSpec
		data []flux.Table
		want []*executetest.Table
	}{
		{
			name: "basic output",
			spec: &universe.ChandeMomentumOscillatorProcedureSpec{
				N: 10,
				Columns: []string{"_value"},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
				},
				Data: [][]interface{}{
					{execute.Time(1)},
					{execute.Time(2)},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "elapsed", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{execute.Time(2), int64(execute.Time(2) - execute.Time(1))},
				},
			}},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			executetest.ProcessTestHelper(
				t,
				tc.data,
				tc.want,
				nil,
				func(d execute.Dataset, c execute.TableBuilderCache) execute.Transformation {
					return universe.NewChandeMomentumOscillatorTransformation(d, c, tc.spec)
				},
			)
		})
	}
}
