package execute_test

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/ast"
	_ "github.com/influxdata/flux/builtin"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/plan/plantest"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/stdlib/universe"
	"go.uber.org/zap/zaptest"
)

func init() {
	execute.RegisterSource(executetest.FromTestKind, executetest.CreateFromSource)
	execute.RegisterSource(executetest.AllocatingFromTestKind, executetest.CreateAllocatingFromSource)
	execute.RegisterTransformation(executetest.ToTestKind, executetest.CreateToTransformation)
	plan.RegisterProcedureSpecWithSideEffect(executetest.ToTestKind, executetest.NewToProcedure, executetest.ToTestKind)
}

func TestExecutor_Execute(t *testing.T) {
	testcases := []struct {
		name      string
		spec      *plantest.PlanSpec
		want      map[string][]*executetest.Table
		allocator *memory.Allocator
		wantErr   error
	}{
		{
			name: `from`,
			spec: &plantest.PlanSpec{
				Nodes: []plan.Node{
					plan.CreatePhysicalNode("from-test", executetest.NewFromProcedureSpec(
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
					plan.CreatePhysicalNode("yield", executetest.NewYieldProcedureSpec("_result")),
				},
				Edges: [][2]int{
					{0, 1},
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
						{execute.Time(0), execute.Time(5), execute.Time(2), 3.0},
						{execute.Time(0), execute.Time(5), execute.Time(3), 4.0},
						{execute.Time(0), execute.Time(5), execute.Time(4), 5.0},
					},
				}},
			},
		},
		{
			name: `from with filter`,
			spec: &plantest.PlanSpec{
				Nodes: []plan.Node{
					plan.CreatePhysicalNode("from-test", executetest.NewFromProcedureSpec(
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
					plan.CreatePhysicalNode("filter", &universe.FilterProcedureSpec{
						Fn: interpreter.ResolvedFunction{
							Fn: &semantic.FunctionExpression{
								Block: &semantic.FunctionBlock{
									Parameters: &semantic.FunctionParameters{
										List: []*semantic.FunctionParameter{
											{
												Key: &semantic.Identifier{Name: "r"},
											},
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
							},
							Scope: flux.Prelude(),
						},
					}),
					plan.CreatePhysicalNode("yield", executetest.NewYieldProcedureSpec("_result")),
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
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
			spec: &plantest.PlanSpec{
				Nodes: []plan.Node{
					plan.CreatePhysicalNode("from-test", executetest.NewFromProcedureSpec(
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
					plan.CreatePhysicalNode("filter", &universe.FilterProcedureSpec{
						Fn: interpreter.ResolvedFunction{
							Scope: flux.Prelude(),
							Fn: &semantic.FunctionExpression{
								Block: &semantic.FunctionBlock{
									Parameters: &semantic.FunctionParameters{
										List: []*semantic.FunctionParameter{
											{
												Key: &semantic.Identifier{Name: "r"},
											},
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
							},
						},
					}),
					plan.CreatePhysicalNode("yield", executetest.NewYieldProcedureSpec("_result")),
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
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
			spec: &plantest.PlanSpec{
				Nodes: []plan.Node{
					plan.CreatePhysicalNode("from-test", executetest.NewFromProcedureSpec(
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
					plan.CreatePhysicalNode("sum", &universe.SumProcedureSpec{
						AggregateConfig: execute.DefaultAggregateConfig,
					}),
					plan.CreatePhysicalNode("mean", &universe.MeanProcedureSpec{
						AggregateConfig: execute.DefaultAggregateConfig,
					}),
					plan.CreatePhysicalNode("yield", executetest.NewYieldProcedureSpec("sum")),
					plan.CreatePhysicalNode("yield", executetest.NewYieldProcedureSpec("mean")),
				},
				Edges: [][2]int{
					{0, 1},
					{0, 2},
					{1, 3},
					{2, 4},
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
			spec: &plantest.PlanSpec{
				Nodes: []plan.Node{
					plan.CreatePhysicalNode("from-test", executetest.NewFromProcedureSpec(
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
					plan.CreatePhysicalNode("sum", &universe.SumProcedureSpec{
						AggregateConfig: execute.DefaultAggregateConfig,
					}),
					plan.CreatePhysicalNode("count", &universe.CountProcedureSpec{
						AggregateConfig: execute.DefaultAggregateConfig,
					}),
					plan.CreatePhysicalNode("join", &universe.MergeJoinProcedureSpec{
						On:         []string{"_start", "_stop"},
						TableNames: []string{"a", "b"},
					}),
					plan.CreatePhysicalNode("yield", executetest.NewYieldProcedureSpec("_result")),
				},
				Edges: [][2]int{
					{0, 1},
					{0, 2},
					{1, 3},
					{2, 3},
					{3, 4},
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
		{
			name: "yield with successor",
			spec: &plantest.PlanSpec{
				Nodes: []plan.Node{
					plan.CreatePhysicalNode("from-test", executetest.NewFromProcedureSpec(
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
					plan.CreatePhysicalNode("yield0", executetest.NewYieldProcedureSpec("from")),
					plan.CreatePhysicalNode("sum", &universe.SumProcedureSpec{
						AggregateConfig: execute.DefaultAggregateConfig,
					}),
					plan.CreatePhysicalNode("yield1", executetest.NewYieldProcedureSpec("sum")),
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
					{2, 3},
				},
			},
			want: map[string][]*executetest.Table{
				"from": []*executetest.Table{{
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
				"sum": []*executetest.Table{{
					KeyCols: []string{"_start", "_stop"},
					ColMeta: []flux.ColMeta{
						{Label: "_start", Type: flux.TTime},
						{Label: "_stop", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(0), execute.Time(5), 15.0},
					},
				}},
			},
		},
		{
			name: "adjacent yields",
			spec: &plantest.PlanSpec{
				Nodes: []plan.Node{
					plan.CreatePhysicalNode("from-test", executetest.NewFromProcedureSpec(
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
					plan.CreatePhysicalNode("sum", &universe.SumProcedureSpec{
						AggregateConfig: execute.DefaultAggregateConfig,
					}),
					plan.CreatePhysicalNode("yield0", executetest.NewYieldProcedureSpec("sum0")),
					plan.CreatePhysicalNode("yield1", executetest.NewYieldProcedureSpec("sum1")),
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
					{2, 3},
				},
			},
			want: map[string][]*executetest.Table{
				"sum0": []*executetest.Table{{
					KeyCols: []string{"_start", "_stop"},
					ColMeta: []flux.ColMeta{
						{Label: "_start", Type: flux.TTime},
						{Label: "_stop", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(0), execute.Time(5), 15.0},
					},
				}},
				"sum1": []*executetest.Table{{
					KeyCols: []string{"_start", "_stop"},
					ColMeta: []flux.ColMeta{
						{Label: "_start", Type: flux.TTime},
						{Label: "_stop", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(0), execute.Time(5), 15.0},
					},
				}},
			},
		},
		{
			name: "terminal output function",
			spec: &plantest.PlanSpec{
				Nodes: []plan.Node{
					plan.CreatePhysicalNode("from-test", executetest.NewFromProcedureSpec(
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
					plan.CreatePhysicalNode("to", &executetest.ToProcedureSpec{}),
				},
				Edges: [][2]int{{0, 1}},
			},
			want: map[string][]*executetest.Table{
				"to": []*executetest.Table{&executetest.Table{
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
			name: `yield preserves nulls`,
			spec: &plantest.PlanSpec{
				Nodes: []plan.Node{
					plan.CreatePhysicalNode("from-test", executetest.NewFromProcedureSpec(
						[]*executetest.Table{&executetest.Table{
							KeyCols: []string{"_start", "_stop"},
							ColMeta: []flux.ColMeta{
								{Label: "_start", Type: flux.TTime},
								{Label: "_stop", Type: flux.TTime},
								{Label: "_time", Type: flux.TTime},
								{Label: "_value", Type: flux.TFloat},
								{Label: "tag", Type: flux.TString},
								{Label: "valid", Type: flux.TBool},
							},
							Data: [][]interface{}{
								{nil, execute.Time(5), execute.Time(0), nil, nil, true},
								{execute.Time(0), nil, execute.Time(1), nil, "t0", nil},
								{execute.Time(0), nil, execute.Time(2), 3.0, "t0", nil},
								{execute.Time(0), execute.Time(5), execute.Time(3), nil, "t0", false},
								{execute.Time(0), nil, execute.Time(4), nil, "t0", true},
							},
						}},
					)),
					plan.CreatePhysicalNode("yield", &universe.YieldProcedureSpec{Name: "_result"}),
				},
				Edges: [][2]int{
					{0, 1},
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
						{Label: "tag", Type: flux.TString},
						{Label: "valid", Type: flux.TBool},
					},
					Data: [][]interface{}{
						{nil, execute.Time(5), execute.Time(0), nil, nil, true},
						{execute.Time(0), nil, execute.Time(1), nil, "t0", nil},
						{execute.Time(0), nil, execute.Time(2), 3.0, "t0", nil},
						{execute.Time(0), execute.Time(5), execute.Time(3), nil, "t0", false},
						{execute.Time(0), nil, execute.Time(4), nil, "t0", true},
					},
				}},
			},
		},
		{
			name: "memory limit exceeded",
			spec: &plantest.PlanSpec{
				Nodes: []plan.Node{
					plan.CreatePhysicalNode("allocating-from-test", &executetest.AllocatingFromProcedureSpec{ByteCount: 65}),
					plan.CreatePhysicalNode("yield", &universe.YieldProcedureSpec{Name: "_result"}),
				},
				Edges: [][2]int{
					{0, 1},
				},
			},
			allocator: &memory.Allocator{Limit: func(v int64) *int64 { return &v }(64)},
			wantErr: &flux.Error{
				Code: codes.ResourceExhausted,
				Err: memory.LimitExceededError{
					Limit:  64,
					Wanted: 65,
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
			plan := plantest.CreatePlanSpec(tc.spec)

			exe := execute.NewExecutor(nil, zaptest.NewLogger(t))

			alloc := tc.allocator
			if alloc == nil {
				alloc = executetest.UnlimitedAllocator
			}

			// Execute the query and preserve any error returned
			results, _, err := exe.Execute(context.Background(), plan, alloc)
			var got map[string][]*executetest.Table
			if err == nil {
				got = make(map[string][]*executetest.Table, len(results))
				for name, r := range results {
					if err = r.Tables().Do(func(tbl flux.Table) error {
						cb, err := executetest.ConvertTable(tbl)
						if err != nil {
							return err
						}
						got[name] = append(got[name], cb)
						return nil
					}); err != nil {
						break
					}
				}
			}

			if tc.wantErr == nil && err != nil {
				t.Fatal(err)
			}

			if tc.wantErr != nil {
				if err == nil {
					t.Fatalf(`expected an error "%v" but got none`, tc.wantErr)
				}

				if diff := cmp.Diff(tc.wantErr, err); diff != "" {
					t.Fatalf("unexpected error: -want/+got: %v", diff)
				}
				return
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
