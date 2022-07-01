package universe_test

import (
	"testing"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/stdlib/universe"
)

func TestMode_Process(t *testing.T) {
	testCases := []struct {
		name string
		spec *universe.ModeProcedureSpec
		data []flux.Table
		want []*executetest.Table
	}{
		{
			name: "no group key strings",
			spec: &universe.ModeProcedureSpec{Column: "tag1"},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "tag0", Type: flux.TString},
						{Label: "tag1", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 2.0, "a", "b"},
						{execute.Time(2), 2.0, "a", "c"},
						{execute.Time(3), 2.0, "a", "b"},
						{execute.Time(4), 2.0, "a", "d"},
					},
				},
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_value", Type: flux.TString},
				},
				Data: [][]interface{}{
					{"b"},
				},
			}},
		},
		{
			name: "no group key strings with null",
			spec: &universe.ModeProcedureSpec{Column: "tag1"},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "tag0", Type: flux.TString},
						{Label: "tag1", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 2.0, "a", "b"},
						{execute.Time(2), 2.0, "a", "c"},
						{execute.Time(3), 2.0, "a", "b"},
						{execute.Time(4), 2.0, "a", "d"},
						{execute.Time(5), 2.0, "a", nil},
						{execute.Time(6), 2.0, "a", nil},
					},
				},
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_value", Type: flux.TString},
				},
				Data: [][]interface{}{
					{"b"},
				},
			}},
		},
		{
			name: "no group key ints",
			spec: &universe.ModeProcedureSpec{Column: "tag1"},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "tag0", Type: flux.TString},
						{Label: "tag1", Type: flux.TInt},
					},
					Data: [][]interface{}{
						{execute.Time(1), 2.0, "a", int64(3)},
						{execute.Time(2), 2.0, "a", int64(2)},
						{execute.Time(3), 2.0, "a", int64(2)},
						{execute.Time(4), 2.0, "a", int64(1)},
					},
				},
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_value", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{int64(2)},
				},
			}},
		},
		{
			name: "no group key ints with null",
			spec: &universe.ModeProcedureSpec{Column: "tag1"},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "tag0", Type: flux.TString},
						{Label: "tag1", Type: flux.TInt},
					},
					Data: [][]interface{}{
						{execute.Time(1), 2.0, "a", int64(3)},
						{execute.Time(2), 2.0, "a", int64(2)},
						{execute.Time(3), 2.0, "a", int64(2)},
						{execute.Time(4), 2.0, "a", int64(1)},
						{execute.Time(5), 2.0, "a", nil},
						{execute.Time(6), 2.0, "a", nil},
					},
				},
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_value", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{int64(2)},
				},
			}},
		},
		{
			name: "no group key ints with more nulls",
			spec: &universe.ModeProcedureSpec{Column: "tag1"},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "tag0", Type: flux.TString},
						{Label: "tag1", Type: flux.TInt},
					},
					Data: [][]interface{}{
						{execute.Time(1), 2.0, "a", int64(3)},
						{execute.Time(2), 2.0, "a", int64(2)},
						{execute.Time(3), 2.0, "a", int64(2)},
						{execute.Time(4), 2.0, "a", int64(1)},
						{execute.Time(5), 2.0, "a", nil},
						{execute.Time(6), 2.0, "a", nil},
						{execute.Time(7), 2.0, "a", nil},
					},
				},
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_value", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{int64(2)},
				},
			}},
		},
		{
			name: "no group key floats",
			spec: &universe.ModeProcedureSpec{Column: "tag1"},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "tag0", Type: flux.TString},
						{Label: "tag1", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), 2.0, "a", float64(1.0)},
						{execute.Time(2), 2.0, "a", float64(3.0)},
						{execute.Time(3), 2.0, "a", float64(2.0)},
						{execute.Time(4), 2.0, "a", float64(1.0)},
					},
				},
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{float64(1.0)},
				},
			}},
		},
		{
			name: "no group key floats with null",
			spec: &universe.ModeProcedureSpec{Column: "tag1"},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "tag0", Type: flux.TString},
						{Label: "tag1", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), 2.0, "a", float64(1.0)},
						{execute.Time(2), 2.0, "a", float64(3.0)},
						{execute.Time(3), 2.0, "a", float64(2.0)},
						{execute.Time(4), 2.0, "a", float64(1.0)},
						{execute.Time(5), 2.0, "a", nil},
						{execute.Time(6), 2.0, "a", nil},
					},
				},
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{float64(1.0)},
				},
			}},
		},
		{
			name: "no group key uints",
			spec: &universe.ModeProcedureSpec{Column: "tag1"},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "tag0", Type: flux.TString},
						{Label: "tag1", Type: flux.TUInt},
					},
					Data: [][]interface{}{
						{execute.Time(1), 2.0, "a", uint64(3)},
						{execute.Time(2), 2.0, "a", uint64(2)},
						{execute.Time(3), 2.0, "a", uint64(2)},
						{execute.Time(4), 2.0, "a", uint64(1)},
					},
				},
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_value", Type: flux.TUInt},
				},
				Data: [][]interface{}{
					{uint64(2)},
				},
			}},
		},
		{
			name: "no group key uints with null",
			spec: &universe.ModeProcedureSpec{Column: "tag1"},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "tag0", Type: flux.TString},
						{Label: "tag1", Type: flux.TUInt},
					},
					Data: [][]interface{}{
						{execute.Time(1), 2.0, "a", uint64(3)},
						{execute.Time(2), 2.0, "a", uint64(2)},
						{execute.Time(3), 2.0, "a", uint64(2)},
						{execute.Time(4), 2.0, "a", uint64(1)},
						{execute.Time(5), 2.0, "a", nil},
						{execute.Time(6), 2.0, "a", nil},
					},
				},
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_value", Type: flux.TUInt},
				},
				Data: [][]interface{}{
					{uint64(2)},
				},
			}},
		},
		{
			name: "no group key bools, null with mode",
			spec: &universe.ModeProcedureSpec{Column: "tag1"},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "tag0", Type: flux.TString},
						{Label: "tag1", Type: flux.TBool},
					},
					Data: [][]interface{}{
						{execute.Time(1), 2.0, "a", true},
						{execute.Time(2), 2.0, "a", false},
						{execute.Time(3), 2.0, "a", false},
						{execute.Time(4), 2.0, "a", false},
						{execute.Time(5), 2.0, "a", true},
						{execute.Time(6), 2.0, "a", true},
						{execute.Time(7), 2.0, "a", nil},
					},
				},
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_value", Type: flux.TBool},
				},
				Data: [][]interface{}{
					{nil},
				},
			}},
		},
		{
			name: "no mode -- all same number of occurrences",
			spec: &universe.ModeProcedureSpec{Column: "tag1"},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "tag0", Type: flux.TString},
						{Label: "tag1", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 2.0, "a", "b"},
						{execute.Time(2), 2.0, "a", "c"},
						{execute.Time(3), 2.0, "a", "b"},
						{execute.Time(4), 2.0, "a", "d"},
						{execute.Time(5), 2.0, "a", "d"},
						{execute.Time(6), 2.0, "a", "c"},
					},
				},
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_value", Type: flux.TString},
				},
				Data: [][]interface{}{
					{nil},
				},
			}},
		},
		{
			name: "multiple modes",
			spec: &universe.ModeProcedureSpec{Column: "tag1"},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "tag0", Type: flux.TString},
						{Label: "tag1", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 2.0, "a", "d"},
						{execute.Time(2), 2.0, "a", "c"},
						{execute.Time(3), 2.0, "a", "d"},
						{execute.Time(4), 2.0, "a", "b"},
						{execute.Time(5), 2.0, "a", "b"},
						{execute.Time(6), 2.0, "a", "c"},
						{execute.Time(7), 2.0, "a", "e"},
					},
				},
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_value", Type: flux.TString},
				},
				Data: [][]interface{}{
					{"b"},
					{"c"},
					{"d"},
				},
			}},
		},
		{
			name: "multiple modes v2",
			spec: &universe.ModeProcedureSpec{Column: "tag1"},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "tag0", Type: flux.TString},
						{Label: "tag1", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 2.0, "a", "b"},
						{execute.Time(2), 2.0, "a", "b"},
						{execute.Time(3), 2.0, "a", "b"},
						{execute.Time(4), 2.0, "a", "b"},
						{execute.Time(5), 2.0, "a", "c"},
						{execute.Time(6), 2.0, "a", "c"},
						{execute.Time(7), 2.0, "a", "c"},
						{execute.Time(8), 2.0, "a", "c"},
						{execute.Time(9), 2.0, "a", "d"},
					},
				},
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_value", Type: flux.TString},
				},
				Data: [][]interface{}{
					{"b"},
					{"c"},
				},
			}},
		},
		{
			name: "column outside group key",
			spec: &universe.ModeProcedureSpec{Column: "tag1"},
			data: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"tag0"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "tag0", Type: flux.TString},
						{Label: "tag1", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 2.0, "a", "b"},
						{execute.Time(2), 2.0, "a", "c"},
						{execute.Time(3), 2.0, "a", "b"},
						{execute.Time(4), 2.0, "a", "d"},
					},
				},
			},
			want: []*executetest.Table{{
				KeyCols: []string{"tag0"},
				ColMeta: []flux.ColMeta{
					{Label: "tag0", Type: flux.TString},
					{Label: "_value", Type: flux.TString},
				},
				Data: [][]interface{}{
					{"a", "b"},
				},
			}},
		},
		{
			name: "column outside group key",
			spec: &universe.ModeProcedureSpec{Column: "tag1"},
			data: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"tag0"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "tag0", Type: flux.TString},
						{Label: "tag1", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 2.0, "a", "b"},
						{execute.Time(2), 2.0, "a", "b"},
						{execute.Time(3), 2.0, "a", "d"},
						{execute.Time(4), 2.0, "a", "d"},
						{execute.Time(5), 2.0, "a", "c"},
					},
				},
			},
			want: []*executetest.Table{{
				KeyCols: []string{"tag0"},
				ColMeta: []flux.ColMeta{
					{Label: "tag0", Type: flux.TString},
					{Label: "_value", Type: flux.TString},
				},
				Data: [][]interface{}{
					{"a", "b"},
					{"a", "d"},
				},
			}},
		},
		{
			name: "column inside group key",
			spec: &universe.ModeProcedureSpec{Column: "tag0"},
			data: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"tag0"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "tag0", Type: flux.TString},
						{Label: "tag1", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 2.0, "a", "b"},
						{execute.Time(2), 2.0, "a", "c"},
						{execute.Time(3), 2.0, "a", "b"},
						{execute.Time(4), 2.0, "a", "d"},
					},
				},
			},
			want: []*executetest.Table{{
				KeyCols: []string{"tag0"},
				ColMeta: []flux.ColMeta{
					{Label: "tag0", Type: flux.TString},
					{Label: "_value", Type: flux.TString},
				},
				Data: [][]interface{}{
					{"a", "a"},
				},
			}},
		},
		{
			name: "two tables",
			spec: &universe.ModeProcedureSpec{Column: "tag1"},
			data: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"tag0"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "tag0", Type: flux.TString},
						{Label: "tag1", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 2.0, "a", "b"},
						{execute.Time(2), 2.0, "a", "c"},
						{execute.Time(3), 2.0, "a", "b"},
						{execute.Time(4), 2.0, "a", "d"},
					},
				},
				&executetest.Table{
					KeyCols: []string{"tag0"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "tag0", Type: flux.TString},
						{Label: "tag1", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 2.0, "x", "b"},
						{execute.Time(2), 2.0, "x", "c"},
						{execute.Time(3), 2.0, "x", "b"},
						{execute.Time(4), 2.0, "x", "e"},
					},
				},
			},
			want: []*executetest.Table{{
				KeyCols: []string{"tag0"},
				ColMeta: []flux.ColMeta{
					{Label: "tag0", Type: flux.TString},
					{Label: "_value", Type: flux.TString},
				},
				Data: [][]interface{}{
					{"a", "b"},
				},
			},
				{
					KeyCols: []string{"tag0"},
					ColMeta: []flux.ColMeta{
						{Label: "tag0", Type: flux.TString},
						{Label: "_value", Type: flux.TString},
					},
					Data: [][]interface{}{
						{"x", "b"},
					},
				},
			},
		},
		{
			name: "null with mode",
			spec: &universe.ModeProcedureSpec{Column: "tag1"},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "tag0", Type: flux.TString},
						{Label: "tag1", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(0), 2.0, "a", nil},
						{execute.Time(1), 2.0, "a", "b"},
						{execute.Time(2), 2.0, "a", "c"},
						{execute.Time(3), 2.0, "a", "b"},
						{execute.Time(4), 2.0, "a", "c"},
					},
				},
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_value", Type: flux.TString},
				},
				Data: [][]interface{}{
					{nil},
				},
			}},
		},
		{
			name: "null without mode",
			spec: &universe.ModeProcedureSpec{Column: "tag1"},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "tag0", Type: flux.TString},
						{Label: "tag1", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(0), 2.0, "a", nil},
						{execute.Time(1), 2.0, "a", "b"},
						{execute.Time(2), 2.0, "a", "c"},
						{execute.Time(3), 2.0, "a", "b"},
						{execute.Time(4), 2.0, "a", "c"},
						{execute.Time(5), 2.0, "a", nil},
					},
				},
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_value", Type: flux.TString},
				},
				Data: [][]interface{}{
					{nil},
				},
			}},
		},
		{
			name: "more nulls than others, floats",
			spec: &universe.ModeProcedureSpec{Column: "tag1"},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "tag0", Type: flux.TString},
						{Label: "tag1", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(0), 2.0, "a", nil},
						{execute.Time(1), 2.0, "a", float64(1.0)},
						{execute.Time(2), 2.0, "a", float64(1.0)},
						{execute.Time(3), 2.0, "a", float64(2.0)},
						{execute.Time(4), 2.0, "a", float64(2.0)},
						{execute.Time(5), 2.0, "a", nil},
						{execute.Time(6), 2.0, "a", nil},
					},
				},
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{nil},
				},
			}},
		},
		{
			name: "more nulls than others, ints",
			spec: &universe.ModeProcedureSpec{Column: "tag1"},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "tag0", Type: flux.TString},
						{Label: "tag1", Type: flux.TInt},
					},
					Data: [][]interface{}{
						{execute.Time(0), 2.0, "a", nil},
						{execute.Time(1), 2.0, "a", int64(1)},
						{execute.Time(2), 2.0, "a", int64(1)},
						{execute.Time(3), 2.0, "a", int64(2)},
						{execute.Time(4), 2.0, "a", int64(2)},
						{execute.Time(5), 2.0, "a", nil},
						{execute.Time(6), 2.0, "a", nil},
					},
				},
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_value", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{nil},
				},
			}},
		},
		{
			name: "more nulls than others, uints",
			spec: &universe.ModeProcedureSpec{Column: "tag1"},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "tag0", Type: flux.TString},
						{Label: "tag1", Type: flux.TUInt},
					},
					Data: [][]interface{}{
						{execute.Time(0), 2.0, "a", nil},
						{execute.Time(1), 2.0, "a", uint64(1)},
						{execute.Time(2), 2.0, "a", uint64(1)},
						{execute.Time(3), 2.0, "a", uint64(2)},
						{execute.Time(4), 2.0, "a", uint64(2)},
						{execute.Time(5), 2.0, "a", nil},
						{execute.Time(6), 2.0, "a", nil},
					},
				},
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_value", Type: flux.TUInt},
				},
				Data: [][]interface{}{
					{nil},
				},
			}},
		},
		{
			name: "more nulls than others, times",
			spec: &universe.ModeProcedureSpec{Column: "tag1"},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "tag0", Type: flux.TString},
						{Label: "tag1", Type: flux.TTime},
					},
					Data: [][]interface{}{
						{execute.Time(0), 2.0, "a", nil},
						{execute.Time(1), 2.0, "a", execute.Time(1)},
						{execute.Time(2), 2.0, "a", execute.Time(1)},
						{execute.Time(3), 2.0, "a", execute.Time(2)},
						{execute.Time(4), 2.0, "a", execute.Time(2)},
						{execute.Time(5), 2.0, "a", nil},
						{execute.Time(6), 2.0, "a", nil},
					},
				},
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_value", Type: flux.TTime},
				},
				Data: [][]interface{}{
					{nil},
				},
			}},
		},
		{
			name: "all nulls - time",
			spec: &universe.ModeProcedureSpec{Column: "tag1"},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "tag0", Type: flux.TString},
						{Label: "tag1", Type: flux.TTime},
					},
					Data: [][]interface{}{
						{execute.Time(0), 2.0, "a", nil},
						{execute.Time(1), 2.0, "a", nil},
						{execute.Time(2), 2.0, "a", nil},
						{execute.Time(3), 2.0, "a", nil},
						{execute.Time(4), 2.0, "a", nil},
						{execute.Time(5), 2.0, "a", nil},
						{execute.Time(6), 2.0, "a", nil},
					},
				},
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_value", Type: flux.TTime},
				},
				Data: [][]interface{}{
					{nil},
				},
			}},
		},
		{
			name: "all nulls - strings",
			spec: &universe.ModeProcedureSpec{Column: "tag1"},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "tag0", Type: flux.TString},
						{Label: "tag1", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(0), 2.0, "a", nil},
						{execute.Time(1), 2.0, "a", nil},
						{execute.Time(2), 2.0, "a", nil},
						{execute.Time(3), 2.0, "a", nil},
						{execute.Time(4), 2.0, "a", nil},
						{execute.Time(5), 2.0, "a", nil},
						{execute.Time(6), 2.0, "a", nil},
					},
				},
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_value", Type: flux.TString},
				},
				Data: [][]interface{}{
					{nil},
				},
			}},
		},
		{
			name: "all nulls - int",
			spec: &universe.ModeProcedureSpec{Column: "tag1"},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "tag0", Type: flux.TString},
						{Label: "tag1", Type: flux.TInt},
					},
					Data: [][]interface{}{
						{execute.Time(0), 2.0, "a", nil},
						{execute.Time(1), 2.0, "a", nil},
						{execute.Time(2), 2.0, "a", nil},
						{execute.Time(3), 2.0, "a", nil},
						{execute.Time(4), 2.0, "a", nil},
						{execute.Time(5), 2.0, "a", nil},
						{execute.Time(6), 2.0, "a", nil},
					},
				},
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_value", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{nil},
				},
			}},
		},
		{
			name: "all nulls - uints",
			spec: &universe.ModeProcedureSpec{Column: "tag1"},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "tag0", Type: flux.TString},
						{Label: "tag1", Type: flux.TUInt},
					},
					Data: [][]interface{}{
						{execute.Time(0), 2.0, "a", nil},
						{execute.Time(1), 2.0, "a", nil},
						{execute.Time(2), 2.0, "a", nil},
						{execute.Time(3), 2.0, "a", nil},
						{execute.Time(4), 2.0, "a", nil},
						{execute.Time(5), 2.0, "a", nil},
						{execute.Time(6), 2.0, "a", nil},
					},
				},
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_value", Type: flux.TUInt},
				},
				Data: [][]interface{}{
					{nil},
				},
			}},
		},
		{
			name: "all nulls - bools",
			spec: &universe.ModeProcedureSpec{Column: "tag1"},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "tag0", Type: flux.TString},
						{Label: "tag1", Type: flux.TBool},
					},
					Data: [][]interface{}{
						{execute.Time(0), 2.0, "a", nil},
						{execute.Time(1), 2.0, "a", nil},
						{execute.Time(2), 2.0, "a", nil},
						{execute.Time(3), 2.0, "a", nil},
						{execute.Time(4), 2.0, "a", nil},
						{execute.Time(5), 2.0, "a", nil},
						{execute.Time(6), 2.0, "a", nil},
					},
				},
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_value", Type: flux.TBool},
				},
				Data: [][]interface{}{
					{nil},
				},
			}},
		},
		{
			name: "all same value",
			spec: &universe.ModeProcedureSpec{Column: "tag1"},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "tag0", Type: flux.TString},
						{Label: "tag1", Type: flux.TInt},
					},
					Data: [][]interface{}{
						{execute.Time(0), 2.0, "a", int64(1)},
						{execute.Time(1), 2.0, "a", int64(1)},
						{execute.Time(2), 2.0, "a", int64(1)},
						{execute.Time(3), 2.0, "a", int64(1)},
						{execute.Time(4), 2.0, "a", int64(1)},
						{execute.Time(5), 2.0, "a", int64(1)},
						{execute.Time(6), 2.0, "a", int64(1)},
					},
				},
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_value", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{nil},
				},
			}},
		},
		{
			name: "time",
			spec: &universe.ModeProcedureSpec{Column: "tag1"},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "tag0", Type: flux.TString},
						{Label: "tag1", Type: flux.TTime},
					},
					Data: [][]interface{}{
						{execute.Time(0), 2.0, "a", execute.Time(1234)},
						{execute.Time(0), 2.0, "a", execute.Time(1234)},
						{execute.Time(0), 2.0, "a", execute.Time(1334)},
					},
				},
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_value", Type: flux.TTime},
				},
				Data: [][]interface{}{
					{execute.Time(1234)},
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
					return universe.NewModeTransformation(d, c, tc.spec)
				},
			)
		})
	}
}
