package universe_test

import (
	"testing"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/stdlib/universe"
)

func TestDistinct_Process(t *testing.T) {
	testCases := []struct {
		name string
		spec *universe.DistinctProcedureSpec
		data []flux.Table
		want []*executetest.Table
	}{
		{
			name: "no group key",
			spec: &universe.DistinctProcedureSpec{Column: "tag1"},
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
					{"c"},
					{"d"},
				},
			}},
		},
		{
			name: "column outside group key",
			spec: &universe.DistinctProcedureSpec{Column: "tag1"},
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
					{"a", "c"},
					{"a", "d"},
				},
			}},
		},
		{
			name: "column inside group key",
			spec: &universe.DistinctProcedureSpec{Column: "tag0"},
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
			spec: &universe.DistinctProcedureSpec{Column: "tag1"},
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
					{"a", "c"},
					{"a", "d"},
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
						{"x", "c"},
						{"x", "e"},
					},
				},
			},
		},
		{
			name: "nulls",
			spec: &universe.DistinctProcedureSpec{Column: "tag1"},
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
						{execute.Time(4), 2.0, "a", nil},
						{execute.Time(5), 2.0, "a", "d"},
					},
				},
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_value", Type: flux.TString},
				},
				Data: [][]interface{}{
					{nil},
					{"b"},
					{"c"},
					{"d"},
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
					return universe.NewDistinctTransformation(d, c, tc.spec)
				},
			)
		})
	}
}
