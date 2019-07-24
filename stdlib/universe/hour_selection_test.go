package universe_test

import (
	"fmt"
	"testing"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/stdlib/universe"
)

const hour int64 = 3600000000000

func TestHourSelection_Process(t *testing.T) {
	testCases := []struct {
		name    string
		spec    *universe.HourSelectionProcedureSpec
		data    []flux.Table
		want    []*executetest.Table
		wantErr error
	}{
		{
			name: "no group key",
			spec: &universe.HourSelectionProcedureSpec{Start: 17, Stop: 19, TimeColumn: execute.DefaultTimeColLabel},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "tag0", Type: flux.TString},
						{Label: "tag1", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(60000000000000), "a", "b"},
						{execute.Time(70000000000000), "a", "c"},
						{execute.Time(80000000000000), "a", "b"},
						{execute.Time(90000000000000), "a", "d"},
					},
				},
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "tag0", Type: flux.TString},
					{Label: "tag1", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(70000000000000), "a", "c"},
				},
			}},
		},
		{
			name: "no group key 2",
			spec: &universe.HourSelectionProcedureSpec{Start: 6, Stop: 18, TimeColumn: execute.DefaultTimeColLabel},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "tag0", Type: flux.TString},
						{Label: "tag1", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(hour), "a", "d"}, // 1 hour
						{execute.Time(hour * 2), "a", "b"},
						{execute.Time(hour * 4), "a", "c"},
						{execute.Time(hour * 8), "a", "e"},
						{execute.Time(hour * 12), "a", "f"},
						{execute.Time(hour * 16), "a", "g"},
						{execute.Time(hour * 22), "a", "h"},
					},
				},
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "tag0", Type: flux.TString},
					{Label: "tag1", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(hour * 8), "a", "e"},
					{execute.Time(hour * 12), "a", "f"},
					{execute.Time(hour * 16), "a", "g"},
				},
			}},
		},
		{
			name: "no group key -- inclusivity",
			spec: &universe.HourSelectionProcedureSpec{Start: 4, Stop: 16, TimeColumn: execute.DefaultTimeColLabel},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "tag0", Type: flux.TString},
						{Label: "tag1", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(hour), "a", "d"}, // 1 hour
						{execute.Time(hour * 2), "a", "b"},
						{execute.Time(hour * 4), "a", "c"},
						{execute.Time(hour * 8), "a", "e"},
						{execute.Time(hour * 12), "a", "f"},
						{execute.Time(hour * 16), "a", "g"},
						{execute.Time(hour * 22), "a", "h"},
					},
				},
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "tag0", Type: flux.TString},
					{Label: "tag1", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(hour * 4), "a", "c"},
					{execute.Time(hour * 8), "a", "e"},
					{execute.Time(hour * 12), "a", "f"},
					{execute.Time(hour * 16), "a", "g"},
				},
			}},
		},
		{
			name: "with group key",
			spec: &universe.HourSelectionProcedureSpec{Start: 4, Stop: 10, TimeColumn: execute.DefaultTimeColLabel},
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
						{execute.Time(hour), 2.0, "a", "b"},
						{execute.Time(hour * 4), 2.0, "a", "c"},
						{execute.Time(hour * 6), 2.0, "a", "b"},
						{execute.Time(hour * 10), 2.0, "a", "d"},
					},
				},
			},
			want: []*executetest.Table{{
				KeyCols: []string{"tag0"},
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "tag0", Type: flux.TString},
					{Label: "tag1", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(hour * 4), 2.0, "a", "c"},
					{execute.Time(hour * 6), 2.0, "a", "b"},
					{execute.Time(hour * 10), 2.0, "a", "d"},
				},
			}},
		},
		{
			name: "two tables",
			spec: &universe.HourSelectionProcedureSpec{Start: 4, Stop: 23, TimeColumn: execute.DefaultTimeColLabel},
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
						{execute.Time(hour), 2.0, "a", "b"},
						{execute.Time(hour * 2), 2.0, "a", "c"},
						{execute.Time(hour * 8), 2.0, "a", "b"},
						{execute.Time(hour * 14), 2.0, "a", "d"},
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
						{execute.Time(hour), 2.0, "x", "b"},
						{execute.Time(hour * 3), 2.0, "x", "c"},
						{execute.Time(hour * 9), 2.0, "x", "b"},
						{execute.Time(hour * 13), 2.0, "x", "e"},
					},
				},
			},
			want: []*executetest.Table{{
				KeyCols: []string{"tag0"},
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "tag0", Type: flux.TString},
					{Label: "tag1", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(hour * 8), 2.0, "a", "b"},
					{execute.Time(hour * 14), 2.0, "a", "d"},
				},
			},
				{
					KeyCols: []string{"tag0"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "tag0", Type: flux.TString},
						{Label: "tag1", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(hour * 9), 2.0, "x", "b"},
						{execute.Time(hour * 13), 2.0, "x", "e"},
					},
				},
			},
		},
		{
			name: "no group key -- across multiple days",
			spec: &universe.HourSelectionProcedureSpec{Start: 2, Stop: 8, TimeColumn: execute.DefaultTimeColLabel},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "tag0", Type: flux.TString},
						{Label: "tag1", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(hour), "a", "d"},
						{execute.Time(hour * 2), "a", "b"},
						{execute.Time(hour * 4), "a", "c"},
						{execute.Time(hour * 8), "a", "e"},
						{execute.Time(hour * 26), "a", "f"},
						{execute.Time(hour * 28), "a", "g"},
						{execute.Time(hour * 32), "a", "h"},
					},
				},
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "tag0", Type: flux.TString},
					{Label: "tag1", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(hour * 2), "a", "b"},
					{execute.Time(hour * 4), "a", "c"},
					{execute.Time(hour * 8), "a", "e"},
					{execute.Time(hour * 26), "a", "f"},
					{execute.Time(hour * 28), "a", "g"},
					{execute.Time(hour * 32), "a", "h"},
				},
			}},
		},
		{
			name: "no group key -- nil",
			spec: &universe.HourSelectionProcedureSpec{Start: 8, Stop: 23, TimeColumn: execute.DefaultTimeColLabel},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "tag0", Type: flux.TString},
						{Label: "tag1", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(hour), "a", "b"},
						{execute.Time(hour * 3), "a", "c"},
						{execute.Time(hour * 5), "a", "b"},
						{execute.Time(hour * 7), "a", "d"},
					},
				},
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "tag0", Type: flux.TString},
					{Label: "tag1", Type: flux.TString},
				},
				Data: [][]interface{}(nil),
			}},
		},
		{
			name: "no group key -- outside range",
			spec: &universe.HourSelectionProcedureSpec{Start: -25, Stop: 8, TimeColumn: execute.DefaultTimeColLabel},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "tag0", Type: flux.TString},
						{Label: "tag1", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(hour * 6), "a", "b"},
						{execute.Time(hour * 20), "a", "c"},
						{execute.Time(hour * 21), "a", "b"},
						{execute.Time(hour * 22), "a", "d"},
					},
				},
			},
			wantErr: fmt.Errorf("start must be between 0 and 23"),
		},
		{
			name: "no group key -- outside range",
			spec: &universe.HourSelectionProcedureSpec{Start: 2, Stop: 74, TimeColumn: execute.DefaultTimeColLabel},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "tag0", Type: flux.TString},
						{Label: "tag1", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(hour * 6), "a", "b"},
						{execute.Time(hour * 20), "a", "c"},
						{execute.Time(hour * 21), "a", "b"},
						{execute.Time(hour * 22), "a", "d"},
					},
				},
			},
			wantErr: fmt.Errorf("stop must be between 0 and 23"),
		},
		{
			name: "no group key -- multiple with same hour",
			spec: &universe.HourSelectionProcedureSpec{Start: 6, Stop: 6, TimeColumn: execute.DefaultTimeColLabel},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "tag0", Type: flux.TString},
						{Label: "tag1", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(hour * 6), "a", "b"},
						{execute.Time(3600000100000 * 6), "a", "c"},
					},
				},
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "tag0", Type: flux.TString},
					{Label: "tag1", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(hour * 6), "a", "b"},
					{execute.Time(3600000100000 * 6), "a", "c"},
				},
			}},
		},
		{
			name: "with two group keys, multiple days",
			spec: &universe.HourSelectionProcedureSpec{Start: 4, Stop: 6, TimeColumn: execute.DefaultTimeColLabel},
			data: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"tag0", "tag1"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "tag0", Type: flux.TString},
						{Label: "tag1", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(hour * 4), 2.0, "a", "b"},
						{execute.Time(hour * 28), 2.0, "a", "b"},
						{execute.Time(hour * 6), 2.0, "a", "b"},
						{execute.Time(hour * 30), 2.0, "a", "b"},
						{execute.Time(hour * 32), 2.0, "a", "b"},
					},
				},
			},
			want: []*executetest.Table{{
				KeyCols: []string{"tag0", "tag1"},
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "tag0", Type: flux.TString},
					{Label: "tag1", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(hour * 4), 2.0, "a", "b"},
					{execute.Time(hour * 28), 2.0, "a", "b"},
					{execute.Time(hour * 6), 2.0, "a", "b"},
					{execute.Time(hour * 30), 2.0, "a", "b"},
				},
			}},
		},
		{
			name: "no time column",
			spec: &universe.HourSelectionProcedureSpec{Start: 4, Stop: 6, TimeColumn: execute.DefaultTimeColLabel},
			data: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"tag0", "tag1"},
					ColMeta: []flux.ColMeta{
						{Label: "_value", Type: flux.TFloat},
						{Label: "tag0", Type: flux.TString},
						{Label: "tag1", Type: flux.TString},
					},
					Data: [][]interface{}{
						{2.0, "a", "b"},
						{2.0, "a", "c"},
						{2.0, "a", "b"},
						{2.0, "a", "d"},
						{2.0, "a", "d"},
					},
				},
			},
			wantErr: fmt.Errorf("invalid time column"),
		},
		{
			name: "no group key -- non-default time column",
			spec: &universe.HourSelectionProcedureSpec{Start: 6, Stop: 6, TimeColumn: "newTime"},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "newTime", Type: flux.TTime},
						{Label: "tag0", Type: flux.TString},
						{Label: "tag1", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(hour * 6), "a", "b"},
						{execute.Time(3600000100000 * 6), "a", "c"},
					},
				},
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "newTime", Type: flux.TTime},
					{Label: "tag0", Type: flux.TString},
					{Label: "tag1", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(hour * 6), "a", "b"},
					{execute.Time(3600000100000 * 6), "a", "c"},
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
					return universe.NewHourSelectionTransformation(d, c, tc.spec)
				},
			)
		})
	}
}
