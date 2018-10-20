package plan_test

import (
	"context"
	"testing"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/functions/inputs"
	"github.com/influxdata/flux/functions/transformations"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/plan/plantest"
	"github.com/influxdata/flux/semantic"
)

func compile(fluxText string, now time.Time) (*flux.Spec, error) {
	return flux.Compile(context.Background(), fluxText, now)
}

// Test the translation of Flux query to logical plan
func TestFluxSpecToLogicalPlan(t *testing.T) {
	now := time.Now().UTC()
	testcases := []struct {
		// Name of the test
		name string

		// Flux query string to translate
		query string

		// Expected logical query plan
		spec *plantest.LogicalPlanSpec
	}{
		{
			name:  `from() |> range()`,
			query: `from(bucket: "my-bucket") |> range(start: -1h)`,
			spec: &plantest.LogicalPlanSpec{
				Nodes: []plan.PlanNode{
					plan.CreateLogicalNode("from0", &inputs.FromProcedureSpec{
						Bucket: "my-bucket",
					}),
					plan.CreateLogicalNode("range1", &transformations.RangeProcedureSpec{
						Bounds: flux.Bounds{
							Start: flux.Time{
								IsRelative: true,
								Relative:   -1 * time.Hour,
							},
							Stop: flux.Time{
								IsRelative: true,
							},
							Now: now,
						},
						TimeCol:  "_time",
						StartCol: "_start",
						StopCol:  "_stop",
					}),
				},
				Edges: [][2]int{
					{0, 1},
				},
			},
		},
		{
			name:  `from() |> range() |> filter()`,
			query: `from(bucket: "my-bucket") |> range(start: -1h) |> filter(fn: (r) => true)`,
			spec: &plantest.LogicalPlanSpec{
				Nodes: []plan.PlanNode{
					plan.CreateLogicalNode("from0", &inputs.FromProcedureSpec{
						Bucket: "my-bucket",
					}),
					plan.CreateLogicalNode("range1", &transformations.RangeProcedureSpec{
						Bounds: flux.Bounds{
							Start: flux.Time{
								IsRelative: true,
								Relative:   -1 * time.Hour,
							},
							Stop: flux.Time{
								IsRelative: true,
							},
							Now: now,
						},
						TimeCol:  "_time",
						StartCol: "_start",
						StopCol:  "_stop",
					}),
					plan.CreateLogicalNode("filter2", &transformations.FilterProcedureSpec{
						Fn: &semantic.FunctionExpression{
							Params: []*semantic.FunctionParam{
								{
									Key: &semantic.Identifier{Name: "r"},
								},
							},
							Body: &semantic.BooleanLiteral{Value: true},
						},
					}),
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
				},
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			spec, err := compile(tc.query, now)

			if err != nil {
				t.Fatal(err)
			}

			want := plantest.CreateLogicalPlanSpec(tc.spec)

			thePlanner := plan.NewLogicalPlanner()
			got, err := thePlanner.Plan(spec)

			if err != nil {
				t.Fatal(err)
			}

			// Comparator function for LogicalPlanNodes
			f := plantest.CompareLogicalPlanNodes

			if err := plantest.ComparePlans(want, got, f); err != nil {
				t.Fatal(err)
			}
		})
	}
}

type MergeFiltersRule struct {
}

func (MergeFiltersRule) Name() string {
	return "mergeFilters"
}

func (MergeFiltersRule) Pattern() plan.Pattern {
	return plan.Pat(transformations.FilterKind,
		plan.Pat(transformations.FilterKind,
			plan.Any()))
}

func (MergeFiltersRule) Rewrite(pn plan.PlanNode) (plan.PlanNode, bool, error) {
	specTop := pn.ProcedureSpec()

	filterSpecTop := specTop.(*transformations.FilterProcedureSpec)
	filterSpecBottom := pn.Predecessors()[0].ProcedureSpec().(*transformations.FilterProcedureSpec)
	mergedFilterSpec := mergeFilterSpecs(filterSpecTop, filterSpecBottom)

	newNode, err := plan.MergeLogicalPlanNodes(pn, pn.Predecessors()[0], mergedFilterSpec)
	if err != nil {
		return pn, false, err
	}

	return newNode, true, nil
}

func mergeFilterSpecs(a, b *transformations.FilterProcedureSpec) plan.ProcedureSpec {
	fn := a.Fn.Copy().(*semantic.FunctionExpression)

	aExp, aOK := a.Fn.Body.(semantic.Expression)
	bExp, bOK := b.Fn.Body.(semantic.Expression)

	if !aOK || !bOK {
		// Note that this is just a unit test, so "return" statements are not handled.
		panic("function body not expression")
	}

	fn.Body = &semantic.LogicalExpression{
		Operator: ast.AndOperator,
		Left:     aExp,
		Right:    bExp,
	}

	return &transformations.FilterProcedureSpec{
		Fn: fn,
	}
}

type PushFilterThroughMapRule struct {
}

func (PushFilterThroughMapRule) Name() string {
	return "pushFilterThroughMap"
}

func (PushFilterThroughMapRule) Pattern() plan.Pattern {
	return plan.Pat(transformations.FilterKind,
		plan.Pat(transformations.MapKind,
			plan.Any()))
}

func (PushFilterThroughMapRule) Rewrite(pn plan.PlanNode) (plan.PlanNode, bool, error) {
	// It will not always be possible to push a filter through a map... but this is just a unit test.

	swapped, err := plan.SwapPlanNodes(pn, pn.Predecessors()[0])
	if err != nil {
		return nil, false, err
	}

	return swapped, true, nil
}

func TestLogicalPlanner(t *testing.T) {
	testcases := []struct {
		name     string
		flux     string
		wantPlan plantest.LogicalPlanSpec
	}{{
		name: "with merge-able filters",
		flux: `
			from(bucket: "telegraf") |>
				filter(fn: (r) => r._measurement == "cpu") |>
				filter(fn: (r) => r._value > 0.5) |>
				filter(fn: (r) => r._value < 0.9) |>
				yield(name: "result")`,
		wantPlan: plantest.LogicalPlanSpec{
			Nodes: []plan.PlanNode{
				plan.CreateLogicalNode("from0", &inputs.FromProcedureSpec{Bucket: "telegraf"}),
				plan.CreateLogicalNode("merged_filter1_merged_filter2_filter3", &transformations.FilterProcedureSpec{Fn: &semantic.FunctionExpression{
					Params: []*semantic.FunctionParam{{Key: &semantic.Identifier{Name: "r"}}},
					Body: &semantic.LogicalExpression{Operator: ast.AndOperator,
						Left: &semantic.LogicalExpression{Operator: ast.AndOperator,
							Left: &semantic.BinaryExpression{Operator: ast.LessThanOperator,
								Left:  &semantic.MemberExpression{Object: &semantic.IdentifierExpression{Name: "r"}, Property: "_value"},
								Right: &semantic.FloatLiteral{Value: 0.9}},
							Right: &semantic.BinaryExpression{Operator: ast.GreaterThanOperator,
								Left:  &semantic.MemberExpression{Object: &semantic.IdentifierExpression{Name: "r"}, Property: "_value"},
								Right: &semantic.FloatLiteral{Value: 0.5}}},
						Right: &semantic.BinaryExpression{Operator: ast.EqualOperator,
							Left:  &semantic.MemberExpression{Object: &semantic.IdentifierExpression{Name: "r"}, Property: "_measurement"},
							Right: &semantic.StringLiteral{Value: "cpu"}}}},
				}),
				plan.CreateLogicalNode("yield4", &transformations.YieldProcedureSpec{Name: "result"}),
			},
			Edges: [][2]int{
				{0, 1},
				{1, 2},
			},
		},
	},
		{
			name: "with swappable map and filter",
			flux: `from(bucket: "telegraf") |> map(fn: (r) => r._value * 2.0) |> filter(fn: (r) => r._value < 10.0) |> yield(name: "result")`,
			wantPlan: plantest.LogicalPlanSpec{
				Nodes: []plan.PlanNode{
					plan.CreateLogicalNode("from0", &inputs.FromProcedureSpec{Bucket: "telegraf"}),
					plan.CreateLogicalNode("filter2_copy", &transformations.FilterProcedureSpec{Fn: &semantic.FunctionExpression{
						Params: []*semantic.FunctionParam{{Key: &semantic.Identifier{Name: "r"}}},
						Body: &semantic.BinaryExpression{Operator: ast.LessThanOperator,
							Left:  &semantic.MemberExpression{Object: &semantic.IdentifierExpression{Name: "r"}, Property: "_value"},
							Right: &semantic.FloatLiteral{Value: 10}},
					}}),
					plan.CreateLogicalNode("map1", &transformations.MapProcedureSpec{
						Fn: &semantic.FunctionExpression{
							Params: []*semantic.FunctionParam{{Key: &semantic.Identifier{Name: "r"}}},
							Body: &semantic.BinaryExpression{Operator: ast.MultiplicationOperator,
								Left:  &semantic.MemberExpression{Object: &semantic.IdentifierExpression{Name: "r"}, Property: "_value"},
								Right: &semantic.FloatLiteral{Value: 2}}},
						MergeKey: true,
					}),
					plan.CreateLogicalNode("yield3", &transformations.YieldProcedureSpec{Name: "result"}),
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
					{2, 3},
				},
			}},
		{
			name: "rules working together",
			flux: `
				from(bucket: "telegraf") |>
					filter(fn: (r) => r._value != 0) |>
					map(fn: (r) => r._value * 10) |>
					filter(fn: (r) => f._value < 100) |>
					yield(name: "result")`,
			wantPlan: plantest.LogicalPlanSpec{
				Nodes: []plan.PlanNode{
					plan.CreateLogicalNode("from0", &inputs.FromProcedureSpec{Bucket: "telegraf"}),
					plan.CreateLogicalNode("merged_filter1_filter3_copy", &transformations.FilterProcedureSpec{Fn: &semantic.FunctionExpression{
						Params: []*semantic.FunctionParam{{Key: &semantic.Identifier{Name: "r"}}},
						Body: &semantic.LogicalExpression{Operator: ast.AndOperator,
							Left: &semantic.BinaryExpression{Operator: ast.LessThanOperator,
								Left:  &semantic.MemberExpression{Object: &semantic.IdentifierExpression{Name: "f"}, Property: "_value"},
								Right: &semantic.IntegerLiteral{Value: 100}},
							Right: &semantic.BinaryExpression{Operator: ast.NotEqualOperator,
								Left:  &semantic.MemberExpression{Object: &semantic.IdentifierExpression{Name: "r"}, Property: "_value"},
								Right: &semantic.IntegerLiteral{}}},
					}}),
					plan.CreateLogicalNode("map2", &transformations.MapProcedureSpec{Fn: &semantic.FunctionExpression{
						Params: []*semantic.FunctionParam{{Key: &semantic.Identifier{Name: "r"}}},
						Body: &semantic.BinaryExpression{Operator: ast.MultiplicationOperator,
							Left:  &semantic.MemberExpression{Object: &semantic.IdentifierExpression{Name: "r"}, Property: "_value"},
							Right: &semantic.IntegerLiteral{Value: 10}}},
						MergeKey: true,
					}),
					plan.CreateLogicalNode("yield4", &transformations.YieldProcedureSpec{Name: "result"}),
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
					{2, 3},
				},
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fluxSpec, err := compile(tc.flux, time.Now().UTC())
			if err != nil {
				t.Fatalf("could not compile flux query: %v", err)
			}

			logicalPlanner := plan.NewLogicalPlanner(plan.OnlyLogicalRules(MergeFiltersRule{}, PushFilterThroughMapRule{}))
			logicalPlan, err := logicalPlanner.Plan(fluxSpec)

			wantPlan := plantest.CreateLogicalPlanSpec(&tc.wantPlan)
			if err := plantest.ComparePlans(wantPlan, logicalPlan, plantest.CompareLogicalPlanNodes); err != nil {
				t.Error(err)
			}
		})
	}
}
