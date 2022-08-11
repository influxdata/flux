package plan_test

import (
	"fmt"
	"testing"

	"github.com/andreyvit/diff"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/plan/plantest"
	"github.com/influxdata/flux/plan/plantest/spec"
	"github.com/influxdata/flux/stdlib/influxdata/influxdb"
	"github.com/influxdata/flux/stdlib/universe"
)

func TestFormatted(t *testing.T) {
	fromSpec := &influxdb.FromProcedureSpec{
		Bucket: influxdb.NameOrID{Name: "my-bucket"},
	}

	// (r) => r._value > 5.0
	filterSpec := &universe.FilterProcedureSpec{
		Fn: interpreter.ResolvedFunction{
			Fn: executetest.FunctionExpression(t, `(r) => r._value > 5.0`),
		},
	}

	type testcase struct {
		name string
		plan *plantest.PlanSpec
		want string
	}

	tcs := []testcase{
		{
			name: "from |> filter",
			plan: &plantest.PlanSpec{
				Nodes: []plan.Node{
					plan.CreateLogicalNode("from", fromSpec),
					plan.CreateLogicalNode("filter", filterSpec),
				},
				Edges: [][2]int{
					{0, 1},
				},
			},
			want: `digraph {
  "from"
  "filter"
  // r._value > 5.000000

  "from" -> "filter"
}
`,
		},
		{
			// Check that we see the formatter combine spec
			// details with attribute details.
			name: "parallel merge attribute",
			plan: &plantest.PlanSpec{
				Nodes: []plan.Node{
					plantest.CreatePhysicalNode("source", spec.MockProcedureSpec{
						OutputAttributesFn: func() plan.PhysicalAttributes {
							return plan.PhysicalAttributes{plan.ParallelRunKey: plan.ParallelRunAttribute{Factor: 8}}
						},
					}),
					plantest.CreatePhysicalNode("merge", spec.MockProcedureSpec{
						RequiredAttributesFn: func() []plan.PhysicalAttributes {
							return []plan.PhysicalAttributes{
								{plan.ParallelRunKey: plan.ParallelRunAttribute{Factor: 8}},
							}
						},
						OutputAttributesFn: func() plan.PhysicalAttributes {
							return plan.PhysicalAttributes{plan.ParallelMergeKey: plan.ParallelMergeAttribute{Factor: 8}}
						},
						PlanDetailsFn: func() string {
							return "*** spec details ***"
						},
					}),
				},
				Edges: [][2]int{
					{0, 1},
				},
			},
			want: `digraph {
  "source"
  "merge"
  // *** spec details ***
  // ParallelMergeFactor: 8

  "source" -> "merge"
}
`,
		},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ps := plantest.CreatePlanSpec(tc.plan)
			got := fmt.Sprintf("%v", plan.Formatted(ps, plan.WithDetails()))
			if tc.want != got {
				t.Fatalf("unexpected output: -want/+got:\n%v", diff.LineDiff(tc.want, got))
			}
		})
	}
}
