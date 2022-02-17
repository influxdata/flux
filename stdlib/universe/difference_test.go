package universe_test

import (
	"testing"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/querytest"
	"github.com/influxdata/flux/stdlib/universe"
	"github.com/influxdata/flux/values"
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
		name    string
		spec    *universe.DifferenceProcedureSpec
		data    []flux.Table
		want    []*executetest.Table
		wantErr error
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
			name: "non-supported string type in difference column",
			spec: &universe.DifferenceProcedureSpec{
				Columns: []string{execute.DefaultValueColLabel},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(1), "a"},
					{execute.Time(2), "b"},
				},
			}},
			wantErr: errors.New(codes.Invalid, `difference does not support column "_value" of type "string"`),
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
		{
			name: "with multiple tables",
			spec: &universe.DifferenceProcedureSpec{
				Columns: []string{"a", "b"},
			},
			data: []flux.Table{&executetest.Table{
				GroupKey: execute.NewGroupKey(
					[]flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
					},
					[]values.Value{values.NewTime(execute.Time(0))},
				),
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "a", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(0), 2.0},
					{execute.Time(1), 6.0},
				},
			},
				&executetest.Table{
					GroupKey: execute.NewGroupKey(
						[]flux.ColMeta{
							{Label: "_time", Type: flux.TTime},
						},
						[]values.Value{values.NewTime(execute.Time(2))},
					),
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "b", Type: flux.TInt},
					},
					Data: [][]interface{}{
						{execute.Time(2), int64(1)},
						{execute.Time(3), int64(2)},
					},
				},
			},
			want: []*executetest.Table{{
				GroupKey: execute.NewGroupKey(
					[]flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
					},
					[]values.Value{values.NewTime(execute.Time(0))},
				),
				KeyCols:   []string{"_time"},
				KeyValues: []interface{}{execute.Time(0)},
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "a", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 4.0},
				},
			},
				{
					GroupKey: execute.NewGroupKey(
						[]flux.ColMeta{
							{Label: "_time", Type: flux.TTime},
						},
						[]values.Value{values.NewTime(execute.Time(2))},
					),
					KeyCols:   []string{"_time"},
					KeyValues: []interface{}{execute.Time(2)},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "b", Type: flux.TInt},
					},
					Data: [][]interface{}{
						{execute.Time(3), int64(1)},
					},
				},
			},
		},
		{
			name: "float with tags and keepFirst",
			spec: &universe.DifferenceProcedureSpec{
				Columns:   []string{execute.DefaultValueColLabel},
				KeepFirst: true,
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
					{execute.Time(1), nil, "a"},
					{execute.Time(2), -1.0, "b"},
				},
			}},
		},
		{
			name: "float with tags and keepFirst and chunks",
			spec: &universe.DifferenceProcedureSpec{
				Columns:   []string{execute.DefaultValueColLabel},
				KeepFirst: true,
			},
			data: []flux.Table{&executetest.RowWiseTable{
				Table: &executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "t", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 2.0, "a"},
						{execute.Time(2), 1.0, "b"},
					},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "t", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(1), nil, "a"},
					{execute.Time(2), -1.0, "b"},
				},
			}},
		},
		{
			name: "int with tags and keepFirst",
			spec: &universe.DifferenceProcedureSpec{
				Columns:   []string{execute.DefaultValueColLabel},
				KeepFirst: true,
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TInt},
					{Label: "t", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(1), int64(2), "a"},
					{execute.Time(2), int64(1), "b"},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TInt},
					{Label: "t", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(1), nil, "a"},
					{execute.Time(2), int64(-1), "b"},
				},
			}},
		},
		{
			name: "int with tags and keepFirst and chunks",
			spec: &universe.DifferenceProcedureSpec{
				Columns:   []string{execute.DefaultValueColLabel},
				KeepFirst: true,
			},
			data: []flux.Table{&executetest.RowWiseTable{
				Table: &executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TInt},
						{Label: "t", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), int64(2), "a"},
						{execute.Time(2), int64(1), "b"},
					},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TInt},
					{Label: "t", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(1), nil, "a"},
					{execute.Time(2), int64(-1), "b"},
				},
			}},
		},
		{
			name: "uint with tags and keepFirst",
			spec: &universe.DifferenceProcedureSpec{
				Columns:   []string{execute.DefaultValueColLabel},
				KeepFirst: true,
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TUInt},
					{Label: "t", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(1), uint64(2), "a"},
					{execute.Time(2), uint64(1), "b"},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TInt},
					{Label: "t", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(1), nil, "a"},
					{execute.Time(2), int64(-1), "b"},
				},
			}},
		},
		{
			name: "uint with tags and keepFirst and chunks",
			spec: &universe.DifferenceProcedureSpec{
				Columns:   []string{execute.DefaultValueColLabel},
				KeepFirst: true,
			},
			data: []flux.Table{&executetest.RowWiseTable{
				Table: &executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TUInt},
						{Label: "t", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), uint64(2), "a"},
						{execute.Time(2), uint64(1), "b"},
					},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TInt},
					{Label: "t", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(1), nil, "a"},
					{execute.Time(2), int64(-1), "b"},
				},
			}},
		},
		{
			name: "float rowwise",
			spec: &universe.DifferenceProcedureSpec{
				Columns: []string{execute.DefaultValueColLabel},
			},
			data: []flux.Table{&executetest.RowWiseTable{
				Table: &executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), 2.0},
						{execute.Time(2), 1.0},
					},
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
			name: "int rowwise",
			spec: &universe.DifferenceProcedureSpec{
				Columns: []string{execute.DefaultValueColLabel},
			},
			data: []flux.Table{&executetest.RowWiseTable{
				Table: &executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TInt},
					},
					Data: [][]interface{}{
						{execute.Time(1), int64(20)},
						{execute.Time(2), int64(10)},
					},
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
			name: "uint rowwise",
			spec: &universe.DifferenceProcedureSpec{
				Columns: []string{execute.DefaultValueColLabel},
			},
			data: []flux.Table{&executetest.RowWiseTable{
				Table: &executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TUInt},
					},
					Data: [][]interface{}{
						{execute.Time(1), uint64(10)},
						{execute.Time(2), uint64(20)},
					},
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
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			executetest.ProcessTestHelper(
				t,
				tc.data,
				tc.want,
				tc.wantErr,
				func(d execute.Dataset, c execute.TableBuilderCache) execute.Transformation {
					return universe.NewDifferenceTransformation(d, c, tc.spec)
				},
			)
		})
	}
}

func TestDifference_Process_With_NonNegative_KeepFirst_InitialZero(t *testing.T) {
	testCases := []struct {
		name    string
		spec    *universe.DifferenceProcedureSpec
		data    []flux.Table
		want    []*executetest.Table
		wantErr error
	}{
		// keepFirst: false, initialZero: false
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
					{execute.Time(2), int64(30)},
					{execute.Time(3), int64(50)},
					{execute.Time(4), int64(30)},
					{execute.Time(5), int64(40)},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{execute.Time(2), int64(10)},
					{execute.Time(3), int64(20)},
					{execute.Time(4), nil},
					{execute.Time(5), int64(10)},
				},
			}},
		},
		// keepFirst: true, initialZero: false
		{
			name: "uint with non negative",
			spec: &universe.DifferenceProcedureSpec{
				Columns:     []string{execute.DefaultValueColLabel},
				NonNegative: true,
				KeepFirst:   true,
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TUInt},
				},
				Data: [][]interface{}{
					{execute.Time(1), uint64(20)},
					{execute.Time(2), uint64(30)},
					{execute.Time(3), uint64(50)},
					{execute.Time(4), uint64(30)},
					{execute.Time(5), uint64(40)},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{execute.Time(1), nil},
					{execute.Time(2), int64(10)},
					{execute.Time(3), int64(20)},
					{execute.Time(4), nil},
					{execute.Time(5), int64(10)},
				},
			}},
		},
		// keepFirst: false, initialZero: true
		{
			name: "non negative one table",
			spec: &universe.DifferenceProcedureSpec{
				Columns:     []string{execute.DefaultValueColLabel},
				NonNegative: true,
				InitialZero: true,
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 20.0},
					{execute.Time(2), 30.0},
					{execute.Time(3), 50.0},
					{execute.Time(4), 30.0},
					{execute.Time(5), -40.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(2), 10.0},
					{execute.Time(3), 20.0},
					{execute.Time(4), 30.0},
					{execute.Time(5), nil},
				},
			}},
		},
		// keepFirst: true, initialZero: true
		{
			name: "float with KeepFirst and InitialZero",
			spec: &universe.DifferenceProcedureSpec{
				Columns:     []string{execute.DefaultValueColLabel},
				NonNegative: true,
				KeepFirst:   true,
				InitialZero: true,
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "t", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(1), 20.0, "a"},
					{execute.Time(2), 30.0, "b"},
					{execute.Time(3), 50.0, "c"},
					{execute.Time(4), 30.0, "d"},
					{execute.Time(5), 40.0, "e"},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "t", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(1), float64(0), "a"},
					{execute.Time(2), 10.0, "b"},
					{execute.Time(3), 20.0, "c"},
					{execute.Time(4), 30.0, "d"},
					{execute.Time(5), 10.0, "e"},
				},
			}},
		},
		{
			name: "with null non negative and InitialZero",
			spec: &universe.DifferenceProcedureSpec{
				Columns:     []string{"a", "b", "c"},
				NonNegative: true,
				InitialZero: true,
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
					{execute.Time(2), 4.0, 0.0, 2.0},
					{execute.Time(3), 6.0, 8.0, 2.0},
					{execute.Time(4), nil, nil, 1.0},
				},
			}},
		},
		{
			name: "float with tags and InitialZero",
			spec: &universe.DifferenceProcedureSpec{
				Columns:     []string{"a", "b", "c"},
				NonNegative: true,
				InitialZero: true,
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
					{execute.Time(2), 4.0, 0.0, 2.0},
					{execute.Time(3), 6.0, 8.0, 2.0},
					{execute.Time(4), nil, nil, 1.0},
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
				tc.wantErr,
				func(d execute.Dataset, c execute.TableBuilderCache) execute.Transformation {
					return universe.NewDifferenceTransformation(d, c, tc.spec)
				},
			)
		})
	}
}
