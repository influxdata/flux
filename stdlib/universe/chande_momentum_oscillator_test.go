package universe_test

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/querytest"
	"github.com/influxdata/flux/stdlib/universe"
	"testing"
)

func TestChandeMomentumOscillatorOperation_Marshaling(t *testing.T) {
	data := []byte(`{"id":"chandeMomentumOscillator","kind":"chandeMomentumOscillator","spec":{"n":1}}`)
	op := &flux.Operation{
		ID: "chandeMomentumOscillator",
		Spec: &universe.ChandeMomentumOscillatorOpSpec{
			N: 1,
		},
	}
	querytest.OperationMarshalingTestHelper(t, data, op)
}

func TestChandeMomentumOscillator_PassThrough(t *testing.T) {
	executetest.TransformationPassThroughTestHelper(t, func(d execute.Dataset, c execute.TableBuilderCache) execute.Transformation {
		s := universe.NewChandeMomentumOscillatorTransformation(
			d,
			c,
			&universe.ChandeMomentumOscillatorProcedureSpec{},
		)
		return s
	})
}

func TestChandeMomentumOscillator_Process(t *testing.T) {
	testCases := []struct {
		name string
		spec *universe.ChandeMomentumOscillatorProcedureSpec
		data []flux.Table
		want []*executetest.Table
	}{
		{
			name: "flux.TInt input",
			spec: &universe.ChandeMomentumOscillatorProcedureSpec{
				N:       10,
				Columns: []string{"_value"},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_value", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{int64(1)},
					{int64(2)},
					{int64(3)},
					{int64(4)},
					{int64(5)},
					{int64(6)},
					{int64(7)},
					{int64(8)},
					{int64(9)},
					{int64(10)},
					{int64(11)},
					{int64(12)},
					{int64(13)},
					{int64(14)},
					{int64(15)},
					{int64(14)},
					{int64(13)},
					{int64(12)},
					{int64(11)},
					{int64(10)},
					{int64(9)},
					{int64(8)},
					{int64(7)},
					{int64(6)},
					{int64(5)},
					{int64(4)},
					{int64(3)},
					{int64(2)},
					{int64(1)},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{100.0},
					{100.0},
					{100.0},
					{100.0},
					{100.0},
					{80.0},
					{60.0},
					{40.0},
					{20.0},
					{0.0},
					{-20.0},
					{-40.0},
					{-60.0},
					{-80.0},
					{-100.0},
					{-100.0},
					{-100.0},
					{-100.0},
					{-100.0},
				},
			}},
		},
		{
			name: "flux.TTime & flux.TInt input",
			spec: &universe.ChandeMomentumOscillatorProcedureSpec{
				N:       1,
				Columns: []string{"_value"},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{execute.Time(12), int64(1)},
					{execute.Time(14), int64(2)},
					{execute.Time(14), int64(4)},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(14), 100.0},
					{execute.Time(14), 100.0},
				},
			}},
		},
		{
			name: "flux.TTime & flux.TUInt input",
			spec: &universe.ChandeMomentumOscillatorProcedureSpec{
				N:       1,
				Columns: []string{"_value"},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TUInt},
				},
				Data: [][]interface{}{
					{execute.Time(12), uint64(1)},
					{execute.Time(14), uint64(2)},
					{execute.Time(14), uint64(4)},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(14), 100.0},
					{execute.Time(14), 100.0},
				},
			}},
		},
		{
			name: "flux.TTime & flux.TFloat input",
			spec: &universe.ChandeMomentumOscillatorProcedureSpec{
				N:       1,
				Columns: []string{"_value"},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(12), 1.0},
					{execute.Time(14), 2.0},
					{execute.Time(14), 4.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(14), 100.0},
					{execute.Time(14), 100.0},
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
