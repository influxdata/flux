package promql_test

import (
	"context"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	_ "github.com/influxdata/flux/builtin"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/stdlib/internal/promql"
	"github.com/influxdata/flux/values"
	"github.com/influxdata/flux/values/valuestest"
)

func TestJoin(t *testing.T) {
	testCases := []struct {
		name        string
		fn          interpreter.ResolvedFunction
		left, right []flux.Table
		want        []*executetest.Table
		wantErr     bool
		skip        string
	}{
		{
			name: "multiple column readers",
			// fn: (left, right) => ({left with w: right._value})
			fn: interpreter.ResolvedFunction{
				Fn:    executetest.FunctionExpression(t, `(left, right) => ({left with w: right._value})`),
				Scope: valuestest.Scope(),
			},
			left: []flux.Table{
				&executetest.RowWiseTable{
					Table: &executetest.Table{
						KeyCols: []string{"_field"},
						ColMeta: []flux.ColMeta{
							{Label: "_time", Type: flux.TTime},
							{Label: "_field", Type: flux.TString},
							{Label: "_value", Type: flux.TFloat},
						},
						Data: [][]interface{}{
							{execute.Time(1), "_value", 1.0},
							{execute.Time(2), "_value", 2.0},
							{execute.Time(2), "_value", 3.0},
						},
					},
				},
			},
			right: []flux.Table{
				&executetest.RowWiseTable{
					Table: &executetest.Table{
						KeyCols: []string{"_field"},
						ColMeta: []flux.ColMeta{
							{Label: "_time", Type: flux.TTime},
							{Label: "_field", Type: flux.TString},
							{Label: "_value", Type: flux.TFloat},
						},
						Data: [][]interface{}{
							{execute.Time(1), "_value", 10.0},
							{execute.Time(1), "_value", 20.0},
							{execute.Time(1), "_value", 30.0},
							{execute.Time(2), "_value", 10.0},
							{execute.Time(2), "_value", 20.0},
							{execute.Time(3), "_value", 30.0},
						},
					},
				},
			},
			want: []*executetest.Table{
				{
					KeyCols: []string{"_field"},
					ColMeta: []flux.ColMeta{
						{Label: "_field", Type: flux.TString},
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "w", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{"_value", execute.Time(1), 1.0, 10.0},
						{"_value", execute.Time(1), 1.0, 20.0},
						{"_value", execute.Time(1), 1.0, 30.0},
						{"_value", execute.Time(2), 2.0, 10.0},
						{"_value", execute.Time(2), 2.0, 20.0},
						{"_value", execute.Time(2), 3.0, 10.0},
						{"_value", execute.Time(2), 3.0, 20.0},
					},
				},
			},
		},
		{
			name: "rows with same time",
			// fn: (left, right) => ({left with w: right._value})
			fn: interpreter.ResolvedFunction{
				Fn:    executetest.FunctionExpression(t, `(left, right) => ({left with w: right._value})`),
				Scope: valuestest.Scope(),
			},
			left: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"_field"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_field", Type: flux.TString},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), "_value", 1.0},
						{execute.Time(2), "_value", 2.0},
						{execute.Time(2), "_value", 3.0},
					},
				},
			},
			right: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"_field"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_field", Type: flux.TString},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), "_value", 10.0},
						{execute.Time(1), "_value", 20.0},
						{execute.Time(1), "_value", 30.0},
						{execute.Time(2), "_value", 10.0},
						{execute.Time(2), "_value", 20.0},
						{execute.Time(3), "_value", 30.0},
					},
				},
			},
			want: []*executetest.Table{
				{
					KeyCols: []string{"_field"},
					ColMeta: []flux.ColMeta{
						{Label: "_field", Type: flux.TString},
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "w", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{"_value", execute.Time(1), 1.0, 10.0},
						{"_value", execute.Time(1), 1.0, 20.0},
						{"_value", execute.Time(1), 1.0, 30.0},
						{"_value", execute.Time(2), 2.0, 10.0},
						{"_value", execute.Time(2), 2.0, 20.0},
						{"_value", execute.Time(2), 3.0, 10.0},
						{"_value", execute.Time(2), 3.0, 20.0},
					},
				},
			},
		},
		{
			name: "multiple tables",
			// fn: (left, right) => left
			fn: interpreter.ResolvedFunction{
				Fn:    executetest.FunctionExpression(t, `(left, right) => left`),
				Scope: valuestest.Scope(),
			},
			left: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"tag1", "tag2"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "tag1", Type: flux.TString},
						{Label: "tag2", Type: flux.TString},
						{Label: "v", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), "1", "2", 1.0},
						{execute.Time(2), "1", "2", 2.0},
						{execute.Time(3), "1", "2", 3.0},
					},
				},
				&executetest.Table{
					KeyCols: []string{"tag1", "tag2"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "tag1", Type: flux.TString},
						{Label: "tag2", Type: flux.TString},
						{Label: "v", Type: flux.TInt},
					},
					Data: [][]interface{}{
						{execute.Time(10), "11", "22", int64(23)},
						{execute.Time(11), "11", "22", int64(19)},
						{execute.Time(12), "11", "22", int64(55)},
					},
				},
				&executetest.Table{
					KeyCols: []string{"tag1", "tag2"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "tag1", Type: flux.TString},
						{Label: "tag2", Type: flux.TString},
						{Label: "v", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(7), "111", "222", 1.1},
						{execute.Time(8), "111", "222", 5.5},
						{execute.Time(9), "111", "222", 3.3},
					},
				},
			},
			right: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"tag1", "tag2"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "tag1", Type: flux.TString},
						{Label: "tag2", Type: flux.TString},
						{Label: "w", Type: flux.TInt},
					},
					Data: [][]interface{}{
						{execute.Time(12), "11", "22", int64(23)},
						{execute.Time(22), "11", "22", int64(19)},
						{execute.Time(55), "11", "22", int64(55)},
					},
				},
				&executetest.Table{
					KeyCols: []string{"tag1", "tag2"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "tag1", Type: flux.TString},
						{Label: "tag2", Type: flux.TString},
						{Label: "w", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), "0", "0", 1.0},
						{execute.Time(2), "0", "0", 2.0},
						{execute.Time(3), "0", "0", 3.0},
					},
				},
				&executetest.Table{
					KeyCols: []string{"tag1", "tag2"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "tag1", Type: flux.TString},
						{Label: "tag2", Type: flux.TString},
						{Label: "w", Type: flux.TInt},
					},
					Data: [][]interface{}{
						{execute.Time(7), "111", "222", int64(16)},
						{execute.Time(8), "111", "222", int64(17)},
						{execute.Time(10), "111", "222", int64(18)},
					},
				},
			},
			want: []*executetest.Table{
				{
					KeyCols: []string{"tag1", "tag2"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "tag1", Type: flux.TString},
						{Label: "tag2", Type: flux.TString},
						{Label: "v", Type: flux.TInt},
					},
					Data: [][]interface{}{
						{execute.Time(12), "11", "22", int64(55)},
					},
				},
				{
					KeyCols: []string{"tag1", "tag2"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "tag1", Type: flux.TString},
						{Label: "tag2", Type: flux.TString},
						{Label: "v", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(7), "111", "222", 1.1},
						{execute.Time(8), "111", "222", 5.5},
					},
				},
			},
		},
		{
			name: "nulls",
			// fn: (left, right) => ({left with w: right.w})
			fn: interpreter.ResolvedFunction{
				Fn:    executetest.FunctionExpression(t, `(left, right) => ({left with w: right.w})`),
				Scope: valuestest.Scope(),
			},
			left: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"tag1", "tag2"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "tag1", Type: flux.TString},
						{Label: "tag2", Type: flux.TString},
						{Label: "v", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), "1", "2", 1.0},
						{execute.Time(2), "1", "2", 2.0},
						{execute.Time(3), "1", "2", nil},
					},
				},
			},
			right: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"tag1", "tag2"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "tag1", Type: flux.TString},
						{Label: "tag2", Type: flux.TString},
						{Label: "w", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(2), "1", "2", nil},
						{execute.Time(3), "1", "2", 30.0},
						{execute.Time(4), "1", "2", 40.0},
					},
				},
			},
			want: []*executetest.Table{
				{
					KeyCols: []string{"tag1", "tag2"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "tag1", Type: flux.TString},
						{Label: "tag2", Type: flux.TString},
						{Label: "v", Type: flux.TFloat},
						{Label: "w", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(2), "1", "2", 2.0, nil},
						{execute.Time(3), "1", "2", nil, 30.0},
					},
				},
			},
		},
		{
			name: "regular",
			// fn: (left, right) => ({left with w: right.w})
			fn: interpreter.ResolvedFunction{
				Fn:    executetest.FunctionExpression(t, `(left, right) => ({left with w: right.w})`),
				Scope: valuestest.Scope(),
			},
			left: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"tag1", "tag2"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "tag1", Type: flux.TString},
						{Label: "tag2", Type: flux.TString},
						{Label: "v", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), "1", "2", 1.0},
						{execute.Time(2), "1", "2", 2.0},
						{execute.Time(3), "1", "2", 3.0},
					},
				},
			},
			right: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"tag1", "tag2"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "tag1", Type: flux.TString},
						{Label: "tag2", Type: flux.TString},
						{Label: "w", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), "1", "2", 10.0},
						{execute.Time(2), "1", "2", 20.0},
						{execute.Time(3), "1", "2", 30.0},
					},
				},
			},
			want: []*executetest.Table{
				{
					KeyCols: []string{"tag1", "tag2"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "tag1", Type: flux.TString},
						{Label: "tag2", Type: flux.TString},
						{Label: "v", Type: flux.TFloat},
						{Label: "w", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), "1", "2", 1.0, 10.0},
						{execute.Time(2), "1", "2", 2.0, 20.0},
						{execute.Time(3), "1", "2", 3.0, 30.0},
					},
				},
			},
		},
		{
			name: "no matches",
			// fn: (left, right) => left
			fn: interpreter.ResolvedFunction{
				Fn:    executetest.FunctionExpression(t, `(left, right) => left`),
				Scope: valuestest.Scope(),
			},
			left: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"_field"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_field", Type: flux.TString},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), "w", 1.0},
						{execute.Time(2), "w", 2.0},
						{execute.Time(3), "w", 3.0},
					},
				},
			},
			right: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"_field"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_field", Type: flux.TString},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(4), "w", 10.0},
						{execute.Time(5), "w", 20.0},
						{execute.Time(6), "w", 30.0},
					},
				},
			},
			want: []*executetest.Table{
				{
					GroupKey: execute.NewGroupKey([]flux.ColMeta{{
						Label: "_field",
						Type:  flux.TString,
					}}, []values.Value{
						values.NewString("w"),
					}),
					KeyCols:   []string{"_field"},
					KeyValues: []interface{}{"w"},
				},
			},
		},
		{
			name: "no matches",
			// fn: (left, right) => left
			fn: interpreter.ResolvedFunction{
				Fn:    executetest.FunctionExpression(t, "(left, right) => left"),
				Scope: valuestest.Scope(),
			},
			left: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"_field"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_field", Type: flux.TString},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), "w", 1.0},
						{execute.Time(2), "w", 2.0},
						{execute.Time(3), "w", 3.0},
					},
				},
			},
			right: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"_field"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_field", Type: flux.TString},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), "v", 10.0},
						{execute.Time(2), "v", 20.0},
						{execute.Time(3), "v", 30.0},
					},
				},
			},
		},
		{
			name: "modify group key",
			// fn: (left, right) => ({A: right.B, B: left.A})
			fn: interpreter.ResolvedFunction{
				Fn:    executetest.FunctionExpression(t, `(left, right) => ({A: right.B, B: left.A})`),
				Scope: valuestest.Scope(),
			},
			left: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"A", "B"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "A", Type: flux.TString},
						{Label: "B", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 1.0, "a", "b"},
						{execute.Time(2), 2.0, "a", "b"},
						{execute.Time(3), 3.0, "a", "b"},
					},
				},
			},
			right: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"A", "B"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "A", Type: flux.TString},
						{Label: "B", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 10.0, "a", "b"},
						{execute.Time(2), 20.0, "a", "b"},
						{execute.Time(3), 30.0, "a", "b"},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "modify group key",
			// fn: (left, right) => ({A: left.A})
			fn: interpreter.ResolvedFunction{
				Fn:    executetest.FunctionExpression(t, `(left, right) => ({A: left.A})`),
				Scope: valuestest.Scope(),
			},
			left: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"A", "B"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "A", Type: flux.TString},
						{Label: "B", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 1.0, "a", "b"},
						{execute.Time(2), 2.0, "a", "b"},
						{execute.Time(3), 3.0, "a", "b"},
					},
				},
			},
			right: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"A", "B"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "A", Type: flux.TString},
						{Label: "B", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(1), 10.0, "a", "b"},
						{execute.Time(2), 20.0, "a", "b"},
						{execute.Time(3), 30.0, "a", "b"},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "left semijoin",
			// fn: (left, right) => left
			fn: interpreter.ResolvedFunction{
				Fn:    executetest.FunctionExpression(t, `(left, right) => left`),
				Scope: valuestest.Scope(),
			},
			left: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"_field"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_field", Type: flux.TString},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), "_value", 1.0},
						{execute.Time(2), "_value", 2.0},
						{execute.Time(3), "_value", 3.0},
					},
				},
			},
			right: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"_field"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_field", Type: flux.TString},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), "_value", 10.0},
						{execute.Time(2), "_value", 20.0},
						{execute.Time(3), "_value", 30.0},
					},
				},
			},
			want: []*executetest.Table{
				{
					KeyCols: []string{"_field"},
					ColMeta: []flux.ColMeta{
						{Label: "_field", Type: flux.TString},
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{"_value", execute.Time(1), 1.0},
						{"_value", execute.Time(2), 2.0},
						{"_value", execute.Time(3), 3.0},
					},
				},
			},
		},
		{
			name: "right semijoin",
			// fn: (left, right) => right
			fn: interpreter.ResolvedFunction{
				Fn:    executetest.FunctionExpression(t, `(left, right) => right`),
				Scope: valuestest.Scope(),
			},
			left: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"_field"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_field", Type: flux.TString},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), "_value", 1.0},
						{execute.Time(2), "_value", 2.0},
						{execute.Time(3), "_value", 3.0},
					},
				},
			},
			right: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"_field"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_field", Type: flux.TString},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), "_value", 10.0},
						{execute.Time(2), "_value", 20.0},
						{execute.Time(3), "_value", 30.0},
					},
				},
			},
			want: []*executetest.Table{
				{
					KeyCols: []string{"_field"},
					ColMeta: []flux.ColMeta{
						{Label: "_field", Type: flux.TString},
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{"_value", execute.Time(1), 10.0},
						{"_value", execute.Time(2), 20.0},
						{"_value", execute.Time(3), 30.0},
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			if tc.skip != "" {
				t.Skip(tc.skip)
			}
			l := execute.DatasetID(executetest.RandomDatasetID())
			r := execute.DatasetID(executetest.RandomDatasetID())

			cache := promql.NewMergeJoinCache(
				context.Background(),
				executetest.UnlimitedAllocator,
				tc.fn,
				l,
				r,
			)
			pjoin := promql.NewMergeJoinTransformation(
				executetest.NewDataset(executetest.RandomDatasetID()),
				cache,
			)

			for _, tbl := range tc.left {
				err := pjoin.Process(l, tbl)
				if err != nil && tc.wantErr {
					return
				}
				if err != nil && !tc.wantErr {
					t.Fatalf("error processing join: %v", err)
				}
			}
			for _, tbl := range tc.right {
				err := pjoin.Process(r, tbl)
				if err != nil && tc.wantErr {
					return
				}
				if err != nil && !tc.wantErr {
					t.Fatalf("error processing join: %v", err)
				}
			}

			got, err := executetest.TablesFromCache(cache)
			if err != nil && tc.wantErr {
				return
			}
			if err != nil && !tc.wantErr {
				t.Fatal(err)
			}
			if err == nil && tc.wantErr {
				t.Fatal("expected runtime error but got nothing")
			}

			executetest.NormalizeTables(got)
			executetest.NormalizeTables(tc.want)

			sort.Sort(executetest.SortedTables(got))
			sort.Sort(executetest.SortedTables(tc.want))

			if !cmp.Equal(tc.want, got) {
				t.Errorf("unexpected tables -want/+got\n%s", cmp.Diff(tc.want, got))
			}
		})
	}
}
