package plan_test

import (
	"fmt"
	"testing"

	"github.com/andreyvit/diff"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/plan/plantest"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/stdlib/influxdata/influxdb"
	"github.com/influxdata/flux/stdlib/universe"
	"github.com/influxdata/flux/values/valuestest"
)

func TestFormatted(t *testing.T) {
	fromSpec := &influxdb.FromProcedureSpec{
		Bucket: "my-bucket",
	}

	// (r) => r._value > 5.0
	filterSpec := &universe.FilterProcedureSpec{
		Fn: interpreter.ResolvedFunction{
			Fn: &semantic.FunctionExpression{
				Block: &semantic.FunctionBlock{
					Parameters: &semantic.FunctionParameters{
						List: []*semantic.FunctionParameter{{Key: &semantic.Identifier{Name: "r"}}},
					},
					Body: &semantic.BinaryExpression{
						Operator: ast.GreaterThanOperator,
						Left: &semantic.MemberExpression{
							Object:   &semantic.IdentifierExpression{Name: "r"},
							Property: "_value",
						},
						Right: &semantic.FloatLiteral{Value: 5},
					},
				},
			},
			Scope: valuestest.NowScope(),
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
  from
  filter
  // r._value > 5.000000

  from -> filter
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
