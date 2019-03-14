package universe_test

import (
	"testing"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/querytest"
	"github.com/influxdata/flux/stdlib/universe"
)

func TestShiftOperation_Marshaling(t *testing.T) {
	data := []byte(`{"id":"shift","kind":"timeShift","spec":{"duration":"1h"}}`)
	op := &flux.Operation{
		ID: "shift",
		Spec: &universe.ShiftOpSpec{
			Shift: flux.Duration(1 * time.Hour),
		},
	}
	querytest.OperationMarshalingTestHelper(t, data, op)
}

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
				Shift:   flux.Duration(1),
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
				Shift:   flux.Duration(2),
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
				Shift:   flux.Duration(1),
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
				Shift:   flux.Duration(1),
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
			executetest.ProcessTestHelper(
				t,
				tc.data,
				tc.want,
				nil,
				func(d execute.Dataset, c execute.TableBuilderCache) execute.Transformation {
					return universe.NewShiftTransformation(d, c, tc.spec)
				},
			)
		})
	}
}
