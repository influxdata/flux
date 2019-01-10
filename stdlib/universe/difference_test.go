package universe_test

import (
	"testing"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/querytest"
	"github.com/influxdata/flux/stdlib/universe"
)

func TestDifferenceOperation_Marshaling(t *testing.T) {
	data := []byte(`{"id":"difference","kind":"difference","spec":{"nonNegative":true}}`)
	op := &flux.Operation{
		ID: "difference",
		Spec: &universe.DifferenceOpSpec{
			NonNegative: true,
		},
	}
	querytest.OperationMarshalingTestHelper(t, data, op)
}

func TestDifference_PassThrough(t *testing.T) {
	executetest.TransformationPassThroughTestHelper(t, func(d execute.Dataset, c execute.TableBuilderCache) execute.Transformation {
		s := universe.NewDifferenceTransformation(
			d,
			c,
			&universe.DifferenceProcedureSpec{},
		)
		return s
	})
}

func TestDifference_Process(t *testing.T) {
	testCases := []struct {
		name string
		spec *universe.DifferenceProcedureSpec
		data []flux.Table
		want []*executetest.Table
	}{
		{
			name: "float",
			spec: &universe.DifferenceProcedureSpec{
				Columns: []string{execute.DefaultValueColLabel},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 2.0},
					{execute.Time(2), 1.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(2), -1.0},
				},
			}},
		},
		{
			name: "int",
			spec: &universe.DifferenceProcedureSpec{
				Columns: []string{execute.DefaultValueColLabel},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{execute.Time(1), int64(20)},
					{execute.Time(2), int64(10)},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{execute.Time(2), int64(-10)},
				},
			}},
		},
		{
			name: "int non negative",
			spec: &universe.DifferenceProcedureSpec{
				Columns:     []string{execute.DefaultValueColLabel},
				NonNegative: true,
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{execute.Time(1), int64(20)},
					{execute.Time(2), int64(10)},
					{execute.Time(3), int64(20)},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{execute.Time(2), nil},
					{execute.Time(3), int64(10)},
				},
			}},
		},
		{
			name: "uint",
			spec: &universe.DifferenceProcedureSpec{
				Columns: []string{execute.DefaultValueColLabel},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TUInt},
				},
				Data: [][]interface{}{
					{execute.Time(1), uint64(10)},
					{execute.Time(2), uint64(20)},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{execute.Time(2), int64(10)},
				},
			}},
		},
		{
			name: "uint with negative result",
			spec: &universe.DifferenceProcedureSpec{
				Columns: []string{execute.DefaultValueColLabel},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TUInt},
				},
				Data: [][]interface{}{
					{execute.Time(1), uint64(20)},
					{execute.Time(2), uint64(10)},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{execute.Time(2), int64(-10)},
				},
			}},
		},
		{
			name: "uint with non negative",
			spec: &universe.DifferenceProcedureSpec{
				Columns:     []string{execute.DefaultValueColLabel},
				NonNegative: true,
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TUInt},
				},
				Data: [][]interface{}{
					{execute.Time(1), uint64(20)},
					{execute.Time(2), uint64(10)},
					{execute.Time(3), uint64(20)},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{execute.Time(2), nil},
					{execute.Time(3), int64(10)},
				},
			}},
		},
		{
			name: "non negative one table",
			spec: &universe.DifferenceProcedureSpec{
				Columns:     []string{execute.DefaultValueColLabel},
				NonNegative: true,
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 2.0},
					{execute.Time(2), 1.0},
					{execute.Time(3), 2.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(2), nil},
					{execute.Time(3), 1.0},
				},
			}},
		},
		{
			name: "float with tags",
			spec: &universe.DifferenceProcedureSpec{
				Columns: []string{execute.DefaultValueColLabel},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "t", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(1), 2.0, "a"},
					{execute.Time(2), 1.0, "b"},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "t", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(2), -1.0, "b"},
				},
			}},
		},
		{
			name: "float with multiple values",
			spec: &universe.DifferenceProcedureSpec{
				Columns: []string{"x", "y"},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "x", Type: flux.TFloat},
					{Label: "y", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 2.0, 20.0},
					{execute.Time(2), 1.0, 10.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "x", Type: flux.TFloat},
					{Label: "y", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(2), -1.0, -10.0},
				},
			}},
		},
		{
			name: "float non negative with multiple values",
			spec: &universe.DifferenceProcedureSpec{
				Columns:     []string{"x", "y"},
				NonNegative: true,
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "x", Type: flux.TFloat},
					{Label: "y", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 2.0, 20.0},
					{execute.Time(2), 1.0, 10.0},
					{execute.Time(3), 2.0, 0.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "x", Type: flux.TFloat},
					{Label: "y", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(2), nil, nil},
					{execute.Time(3), 1.0, nil},
				},
			}},
		},
		{
			name: "float no values",
			spec: &universe.DifferenceProcedureSpec{
				Columns:     []string{"x", "y"},
				NonNegative: true,
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "x", Type: flux.TFloat},
					{Label: "y", Type: flux.TFloat},
				},
				Data: [][]interface{}{},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "x", Type: flux.TFloat},
					{Label: "y", Type: flux.TFloat},
				},
				Data: [][]interface{}(nil),
			}},
		},
		{
			name: "float single value",
			spec: &universe.DifferenceProcedureSpec{
				Columns:     []string{"x", "y"},
				NonNegative: true,
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "x", Type: flux.TFloat},
					{Label: "y", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(3), 10.0, 20.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "x", Type: flux.TFloat},
					{Label: "y", Type: flux.TFloat},
				},
				Data: [][]interface{}(nil),
			}},
		},
		{
			name: "with null",
			spec: &universe.DifferenceProcedureSpec{
				Columns: []string{"a", "b", "c"},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "a", Type: flux.TFloat},
					{Label: "b", Type: flux.TFloat},
					{Label: "c", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(0), nil, 1.0, 2.0},
					{execute.Time(1), 6.0, 2.0, nil},
					{execute.Time(2), 4.0, 2.0, 4.0},
					{execute.Time(3), 10.0, 10.0, 2.0},
					{execute.Time(4), nil, nil, 1.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "a", Type: flux.TFloat},
					{Label: "b", Type: flux.TFloat},
					{Label: "c", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), nil, 1.0, nil},
					{execute.Time(2), -2.0, 0.0, 2.0},
					{execute.Time(3), 6.0, 8.0, -2.0},
					{execute.Time(4), nil, nil, -1.0},
				},
			}},
		},
		{
			name: "with null non negative",
			spec: &universe.DifferenceProcedureSpec{
				Columns:     []string{"a", "b", "c"},
				NonNegative: true,
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "a", Type: flux.TFloat},
					{Label: "b", Type: flux.TFloat},
					{Label: "c", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(0), nil, 1.0, 2.0},
					{execute.Time(1), 6.0, 2.0, nil},
					{execute.Time(2), 4.0, 2.0, 4.0},
					{execute.Time(3), 10.0, 10.0, 2.0},
					{execute.Time(4), nil, nil, 1.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "a", Type: flux.TFloat},
					{Label: "b", Type: flux.TFloat},
					{Label: "c", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), nil, 1.0, nil},
					{execute.Time(2), nil, 0.0, 2.0},
					{execute.Time(3), 6.0, 8.0, nil},
					{execute.Time(4), nil, nil, nil},
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
					return universe.NewDifferenceTransformation(d, c, tc.spec)
				},
			)
		})
	}
}
