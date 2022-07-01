package universe_test

import (
	"testing"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/querytest"
	"github.com/influxdata/flux/stdlib/universe"
)

func TestUniqueOperation_Marshaling(t *testing.T) {
	data := []byte(`{"id":"unique","kind":"unique","spec":{"column":"_value"}}`)
	op := &flux.Operation{
		ID: "unique",
		Spec: &universe.UniqueOpSpec{
			Column: "_value",
		},
	}
	querytest.OperationMarshalingTestHelper(t, data, op)
}

func TestUnique_PassThrough(t *testing.T) {
	executetest.TransformationPassThroughTestHelper(t, func(d execute.Dataset, c execute.TableBuilderCache) execute.Transformation {
		s := universe.NewUniqueTransformation(
			d,
			c,
			&universe.UniqueProcedureSpec{
				Column: "_value",
			},
		)
		return s
	})
}

func TestUnique_Process(t *testing.T) {
	testCases := []struct {
		name string
		spec *universe.UniqueProcedureSpec
		data []flux.Table
		want []*executetest.Table
	}{
		{
			name: "one table",
			spec: &universe.UniqueProcedureSpec{
				Column: "_value",
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 2.0},
					{execute.Time(2), 1.0},
					{execute.Time(3), 3.0},
					{execute.Time(4), 1.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 2.0},
					{execute.Time(2), 1.0},
					{execute.Time(3), 3.0},
				},
			}},
		},
		{
			name: "unique tag",
			spec: &universe.UniqueProcedureSpec{
				Column: "t1",
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "t1", Type: flux.TString},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), "a", 2.0},
					{execute.Time(2), "a", 1.0},
					{execute.Time(3), "b", 3.0},
					{execute.Time(4), "c", 1.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "t1", Type: flux.TString},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), "a", 2.0},
					{execute.Time(3), "b", 3.0},
					{execute.Time(4), "c", 1.0},
				},
			}},
		},
		{
			name: "unique times",
			spec: &universe.UniqueProcedureSpec{
				Column: "_time",
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "t1", Type: flux.TString},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), "a", 2.0},
					{execute.Time(2), "a", 1.0},
					{execute.Time(3), "b", 3.0},
					{execute.Time(3), "c", 1.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "t1", Type: flux.TString},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), "a", 2.0},
					{execute.Time(2), "a", 1.0},
					{execute.Time(3), "b", 3.0},
				},
			}},
		},
		{
			name: "with nulls",
			spec: &universe.UniqueProcedureSpec{
				Column: "_value",
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 2.0},
					{nil, 1.0},
					{execute.Time(3), 3.0},
					{execute.Time(4), 1.0},
					{execute.Time(5), 2.0},
					{execute.Time(6), 1.0},
					{execute.Time(7), nil},
					{execute.Time(8), nil},
					{execute.Time(9), 3.0},
					{execute.Time(10), nil},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 2.0},
					{nil, 1.0},
					{execute.Time(3), 3.0},
					{execute.Time(7), nil},
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
					return universe.NewUniqueTransformation(d, c, tc.spec)
				},
			)
		})
	}
}
