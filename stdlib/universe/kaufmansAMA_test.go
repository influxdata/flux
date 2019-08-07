package universe_test

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/querytest"
	"github.com/influxdata/flux/stdlib/universe"
	"testing"
)

func TestKamaOperation_Marshaling(t *testing.T) {
	data := []byte(`{"id":"kaufmansAMA","kind":"kaufmansAMA","spec":{"n":1}}`)
	op := &flux.Operation{
		ID: "kaufmansAMA",
		Spec: &universe.KamaOpSpec{
			N: 1,
		},
	}
	querytest.OperationMarshalingTestHelper(t, data, op)
}

func TestKama_PassThrough(t *testing.T) {
	executetest.TransformationPassThroughTestHelper(t, func(d execute.Dataset, c execute.TableBuilderCache) execute.Transformation {
		s := universe.NewkamaTransformation(
			d,
			c,
			&universe.KamaProcedureSpec{},
		)
		return s
	})
}

func TestKama_Process(t *testing.T) {
	testCases := []struct {
		name string
		spec *universe.KamaProcedureSpec
		data []flux.Table
		want []*executetest.Table
	}{
		{
			name: "flux.TInt input",
			spec: &universe.KamaProcedureSpec{
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
					{10.444444444444445},
					{11.135802469135802},
					{11.964334705075446},
					{12.869074836153025},
					{13.81615268675168},
					{13.871008014588556},
					{13.71308456353558},
					{13.553331356741122},
					{13.46599437575161},
					{13.4515677602438},
					{13.29930139347417},
					{12.805116570729282},
					{11.752584300922965},
					{10.036160535131101},
					{7.797866963961722},
					{6.109926091089845},
					{4.727736717272135},
					{3.515409287373408},
					{2.3974496040963373},
				},
			}},
		},
		{
			name: "flux.TTime & flux.TInt input",
			spec: &universe.KamaProcedureSpec{
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
					{execute.Time(14), 1.4444444444444446},
					{execute.Time(14), 2.5802469135802473},
				},
			}},
		},
		{
			name: "flux.TTime & flux.TUInt input",
			spec: &universe.KamaProcedureSpec{
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
					{execute.Time(14), 1.4444444444444446},
					{execute.Time(14), 2.5802469135802473},
				},
			}},
		},
		{
			name: "flux.TTime & flux.TFloat input",
			spec: &universe.KamaProcedureSpec{
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
					{execute.Time(14), 1.4444444444444446},
					{execute.Time(14), 2.5802469135802473},
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
					return universe.NewkamaTransformation(d, c, tc.spec)
				},
			)
		})
	}
}
