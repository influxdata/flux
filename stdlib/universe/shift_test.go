package universe_test

import (
	"testing"

	"github.com/InfluxCommunity/flux"
	"github.com/InfluxCommunity/flux/execute"
	"github.com/InfluxCommunity/flux/execute/executetest"
	"github.com/InfluxCommunity/flux/memory"
	"github.com/InfluxCommunity/flux/stdlib/universe"
)

func TestShift_Process(t *testing.T) {
	cols := []flux.ColMeta{
		{Label: "t1", Type: flux.TString},
		{Label: execute.DefaultTimeColLabel, Type: flux.TTime},
		{Label: execute.DefaultValueColLabel, Type: flux.TFloat},
	}

	testCases := []struct {
		name string
		spec *universe.ShiftProcedureSpec
		data []flux.Table
		want []*executetest.Table
	}{
		{
			name: "one table",
			spec: &universe.ShiftProcedureSpec{
				Columns: []string{execute.DefaultTimeColLabel},
				Shift:   flux.ConvertDuration(1),
			},
			data: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"t1"},
					ColMeta: cols,
					Data: [][]interface{}{
						{"a", execute.Time(1), 2.0},
						{"a", execute.Time(2), 1.0},
					},
				},
			},
			want: []*executetest.Table{
				{
					KeyCols: []string{"t1"},
					ColMeta: cols,
					Data: [][]interface{}{
						{"a", execute.Time(2), 2.0},
						{"a", execute.Time(3), 1.0},
					},
				},
			},
		},
		{
			name: "multiple tables",
			spec: &universe.ShiftProcedureSpec{
				Columns: []string{execute.DefaultTimeColLabel},
				Shift:   flux.ConvertDuration(2),
			},
			data: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"t1"},
					ColMeta: cols,
					Data: [][]interface{}{
						{"a", execute.Time(1), 2.0},
						{"a", execute.Time(2), 1.0},
					},
				},
				&executetest.Table{
					KeyCols: []string{"t1"},
					ColMeta: cols,
					Data: [][]interface{}{
						{"b", execute.Time(3), 3.0},
						{"b", execute.Time(4), 4.0},
					},
				},
			},
			want: []*executetest.Table{
				{
					KeyCols: []string{"t1"},
					ColMeta: cols,
					Data: [][]interface{}{
						{"a", execute.Time(3), 2.0},
						{"a", execute.Time(4), 1.0},
					},
				},
				{
					KeyCols: []string{"t1"},
					ColMeta: cols,
					Data: [][]interface{}{
						{"b", execute.Time(5), 3.0},
						{"b", execute.Time(6), 4.0},
					},
				},
			},
		},
		{
			name: "null time",
			spec: &universe.ShiftProcedureSpec{
				Columns: []string{execute.DefaultTimeColLabel},
				Shift:   flux.ConvertDuration(1),
			},
			data: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"t1"},
					ColMeta: cols,
					Data: [][]interface{}{
						{"a", execute.Time(1), 2.0},
						{"a", nil, 1.0},
					},
				},
			},
			want: []*executetest.Table{
				{
					KeyCols: []string{"t1"},
					ColMeta: cols,
					Data: [][]interface{}{
						{"a", execute.Time(2), 2.0},
						{"a", nil, 1.0},
					},
				},
			},
		},
		{
			name: "null value",
			spec: &universe.ShiftProcedureSpec{
				Columns: []string{execute.DefaultTimeColLabel},
				Shift:   flux.ConvertDuration(1),
			},
			data: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"t1"},
					ColMeta: cols,
					Data: [][]interface{}{
						{"a", execute.Time(1), 2.0},
						{"a", execute.Time(2), nil},
					},
				},
			},
			want: []*executetest.Table{
				{
					KeyCols: []string{"t1"},
					ColMeta: cols,
					Data: [][]interface{}{
						{"a", execute.Time(2), 2.0},
						{"a", execute.Time(3), nil},
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			executetest.ProcessTestHelper2(
				t,
				tc.data,
				tc.want,
				nil,
				func(id execute.DatasetID, alloc memory.Allocator) (execute.Transformation, execute.Dataset) {
					tr, d, err := universe.NewShiftTransformation(id, tc.spec, alloc)
					if err != nil {
						t.Fatal(err)
					}
					return tr, d
				},
			)
		})
	}
}
