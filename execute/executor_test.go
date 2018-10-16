package execute_test

import (
	"context"
	"testing"
	"time"
	"math"

	"github.com/influxdata/flux/ast"
	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	_ "github.com/influxdata/flux/builtin"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/functions/transformations"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/planner/plantest"
	"github.com/influxdata/flux/planner"
	"go.uber.org/zap/zaptest"
)

func init() {
	execute.RegisterSource("from-test", executetest.CreateFromSource)
}

func TestExecutor_Execute(t *testing.T) {
	testcases := []struct {
		name string
		spec *plantest.PhysicalPlanSpec
		want map[string][]*executetest.Table
	}{
		{
			name: `from`,
			spec: &plantest.PhysicalPlanSpec{
				Nodes: []planner.PlanNode{
					planner.CreatePhysicalNode("from-test", executetest.NewFromProcedureSpec(
						[]*executetest.Table{&executetest.Table{
							KeyCols: []string{"_start", "_stop"},
							ColMeta: []flux.ColMeta{
								{Label: "_start", Type: flux.TTime},
								{Label: "_stop", Type: flux.TTime},
								{Label: "_time", Type: flux.TTime},
								{Label: "_value", Type: flux.TFloat},
							},
							Data: [][]interface{}{
								{execute.Time(0), execute.Time(5), execute.Time(0), 1.0},
								{execute.Time(0), execute.Time(5), execute.Time(1), 2.0},
								{execute.Time(0), execute.Time(5), execute.Time(2), 3.0},
								{execute.Time(0), execute.Time(5), execute.Time(3), 4.0},
								{execute.Time(0), execute.Time(5), execute.Time(4), 5.0},
							},
						}},
					)),
				},
				Results: map[string]int{"_result": 0},
			},
			want: map[string][]*executetest.Table{
				"_result": []*executetest.Table{{
					KeyCols: []string{"_start", "_stop"},
					ColMeta: []flux.ColMeta{
						{Label: "_start", Type: flux.TTime},
						{Label: "_stop", Type: flux.TTime},
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(0), execute.Time(5), execute.Time(0), 1.0},
						{execute.Time(0), execute.Time(5), execute.Time(1), 2.0},
						{execute.Time(0), execute.Time(5), execute.Time(2), 3.0},
						{execute.Time(0), execute.Time(5), execute.Time(3), 4.0},
						{execute.Time(0), execute.Time(5), execute.Time(4), 5.0},
					},
				}},
			},
		},
		{
			name: `from with filter`,
			spec: &plantest.PhysicalPlanSpec{
				Nodes: []planner.PlanNode{
					planner.CreatePhysicalNode("from-test", executetest.NewFromProcedureSpec(
						[]*executetest.Table{&executetest.Table{
							KeyCols: []string{"_start", "_stop"},
							ColMeta: []flux.ColMeta{
								{Label: "_start", Type: flux.TTime},
								{Label: "_stop", Type: flux.TTime},
								{Label: "_time", Type: flux.TTime},
								{Label: "_value", Type: flux.TFloat},
							},
							Data: [][]interface{}{
								{execute.Time(0), execute.Time(5), execute.Time(0), 1.0},
								{execute.Time(0), execute.Time(5), execute.Time(1), 2.0},
								{execute.Time(0), execute.Time(5), execute.Time(2), 3.0},
								{execute.Time(0), execute.Time(5), execute.Time(3), 4.0},
								{execute.Time(0), execute.Time(5), execute.Time(4), 5.0},
							},
						}},
					)),
					planner.CreatePhysicalNode("filter", &transformations.FilterProcedureSpec{
						Fn: &semantic.FunctionExpression{
							Params: []*semantic.FunctionParam{
								{
									Key: &semantic.Identifier{Name: "r"},
								},
							},
							Body: &semantic.BinaryExpression{
								Operator: ast.LessThanOperator,
								Left: &semantic.MemberExpression{
									Property: "_value",
									Object: &semantic.IdentifierExpression{
										Name: "r",
									},
								},
								Right: &semantic.FloatLiteral{Value: 2.5},
							},
						},
					}),
				},
				Edges: [][2]int{
					{0, 1},
				},
				Results: map[string]int{
					"_result": 1,
				},
			},
			want: map[string][]*executetest.Table{
				"_result": []*executetest.Table{{
					KeyCols: []string{"_start", "_stop"},
					ColMeta: []flux.ColMeta{
						{Label: "_start", Type: flux.TTime},
						{Label: "_stop", Type: flux.TTime},
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(0), execute.Time(5), execute.Time(0), 1.0},
						{execute.Time(0), execute.Time(5), execute.Time(1), 2.0},
					},
				}},
			},
		},
		{
			name: `from with filter with multiple tables`,
			spec: &plantest.PhysicalPlanSpec{
				Nodes: []planner.PlanNode{
					planner.CreatePhysicalNode("from-test", executetest.NewFromProcedureSpec(
						[]*executetest.Table{
							{
								KeyCols: []string{"_start", "_stop"},
								ColMeta: []flux.ColMeta{
									{Label: "_start", Type: flux.TTime},
									{Label: "_stop", Type: flux.TTime},
									{Label: "_time", Type: flux.TTime},
									{Label: "_value", Type: flux.TFloat},
								},
								Data: [][]interface{}{
									{execute.Time(0), execute.Time(5), execute.Time(0), 1.0},
									{execute.Time(0), execute.Time(5), execute.Time(1), 2.0},
									{execute.Time(0), execute.Time(5), execute.Time(2), 3.0},
									{execute.Time(0), execute.Time(5), execute.Time(3), 4.0},
									{execute.Time(0), execute.Time(5), execute.Time(4), 5.0},
								},
							},
							{
								KeyCols: []string{"_start", "_stop"},
								ColMeta: []flux.ColMeta{
									{Label: "_start", Type: flux.TTime},
									{Label: "_stop", Type: flux.TTime},
									{Label: "_time", Type: flux.TTime},
									{Label: "_value", Type: flux.TFloat},
								},
								Data: [][]interface{}{
									{execute.Time(5), execute.Time(10), execute.Time(5), 5.0},
									{execute.Time(5), execute.Time(10), execute.Time(6), 6.0},
									{execute.Time(5), execute.Time(10), execute.Time(7), 7.0},
									{execute.Time(5), execute.Time(10), execute.Time(8), 8.0},
									{execute.Time(5), execute.Time(10), execute.Time(9), 9.0},
								},
							},
						},
					)),
					planner.CreatePhysicalNode("filter", &transformations.FilterProcedureSpec{
						Fn: &semantic.FunctionExpression{
							Params: []*semantic.FunctionParam{
								{
									Key: &semantic.Identifier{Name: "r"},
								},
							},
							Body: &semantic.BinaryExpression{
								Operator: ast.LessThanOperator,
								Left: &semantic.MemberExpression{
									Property: "_value",
									Object: &semantic.IdentifierExpression{
										Name: "r",
									},
								},
								Right: &semantic.FloatLiteral{Value: 7.5},
							},
						},
					}),
				},
				Edges: [][2]int{
					{0, 1},
				},
				Results: map[string]int{
					"_result": 1,
				},
			},
			want: map[string][]*executetest.Table{
				"_result": []*executetest.Table{
					{
						KeyCols: []string{"_start", "_stop"},
						ColMeta: []flux.ColMeta{
							{Label: "_start", Type: flux.TTime},
							{Label: "_stop", Type: flux.TTime},
							{Label: "_time", Type: flux.TTime},
							{Label: "_value", Type: flux.TFloat},
						},
						Data: [][]interface{}{
							{execute.Time(0), execute.Time(5), execute.Time(0), 1.0},
							{execute.Time(0), execute.Time(5), execute.Time(1), 2.0},
							{execute.Time(0), execute.Time(5), execute.Time(2), 3.0},
							{execute.Time(0), execute.Time(5), execute.Time(3), 4.0},
							{execute.Time(0), execute.Time(5), execute.Time(4), 5.0},
						},
					},
					{
						KeyCols: []string{"_start", "_stop"},
						ColMeta: []flux.ColMeta{
							{Label: "_start", Type: flux.TTime},
							{Label: "_stop", Type: flux.TTime},
							{Label: "_time", Type: flux.TTime},
							{Label: "_value", Type: flux.TFloat},
						},
						Data: [][]interface{}{
							{execute.Time(5), execute.Time(10), execute.Time(5), 5.0},
							{execute.Time(5), execute.Time(10), execute.Time(6), 6.0},
							{execute.Time(5), execute.Time(10), execute.Time(7), 7.0},
						},
					},
				},
			},
		},
		{
			name: `multiple aggregates`,
			spec: &plantest.PhysicalPlanSpec{
				Nodes: []planner.PlanNode{
					planner.CreatePhysicalNode("from-test", executetest.NewFromProcedureSpec(
						[]*executetest.Table{
							{
								KeyCols: []string{"_start", "_stop"},
								ColMeta: []flux.ColMeta{
									{Label: "_start", Type: flux.TTime},
									{Label: "_stop", Type: flux.TTime},
									{Label: "_time", Type: flux.TTime},
									{Label: "_value", Type: flux.TFloat},
								},
								Data: [][]interface{}{
									{execute.Time(0), execute.Time(5), execute.Time(0), 1.0},
									{execute.Time(0), execute.Time(5), execute.Time(1), 2.0},
									{execute.Time(0), execute.Time(5), execute.Time(2), 3.0},
									{execute.Time(0), execute.Time(5), execute.Time(3), 4.0},
									{execute.Time(0), execute.Time(5), execute.Time(4), 5.0},
								},
							},
							{
								KeyCols: []string{"_start", "_stop"},
								ColMeta: []flux.ColMeta{
									{Label: "_start", Type: flux.TTime},
									{Label: "_stop", Type: flux.TTime},
									{Label: "_time", Type: flux.TTime},
									{Label: "_value", Type: flux.TFloat},
								},
								Data: [][]interface{}{
									{execute.Time(5), execute.Time(10), execute.Time(5), 5.0},
									{execute.Time(5), execute.Time(10), execute.Time(6), 6.0},
									{execute.Time(5), execute.Time(10), execute.Time(7), 7.0},
									{execute.Time(5), execute.Time(10), execute.Time(8), 8.0},
									{execute.Time(5), execute.Time(10), execute.Time(9), 9.0},
								},
							},
						},
					)),
					planner.CreatePhysicalNode("sum", &transformations.SumProcedureSpec{
						AggregateConfig: execute.DefaultAggregateConfig,
					}),
					planner.CreatePhysicalNode("mean", &transformations.MeanProcedureSpec{
						AggregateConfig: execute.DefaultAggregateConfig,
					}),

				},
				Edges: [][2]int{
					{0, 1},
					{0, 2},
				},
				Results: map[string]int{
					"sum": 1,
					"mean": 2,
				},
			},
			want: map[string][]*executetest.Table{
				"sum": []*executetest.Table{
					{
						KeyCols: []string{"_start", "_stop"},
						ColMeta: []flux.ColMeta{
							{Label: "_start", Type: flux.TTime},
							{Label: "_stop", Type: flux.TTime},
							{Label: "_value", Type: flux.TFloat},
						},
						Data: [][]interface{}{
							{execute.Time(0), execute.Time(5), 15.0},
						},
					},
					{
						KeyCols: []string{"_start", "_stop"},
						ColMeta: []flux.ColMeta{
							{Label: "_start", Type: flux.TTime},
							{Label: "_stop", Type: flux.TTime},
							{Label: "_value", Type: flux.TFloat},
						},
						Data: [][]interface{}{
							{execute.Time(5), execute.Time(10), 35.0},
						},
					},
				},
				"mean": []*executetest.Table{
					{
						KeyCols: []string{"_start", "_stop"},
						ColMeta: []flux.ColMeta{
							{Label: "_start", Type: flux.TTime},
							{Label: "_stop", Type: flux.TTime},
							{Label: "_value", Type: flux.TFloat},
						},
						Data: [][]interface{}{
							{execute.Time(0), execute.Time(5), 3.0},
						},
					},
					{
						KeyCols: []string{"_start", "_stop"},
						ColMeta: []flux.ColMeta{
							{Label: "_start", Type: flux.TTime},
							{Label: "_stop", Type: flux.TTime},
							{Label: "_value", Type: flux.TFloat},
						},
						Data: [][]interface{}{
							{execute.Time(5), execute.Time(10), 7.0},
						},
					},
				},
			},
		},
		{
			name: `diamond join`,
			spec: &plantest.PhysicalPlanSpec{
				Nodes: []planner.PlanNode{
					planner.CreatePhysicalNode("from-test", executetest.NewFromProcedureSpec(
						[]*executetest.Table{
							{
								KeyCols: []string{"_start", "_stop"},
								ColMeta: []flux.ColMeta{
									{Label: "_start", Type: flux.TTime},
									{Label: "_stop", Type: flux.TTime},
									{Label: "_time", Type: flux.TTime},
									{Label: "_value", Type: flux.TFloat},
								},
								Data: [][]interface{}{
									{execute.Time(0), execute.Time(5), execute.Time(0), 1.0},
									{execute.Time(0), execute.Time(5), execute.Time(1), 2.0},
									{execute.Time(0), execute.Time(5), execute.Time(2), 3.0},
									{execute.Time(0), execute.Time(5), execute.Time(3), 4.0},
									{execute.Time(0), execute.Time(5), execute.Time(4), 5.0},
								},
							},
							{
								KeyCols: []string{"_start", "_stop"},
								ColMeta: []flux.ColMeta{
									{Label: "_start", Type: flux.TTime},
									{Label: "_stop", Type: flux.TTime},
									{Label: "_time", Type: flux.TTime},
									{Label: "_value", Type: flux.TFloat},
								},
								Data: [][]interface{}{
									{execute.Time(5), execute.Time(10), execute.Time(5), 1.0},
									{execute.Time(5), execute.Time(10), execute.Time(6), 2.0},
									{execute.Time(5), execute.Time(10), execute.Time(7), 3.0},
									{execute.Time(5), execute.Time(10), execute.Time(8), 4.0},
									{execute.Time(5), execute.Time(10), execute.Time(9), 5.0},
								},
							},
							{
								KeyCols: []string{"_start", "_stop"},
								ColMeta: []flux.ColMeta{
									{Label: "_start", Type: flux.TTime},
									{Label: "_stop", Type: flux.TTime},
									{Label: "_time", Type: flux.TTime},
									{Label: "_value", Type: flux.TFloat},
								},
								Data: [][]interface{}{
									{execute.Time(10), execute.Time(15), execute.Time(10), 1.0},
									{execute.Time(10), execute.Time(15), execute.Time(11), 2.0},
									{execute.Time(10), execute.Time(15), execute.Time(12), 3.0},
									{execute.Time(10), execute.Time(15), execute.Time(13), 4.0},
									{execute.Time(10), execute.Time(15), execute.Time(14), 5.0},
								},
							},
						},
					)),
					planner.CreatePhysicalNode("sum", &transformations.SumProcedureSpec{
						AggregateConfig: execute.DefaultAggregateConfig,
					}),
					planner.CreatePhysicalNode("count", &transformations.CountProcedureSpec{
						AggregateConfig: execute.DefaultAggregateConfig,
					}),
					planner.CreatePhysicalNode("join", &transformations.MergeJoinProcedureSpec{
						On: []string{"_start", "_stop"},
						TableNames: map[planner.ProcedureID]string{
							planner.ProcedureIDFromOperationID("sum"): "a",
							planner.ProcedureIDFromOperationID("count"): "b",
						},
					}),

				},
				Edges: [][2]int{
					{0, 1},
					{0, 2},
					{1, 3},
					{2, 3},
				},
				Results: map[string]int{
					"_result": 3,
				},
			},
			want: map[string][]*executetest.Table{
				"_result": []*executetest.Table{
					{
						KeyCols: []string{"_start", "_stop"},
						ColMeta: []flux.ColMeta{
							{Label: "_start", Type: flux.TTime},
							{Label: "_stop", Type: flux.TTime},
							{Label: "_value_a", Type: flux.TFloat},
							{Label: "_value_b", Type: flux.TInt},
						},
						Data: [][]interface{}{
							{execute.Time(0), execute.Time(5), 15.0, int64(5)},
						},
					},
					{
						KeyCols: []string{"_start", "_stop"},
						ColMeta: []flux.ColMeta{
							{Label: "_start", Type: flux.TTime},
							{Label: "_stop", Type: flux.TTime},
							{Label: "_value_a", Type: flux.TFloat},
							{Label: "_value_b", Type: flux.TInt},
						},
						Data: [][]interface{}{
							{execute.Time(5), execute.Time(10), 15.0, int64(5)},
						},
					},
					{
						KeyCols: []string{"_start", "_stop"},
						ColMeta: []flux.ColMeta{
							{Label: "_start", Type: flux.TTime},
							{Label: "_stop", Type: flux.TTime},
							{Label: "_value_a", Type: flux.TFloat},
							{Label: "_value_b", Type: flux.TInt},
						},
						Data: [][]interface{}{
							{execute.Time(10), execute.Time(15), 15.0, int64(5)},
						},
					},
				},
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			tc.spec.Resources = flux.ResourceManagement{
				ConcurrencyQuota: 1,
				MemoryBytesQuota: math.MaxInt64,
			}

			tc.spec.Now = time.Now()

			// Construct physical query plan
			plan := plantest.CreatePhysicalPlanSpec(tc.spec)

			exe := execute.NewExecutor(nil, zaptest.NewLogger(t))
			results, err := exe.Execute(context.Background(), plan, executetest.UnlimitedAllocator)
			if err != nil {
				t.Fatal(err)
			}
			got := make(map[string][]*executetest.Table, len(results))
			for name, r := range results {
				if err := r.Tables().Do(func(tbl flux.Table) error {
					cb, err := executetest.ConvertTable(tbl)
					if err != nil {
						return err
					}
					got[name] = append(got[name], cb)
					return nil
				}); err != nil {
					t.Fatal(err)
				}
			}

			for _, g := range got {
				executetest.NormalizeTables(g)
			}
			for _, w := range tc.want {
				executetest.NormalizeTables(w)
			}

			if !cmp.Equal(got, tc.want) {
				t.Error("unexpected results -want/+got", cmp.Diff(tc.want, got))
			}
		})
	}
}
