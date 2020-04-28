package bigtable

import (
	"context"
	"fmt"
	"testing"
	"time"

	"cloud.google.com/go/bigtable"
	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/mock"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/stdlib/universe"
	"github.com/influxdata/flux/values"

	_ "github.com/influxdata/flux/stdlib/date"
	_ "github.com/influxdata/flux/stdlib/influxdata/influxdb"
	_ "github.com/influxdata/flux/stdlib/math"
	_ "github.com/influxdata/flux/stdlib/regexp"
	_ "github.com/influxdata/flux/stdlib/strings"
	_ "github.com/influxdata/flux/stdlib/system"
)

func init() {
	runtime.FinalizeBuiltIns()
}

func TestBigtableDecode(t *testing.T) {
	t.Run("Bigtable Mock RowReader", func(t *testing.T) {
		timestamp, _ := values.ParseTime("2019-08-015 09:20:00")
		var reader execute.RowReader = &MockRowReader{
			cursor: -1,
			rows: [][]values.Value{
				{
					values.NewString("1"),
					values.NewTime(timestamp),
					values.NewString("fam"),
					values.NewString("aa"),
					values.NewString("ba"),
				},
				{
					values.NewString("2"),
					values.NewTime(timestamp),
					values.NewString("fam"),
					values.NewString("ab"),
					values.NewString("bb"),
				},
				{
					values.NewString("3"),
					values.NewTime(timestamp),
					values.NewString("fam"),
					values.NewString("ac"),
					values.NewString("bc"),
				},
				{
					values.NewString("4"),
					values.NewTime(timestamp),
					values.NewString("fam"),
					values.NewString("ad"),
					values.NewString("bd"),
				},
			},
			columnNames: []string{"rowKey", "_time", "family", "a", "b"},
		}

		decoder := &BigtableDecoder{reader: &reader, administration: &mock.Administration{}}
		table, err := decoder.Decode(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		want := &executetest.Table{
			ColMeta: []flux.ColMeta{
				{Label: "rowKey", Type: flux.TString},
				{Label: "_time", Type: flux.TTime},
				{Label: "family", Type: flux.TString},
				{Label: "a", Type: flux.TString},
				{Label: "b", Type: flux.TString},
			},
			Data: [][]interface{}{
				{"1", timestamp, "fam", "aa", "ba"},
				{"2", timestamp, "fam", "ab", "bb"},
				{"3", timestamp, "fam", "ac", "bc"},
				{"4", timestamp, "fam", "ad", "bd"},
			},
		}

		if !cmp.Equal(table.Cols(), want.Cols()) {
			t.Fatalf("unexpected result -want/+got\n\n%s\n\n", cmp.Diff(want.Cols(), table.Cols()))
		}
		if !cmp.Equal(table.Key(), want.Key()) {
			t.Fatalf("unexpected result -want/+got\n\n%s\n\n", cmp.Diff(want.Key(), table.Key()))
		}
		if !cmp.Equal(table.Key().Cols(), []flux.ColMeta(nil)) {
			t.Fatalf("unexpected result -want/+got\n\n%s\n\n", cmp.Diff([]flux.ColMeta(nil), table.Key().Cols()))
		}

		buffer := execute.NewColListTableBuilder(table.Key(), executetest.UnlimitedAllocator)
		if err := execute.AddTableCols(table, buffer); err != nil {
			t.Fatal(err)
		}
		if err := execute.AppendTable(table, buffer); err != nil {
			t.Fatal(err)
		}

		wantBuffer := execute.NewColListTableBuilder(want.Key(), executetest.UnlimitedAllocator)
		if err := execute.AddTableCols(want, wantBuffer); err != nil {
			t.Fatal(err)
		}
		if err := execute.AppendTable(want, wantBuffer); err != nil {
			t.Fatal(err)
		}

		for i := 0; i < 4; i++ {
			want := wantBuffer.GetRow(i)
			got := buffer.GetRow(i)
			if !got.Equal(want) {
				t.Fatalf("unexpected result -want/+got:\n%s", cmp.Diff(want, got))
			}
		}
	})
}

func TestNodeRewrite(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		name        string
		queryNode   plan.Node
		rewriteNode plan.Node
		rewriteFunc func(plan.Node, plan.Node) (plan.Node, bool)
		wantNode    plan.Node
		wantBool    bool
	}{
		{
			name:      "|> filter(fn: (r) => r.rowKey == ... )",
			queryNode: &plan.PhysicalPlanNode{Spec: &FromBigtableProcedureSpec{}},
			rewriteNode: &plan.PhysicalPlanNode{
				Spec: &universe.FilterProcedureSpec{
					Fn: interpreter.ResolvedFunction{
						Fn: executetest.FunctionExpression(t, `(r) => r.rowKey == "single row"`),
					},
				},
			},
			rewriteFunc: AddFilterToNode,
			wantNode:    &plan.PhysicalPlanNode{Spec: &FromBigtableProcedureSpec{RowSet: bigtable.SingleRow("single row")}},
			wantBool:    true,
		},
		{
			name:      "|> filter(fn: (r) => r.rowKey >= ... )",
			queryNode: &plan.PhysicalPlanNode{Spec: &FromBigtableProcedureSpec{}},
			rewriteNode: &plan.PhysicalPlanNode{
				Spec: &universe.FilterProcedureSpec{
					Fn: interpreter.ResolvedFunction{
						Fn: executetest.FunctionExpression(t, `(r) => r.rowKey >= "greater than or equal"`),
					},
				},
			},
			rewriteFunc: AddFilterToNode,
			wantNode:    &plan.PhysicalPlanNode{Spec: &FromBigtableProcedureSpec{RowSet: bigtable.InfiniteRange("greater than or equal")}},
			wantBool:    true,
		},
		{
			name:      "|> filter(fn: (r) => r._time >= ... )",
			queryNode: &plan.PhysicalPlanNode{Spec: &FromBigtableProcedureSpec{Filter: bigtable.PassAllFilter()}},
			rewriteNode: &plan.PhysicalPlanNode{
				Spec: &universe.FilterProcedureSpec{
					Fn: interpreter.ResolvedFunction{
						Fn: executetest.FunctionExpression(t, `(r) => r._time >= %s`, now.Format(time.RFC3339Nano)),
					},
				},
			},
			rewriteFunc: AddFilterToNode,
			wantNode:    &plan.PhysicalPlanNode{Spec: &FromBigtableProcedureSpec{Filter: bigtable.ChainFilters(bigtable.PassAllFilter(), bigtable.TimestampRangeFilter(now, time.Time{}))}},
			wantBool:    true,
		},
		{
			name:      "|> filter(fn: (r) => r._time < ... )",
			queryNode: &plan.PhysicalPlanNode{Spec: &FromBigtableProcedureSpec{Filter: bigtable.PassAllFilter()}},
			rewriteNode: &plan.PhysicalPlanNode{
				Spec: &universe.FilterProcedureSpec{
					Fn: interpreter.ResolvedFunction{
						Fn: executetest.FunctionExpression(t, `(r) => r._time < %s`, now.Format(time.RFC3339Nano)),
					},
				},
			},
			rewriteFunc: AddFilterToNode,
			wantNode:    &plan.PhysicalPlanNode{Spec: &FromBigtableProcedureSpec{Filter: bigtable.ChainFilters(bigtable.PassAllFilter(), bigtable.TimestampRangeFilter(time.Time{}, now))}},
			wantBool:    true,
		},
		{
			name:      "|> filter(fn: (r) => r.rowKey >= ... and r.rowKey < ...)",
			queryNode: &plan.PhysicalPlanNode{Spec: &FromBigtableProcedureSpec{Filter: bigtable.PassAllFilter()}},
			rewriteNode: &plan.PhysicalPlanNode{
				Spec: &universe.FilterProcedureSpec{
					Fn: interpreter.ResolvedFunction{
						Fn: executetest.FunctionExpression(t, `(r) => r.rowKey >= "start" and r.rowKey < "end"`),
					},
				},
			},
			rewriteFunc: AddFilterToNode,
			wantNode:    &plan.PhysicalPlanNode{Spec: &FromBigtableProcedureSpec{RowSet: bigtable.NewRange("start", "end"), Filter: bigtable.PassAllFilter()}},
			wantBool:    true,
		},
		{
			name:      "|> filter(fn: (r) => r.rowKey < ... and r.rowKey >= ...)",
			queryNode: &plan.PhysicalPlanNode{Spec: &FromBigtableProcedureSpec{Filter: bigtable.PassAllFilter()}},
			rewriteNode: &plan.PhysicalPlanNode{
				Spec: &universe.FilterProcedureSpec{
					Fn: interpreter.ResolvedFunction{
						Fn: executetest.FunctionExpression(t, `(r) => r.rowKey < "end" and r.rowKey >= "start"`),
					},
				},
			},
			rewriteFunc: AddFilterToNode,
			wantNode:    &plan.PhysicalPlanNode{Spec: &FromBigtableProcedureSpec{RowSet: bigtable.NewRange("start", "end"), Filter: bigtable.PassAllFilter()}},
			wantBool:    true,
		},
		{
			name:      "|> filter(fn: (r) => r._time >= ...)",
			queryNode: &plan.PhysicalPlanNode{Spec: &FromBigtableProcedureSpec{Filter: bigtable.PassAllFilter()}},
			rewriteNode: &plan.PhysicalPlanNode{
				Spec: &universe.FilterProcedureSpec{
					Fn: interpreter.ResolvedFunction{
						Fn: executetest.FunctionExpression(t, `(r) => r._time >= %s`, now.Format(time.RFC3339Nano)),
					},
				},
			},
			rewriteFunc: AddFilterToNode,
			wantNode:    &plan.PhysicalPlanNode{Spec: &FromBigtableProcedureSpec{Filter: bigtable.ChainFilters(bigtable.PassAllFilter(), bigtable.TimestampRangeFilter(now, time.Time{}))}},
			wantBool:    true,
		},
		{
			name:      "|> filter(fn: (r) => r._time >= ...)",
			queryNode: &plan.PhysicalPlanNode{Spec: &FromBigtableProcedureSpec{Filter: bigtable.PassAllFilter()}},
			rewriteNode: &plan.PhysicalPlanNode{
				Spec: &universe.FilterProcedureSpec{
					Fn: interpreter.ResolvedFunction{
						Fn: executetest.FunctionExpression(t, `(r) => r._time >= %s`, now.Format(time.RFC3339Nano)),
					},
				},
			},
			rewriteFunc: AddFilterToNode,
			wantNode:    &plan.PhysicalPlanNode{Spec: &FromBigtableProcedureSpec{Filter: bigtable.ChainFilters(bigtable.PassAllFilter(), bigtable.TimestampRangeFilter(now, time.Time{}))}},
			wantBool:    true,
		},
		{
			name:      "|> filter(fn: (r) => strings.hasPrefix(v: r.rowKey, prefix: ...)",
			queryNode: &plan.PhysicalPlanNode{Spec: &FromBigtableProcedureSpec{Filter: bigtable.PassAllFilter()}},
			rewriteNode: &plan.PhysicalPlanNode{
				Spec: &universe.FilterProcedureSpec{
					Fn: interpreter.ResolvedFunction{
						Fn: executetest.FunctionExpression(t, `
import "strings"
(r) => strings.hasPrefix(v: r.rowKey, prefix: "the prefix")
`),
					},
				},
			},
			rewriteFunc: AddFilterToNode,
			wantNode:    &plan.PhysicalPlanNode{Spec: &FromBigtableProcedureSpec{RowSet: bigtable.PrefixRange("the prefix"), Filter: bigtable.PassAllFilter()}},
			wantBool:    true,
		},
		{
			name:      "|> filter(fn: (r) => r.family == ...)",
			queryNode: &plan.PhysicalPlanNode{Spec: &FromBigtableProcedureSpec{Filter: bigtable.PassAllFilter()}},
			rewriteNode: &plan.PhysicalPlanNode{
				Spec: &universe.FilterProcedureSpec{
					Fn: interpreter.ResolvedFunction{
						Fn: executetest.FunctionExpression(t, `(r) => r.family == "family"`),
					},
				},
			},
			rewriteFunc: AddFilterToNode,
			wantNode:    &plan.PhysicalPlanNode{Spec: &FromBigtableProcedureSpec{Filter: bigtable.ChainFilters(bigtable.PassAllFilter(), bigtable.FamilyFilter("family"))}},
			wantBool:    true,
		},
		{
			name:        "|> limit(n: ...)",
			queryNode:   &plan.PhysicalPlanNode{Spec: &FromBigtableProcedureSpec{Filter: bigtable.PassAllFilter(), ReadOptions: make([]bigtable.ReadOption, 0)}},
			rewriteNode: &plan.PhysicalPlanNode{Spec: &universe.LimitProcedureSpec{N: 4, Offset: 0}},
			rewriteFunc: AddLimitToNode,
			wantNode:    &plan.PhysicalPlanNode{Spec: &FromBigtableProcedureSpec{ReadOptions: []bigtable.ReadOption{bigtable.LimitRows(4)}, Filter: bigtable.PassAllFilter()}},
			wantBool:    true,
		},
		{
			name:        "|> limit(n: ..., offset: 2)",
			queryNode:   &plan.PhysicalPlanNode{Spec: &FromBigtableProcedureSpec{Filter: bigtable.PassAllFilter(), ReadOptions: make([]bigtable.ReadOption, 0)}},
			rewriteNode: &plan.PhysicalPlanNode{Spec: &universe.LimitProcedureSpec{N: 4, Offset: 2}},
			rewriteFunc: AddLimitToNode,
			wantNode:    &plan.PhysicalPlanNode{Spec: &universe.LimitProcedureSpec{N: 4, Offset: 2}},
			wantBool:    false,
		},
	}

	rowRangeTransformer := cmp.Transformer("", func(in bigtable.RowRange) string {
		return in.String()
	})

	filterTransformer := cmp.Transformer("", func(in bigtable.Filter) string {
		if in != nil {
			return in.String()
		}
		return ""
	})

	readOptionTransformer := cmp.Transformer("", func(in []bigtable.ReadOption) int {
		return len(in)
	})

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			gotNode, gotBool := tc.rewriteFunc(tc.queryNode, tc.rewriteNode)
			if gotBool != tc.wantBool {
				t.Fatalf("unexpected result -want/+got\n\n%s\n\n", cmp.Diff(gotBool, tc.wantBool))
			}
			if !cmp.Equal(tc.wantNode.ProcedureSpec(), gotNode.ProcedureSpec(), rowRangeTransformer, filterTransformer, readOptionTransformer) {
				t.Fatalf("unexpected result -want/+got\n\n%s\n\n", cmp.Diff(tc.wantNode.ProcedureSpec(), gotNode.ProcedureSpec(), rowRangeTransformer, filterTransformer, readOptionTransformer))
			}
		})
	}
}

type MockRowReader struct {
	cursor      int
	rows        [][]values.Value
	columnNames []string
}

func (m *MockRowReader) Next() bool {
	m.cursor++
	return m.cursor < len(m.rows)
}

func (m *MockRowReader) GetNextRow() ([]values.Value, error) {
	a := len(m.rows)
	if m.cursor >= a {
		return nil, fmt.Errorf("out of range")
	}
	return m.rows[m.cursor], nil
}

func (m *MockRowReader) ColumnNames() []string {
	return m.columnNames
}

func (m *MockRowReader) ColumnTypes() []flux.ColType {
	return nil
}

func (m *MockRowReader) SetColumns([]interface{}) {}

func (m *MockRowReader) Close() error { return nil }
