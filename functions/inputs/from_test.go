package inputs_test

import (
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/semantic"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/functions/inputs"
	"github.com/influxdata/flux/functions/transformations"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/plan/plantest"
	"github.com/influxdata/flux/semantic/semantictest"
)

type fromTestAttrs struct {
	ID   plan.NodeID
	Spec plan.ProcedureSpec
}

func yield(name string) *transformations.YieldProcedureSpec {
	return &transformations.YieldProcedureSpec{Name: name}
}

func fluxTime(t int64) flux.Time {
	return flux.Time{
		Absolute: time.Unix(0, t).UTC(),
	}
}

func makeFilterFn(exprs ...semantic.Expression) *semantic.FunctionExpression {
	body := semantic.ExprsToConjunction(exprs...)
	return &semantic.FunctionExpression{
		Params: []*semantic.FunctionParam{{Key: &semantic.Identifier{Name: "r"}}},
		Body:   body,
	}
}

func TestFrom_PlannerTransformationRules(t *testing.T) {
	var (
		fromWithBounds = &inputs.FromProcedureSpec{
			BoundsSet: true,
			Bounds: flux.Bounds{
				Start: fluxTime(5),
				Stop:  fluxTime(10),
			},
		}
		fromWithIntersectedBounds = &inputs.FromProcedureSpec{
			BoundsSet: true,
			Bounds: flux.Bounds{
				Start: fluxTime(9),
				Stop:  fluxTime(10),
			},
		}
		rangeWithBounds = &transformations.RangeProcedureSpec{
			Bounds: flux.Bounds{
				Start: fluxTime(5),
				Stop:  fluxTime(10),
			},
		}
		rangeWithDifferentBounds = &transformations.RangeProcedureSpec{
			Bounds: flux.Bounds{
				Start: fluxTime(9),
				Stop:  fluxTime(14),
			},
		}
		from  = &inputs.FromProcedureSpec{}
		mean  = &transformations.MeanProcedureSpec{}
		count = &transformations.CountProcedureSpec{}

		pushableExpr1 = &semantic.BinaryExpression{Operator: ast.EqualOperator,
			Left:  &semantic.MemberExpression{Object: &semantic.IdentifierExpression{Name: "r"}, Property: "_measurement"},
			Right: &semantic.StringLiteral{Value: "cpu"}}

		pushableExpr2 = &semantic.BinaryExpression{Operator: ast.EqualOperator,
			Left:  &semantic.MemberExpression{Object: &semantic.IdentifierExpression{Name: "r"}, Property: "_field"},
			Right: &semantic.StringLiteral{Value: "cpu"}}

		unpushableExpr = &semantic.BinaryExpression{Operator: ast.LessThanOperator,
			Left:  &semantic.MemberExpression{Object: &semantic.IdentifierExpression{Name: "r"}, Property: "_value"},
			Right: &semantic.FloatLiteral{Value: 0.5}}

		statementFn = &semantic.FunctionExpression{
			Params: []*semantic.FunctionParam{{Key: &semantic.Identifier{Name: "r"}}},
			Body: &semantic.ReturnStatement{Argument: &semantic.BooleanLiteral{Value: true}},
		}
	)

	tests := []struct {
		name   string
		rules  []plan.Rule
		before *plantest.PlanSpec
		after  *plantest.PlanSpec
	}{
		{
			name: "from range",
			// from -> range  =>  from
			rules: []plan.Rule{&inputs.MergeFromRangeRule{}},
			before: &plantest.PlanSpec{
				Nodes: []plan.PlanNode{
					plan.CreatePhysicalNode("from", from),
					plan.CreatePhysicalNode("range", rangeWithBounds),
				},
				Edges: [][2]int{{0, 1}},
			},
			after: &plantest.PlanSpec{
				Nodes: []plan.PlanNode{
					plan.CreatePhysicalNode("merged_from_range", fromWithBounds),
				},
			},
		},
		{
			name: "from range with successor node",
			// from -> range -> count  =>  from -> count
			rules: []plan.Rule{&inputs.MergeFromRangeRule{}},
			before: &plantest.PlanSpec{
				Nodes: []plan.PlanNode{
					plan.CreatePhysicalNode("from", from),
					plan.CreatePhysicalNode("range", rangeWithBounds),
					plan.CreatePhysicalNode("count", count),
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
				},
			},
			after: &plantest.PlanSpec{
				Nodes: []plan.PlanNode{
					plan.CreatePhysicalNode("merged_from_range", fromWithBounds),
					plan.CreatePhysicalNode("count", count),
				},
				Edges: [][2]int{{0, 1}},
			},
		},
		{
			name: "from with multiple ranges",
			// from -> range -> range  =>  from
			rules: []plan.Rule{&inputs.MergeFromRangeRule{}},
			before: &plantest.PlanSpec{
				Nodes: []plan.PlanNode{
					plan.CreatePhysicalNode("from", from),
					plan.CreatePhysicalNode("range0", rangeWithBounds),
					plan.CreatePhysicalNode("range1", rangeWithDifferentBounds),
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
				},
			},
			after: &plantest.PlanSpec{
				Nodes: []plan.PlanNode{
					plan.CreatePhysicalNode("merged_merged_from_range0_range1", fromWithIntersectedBounds),
				},
			},
		},
		{
			name: "from range with multiple successor node",
			// count      mean
			//     \     /          count     mean
			//      range       =>      \    /
			//        |                  from
			//       from
			rules: []plan.Rule{&inputs.MergeFromRangeRule{}},
			before: &plantest.PlanSpec{
				Nodes: []plan.PlanNode{
					plan.CreatePhysicalNode("from", from),
					plan.CreatePhysicalNode("range", rangeWithBounds),
					plan.CreatePhysicalNode("count", count),
					plan.CreatePhysicalNode("yield0", yield("count")),
					plan.CreatePhysicalNode("mean", mean),
					plan.CreatePhysicalNode("yield1", yield("mean")),
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
					{2, 3},
					{1, 4},
					{4, 5},
				},
			},
			after: &plantest.PlanSpec{
				Nodes: []plan.PlanNode{
					plan.CreatePhysicalNode("merged_from_range", fromWithBounds),
					plan.CreatePhysicalNode("count", count),
					plan.CreatePhysicalNode("yield0", yield("count")),
					plan.CreatePhysicalNode("mean", mean),
					plan.CreatePhysicalNode("yield1", yield("mean")),
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
					{0, 3},
					{3, 4},
				},
			},
		},
		{
			name: "cannot push range into from",
			// range    count                                      range    count
			//     \    /       =>   cannot push range into a   =>     \    /
			//      from           from with multiple sucessors         from
			rules: []plan.Rule{&inputs.MergeFromRangeRule{}},
			before: &plantest.PlanSpec{
				Nodes: []plan.PlanNode{
					plan.CreatePhysicalNode("from", from),
					plan.CreatePhysicalNode("range", rangeWithBounds),
					plan.CreatePhysicalNode("yield0", yield("range")),
					plan.CreatePhysicalNode("count", count),
					plan.CreatePhysicalNode("yield1", yield("count")),
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
					{0, 3},
					{3, 4},
				},
			},
			after: &plantest.PlanSpec{
				Nodes: []plan.PlanNode{
					plan.CreatePhysicalNode("from", from),
					plan.CreatePhysicalNode("range", rangeWithBounds),
					plan.CreatePhysicalNode("yield0", yield("range")),
					plan.CreatePhysicalNode("count", count),
					plan.CreatePhysicalNode("yield1", yield("count")),
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
					{0, 3},
					{3, 4},
				},
			},
		},
		{
			name: "from filter",
			// from -> filter  =>  from
			rules: []plan.Rule{inputs.MergeFromFilterRule{}},
			before: &plantest.PlanSpec{
				Nodes: []plan.PlanNode{
					plan.CreatePhysicalNode("from", from),
					plan.CreatePhysicalNode("filter", &transformations.FilterProcedureSpec{Fn: makeFilterFn(pushableExpr1)}),
				},
				Edges: [][2]int{
					{0, 1},
				},
			},
			after: &plantest.PlanSpec{
				Nodes: []plan.PlanNode{
					plan.CreatePhysicalNode("merged_from_filter", &inputs.FromProcedureSpec{
						FilterSet: true,
						Filter:    makeFilterFn(pushableExpr1),
					}),
				},
			},
		},
		{
			name: "from filter filter",
			// from -> filter -> filter  =>  from    (rule applied twice)
			rules: []plan.Rule{inputs.MergeFromFilterRule{}},
			before: &plantest.PlanSpec{
				Nodes: []plan.PlanNode{
					plan.CreatePhysicalNode("from", from),
					plan.CreatePhysicalNode("filter1", &transformations.FilterProcedureSpec{Fn: makeFilterFn(pushableExpr1)}),
					plan.CreatePhysicalNode("filter2", &transformations.FilterProcedureSpec{Fn: makeFilterFn(pushableExpr2)}),
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
				},
			},
			after: &plantest.PlanSpec{
				Nodes: []plan.PlanNode{
					plan.CreatePhysicalNode("merged_merged_from_filter1_filter2",
						&inputs.FromProcedureSpec{
							FilterSet: true,
							Filter:    makeFilterFn(pushableExpr1, pushableExpr2),
						}),
				},
			},
		},
		{
			name: "from partially-pushable-filter",
			// from -> partially-pushable-filter  =>  from-with-filter -> unpushable-filter
			rules: []plan.Rule{inputs.MergeFromFilterRule{}},
			before: &plantest.PlanSpec{
				Nodes: []plan.PlanNode{
					plan.CreatePhysicalNode("from", from),
					plan.CreatePhysicalNode("filter", &transformations.FilterProcedureSpec{Fn: makeFilterFn(pushableExpr1, unpushableExpr)}),
				},
				Edges: [][2]int{
					{0, 1},
				},
			},
			after: &plantest.PlanSpec{
				Nodes: []plan.PlanNode{
					plan.CreatePhysicalNode("from",
						&inputs.FromProcedureSpec{
							FilterSet: true,
							Filter:    makeFilterFn(pushableExpr1),
						}),
					plan.CreatePhysicalNode("filter", &transformations.FilterProcedureSpec{Fn: makeFilterFn(unpushableExpr)}),
				},
				Edges: [][2]int{
					{0, 1},
				},
			},
		},
		{
			name: "from range filter",
			// from -> range -> filter  =>  from
			rules: []plan.Rule{inputs.MergeFromFilterRule{}, inputs.MergeFromRangeRule{}},
			before: &plantest.PlanSpec{
				Nodes: []plan.PlanNode{
					plan.CreatePhysicalNode("from", from),
					plan.CreatePhysicalNode("range", rangeWithBounds),
					plan.CreatePhysicalNode("filter", &transformations.FilterProcedureSpec{Fn: makeFilterFn(pushableExpr1)}),
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
				},
			},
			after: &plantest.PlanSpec{
				Nodes: []plan.PlanNode{
					plan.CreatePhysicalNode("merged_merged_from_range_filter", &inputs.FromProcedureSpec{
						FilterSet: true,
						Filter:    makeFilterFn(pushableExpr1),
						BoundsSet: true,
						Bounds: flux.Bounds{
							Start: fluxTime(5),
							Stop:  fluxTime(10),
						},
					}),
				},
			},
		},
		{
			name: "from unpushable filter",
			// from -> filter  =>  from -> filter   (no change)
			rules: []plan.Rule{inputs.MergeFromFilterRule{}},
			before: &plantest.PlanSpec{
				Nodes: []plan.PlanNode{
					plan.CreatePhysicalNode("from", from),
					plan.CreatePhysicalNode("filter", &transformations.FilterProcedureSpec{Fn: makeFilterFn(unpushableExpr)}),
				},
				Edges: [][2]int{
					{0, 1},
				},
			},
			after: &plantest.PlanSpec{
				Nodes: []plan.PlanNode{
					plan.CreatePhysicalNode("from", from),
					plan.CreatePhysicalNode("filter", &transformations.FilterProcedureSpec{Fn: makeFilterFn(unpushableExpr)}),
				},
				Edges: [][2]int{
					{0, 1},
				},
			},
		},
		{
			name: "from with statement filter",
			// from -> filter(with statement function)  =>  from -> filter(with statement function)  (no change)
			rules: []plan.Rule{inputs.MergeFromFilterRule{}},
			before: &plantest.PlanSpec{
				Nodes: []plan.PlanNode{
					plan.CreatePhysicalNode("from", from),
					plan.CreatePhysicalNode("filter", &transformations.FilterProcedureSpec{Fn: statementFn}),
				},
				Edges: [][2]int{
					{0, 1},
				},
			},
			after: &plantest.PlanSpec{
				Nodes: []plan.PlanNode{
					plan.CreatePhysicalNode("from", from),
					plan.CreatePhysicalNode("filter", &transformations.FilterProcedureSpec{Fn: statementFn}),
				},
				Edges: [][2]int{
					{0, 1},
				},
			},

		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			//t.Parallel()

			before := plantest.CreatePlanSpec(tc.before)
			after := plantest.CreatePlanSpec(tc.after)

			physicalPlanner := plan.NewPhysicalPlanner(
				plan.OnlyPhysicalRules(tc.rules...),
			)

			pp, err := physicalPlanner.Plan(before)
			if err != nil {
				t.Fatal(err)
			}

			want := make([]fromTestAttrs, 0)
			after.BottomUpWalk(func(node plan.PlanNode) error {
				want = append(want, fromTestAttrs{
					ID:   node.ID(),
					Spec: node.ProcedureSpec(),
				})
				return nil
			})

			got := make([]fromTestAttrs, 0)
			pp.BottomUpWalk(func(node plan.PlanNode) error {
				got = append(got, fromTestAttrs{
					ID:   node.ID(),
					Spec: node.ProcedureSpec(),
				})
				return nil
			})

			if !cmp.Equal(want, got, semantictest.CmpOptions...) {
				t.Errorf("transformed plan not as expected, -want/+got:\n%v",
					cmp.Diff(want, got, semantictest.CmpOptions...))
			}
		})
	}
}
