package lang_test

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/ast"
	_ "github.com/influxdata/flux/builtin"
	"github.com/influxdata/flux/lang"
	"github.com/influxdata/flux/parser"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/plan/plantest"
	"github.com/influxdata/flux/stdlib/csv"
	"github.com/influxdata/flux/stdlib/universe"
)

func TestASTCompiler(t *testing.T) {
	testcases := []struct {
		name   string
		now    time.Time
		file   *ast.File
		script string
		want   plantest.PlanSpec
	}{
		{
			name: "override now time using now option",
			now:  time.Unix(1, 1),
			script: `
import "csv"
option now = () => 2017-10-10T00:01:00Z
csv.from(csv: "foo,bar") |> range(start: 2017-10-10T00:00:00Z)
`,
			want: plantest.PlanSpec{
				Nodes: []plan.Node{
					&plan.PhysicalPlanNode{Spec: &csv.FromCSVProcedureSpec{}},
					&plan.PhysicalPlanNode{Spec: &universe.RangeProcedureSpec{}},
					&plan.PhysicalPlanNode{Spec: &universe.YieldProcedureSpec{}},
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
				},
				Resources: flux.ResourceManagement{ConcurrencyQuota: 1, MemoryBytesQuota: math.MaxInt64},
				Now:       parser.MustParseTime("2017-10-10T00:01:00Z").Value,
			},
		},
		{
			name: "get now time from compiler",
			now:  parser.MustParseTime("2018-10-10T00:00:00Z").Value,
			script: `
import "csv"
csv.from(csv: "foo,bar") |> range(start: 2017-10-10T00:00:00Z)
`,
			want: plantest.PlanSpec{
				Nodes: []plan.Node{
					&plan.PhysicalPlanNode{Spec: &csv.FromCSVProcedureSpec{}},
					&plan.PhysicalPlanNode{Spec: &universe.RangeProcedureSpec{}},
					&plan.PhysicalPlanNode{Spec: &universe.YieldProcedureSpec{}},
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
				},
				Resources: flux.ResourceManagement{ConcurrencyQuota: 1, MemoryBytesQuota: math.MaxInt64},
				Now:       parser.MustParseTime("2018-10-10T00:00:00Z").Value,
			},
		},
		{
			name: "prepend file",
			file: &ast.File{
				Body: []ast.Statement{
					&ast.OptionStatement{
						Assignment: &ast.VariableAssignment{
							ID: &ast.Identifier{Name: "now"},
							Init: &ast.FunctionExpression{
								Body: &ast.DateTimeLiteral{
									Value: parser.MustParseTime("2018-10-10T00:00:00Z").Value,
								},
							},
						},
					},
				},
			},
			script: `
import "csv"
csv.from(csv: "foo,bar") |> range(start: 2017-10-10T00:00:00Z)
`,
			want: plantest.PlanSpec{
				Nodes: []plan.Node{
					&plan.PhysicalPlanNode{Spec: &csv.FromCSVProcedureSpec{}},
					&plan.PhysicalPlanNode{Spec: &universe.RangeProcedureSpec{}},
					&plan.PhysicalPlanNode{Spec: &universe.YieldProcedureSpec{}},
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
				},
				Resources: flux.ResourceManagement{ConcurrencyQuota: 1, MemoryBytesQuota: math.MaxInt64},
				Now:       parser.MustParseTime("2018-10-10T00:00:00Z").Value,
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			astPkg, err := flux.Parse(tc.script)
			if err != nil {
				t.Fatalf("failed to parse script: %v", err)
			}

			c := lang.ASTCompiler{
				AST: astPkg,
				Now: tc.now,
			}

			if tc.file != nil {
				c.PrependFile(tc.file)
			}

			prog, err := c.Compile(context.Background())
			if err != nil {
				t.Fatalf("failed to compile AST: %v", err)
			}

			got := prog.(lang.Program).PlanSpec
			want := plantest.CreatePlanSpec(&tc.want)
			if err := plantest.ComparePlansShallow(want, got); err != nil {
				t.Error(err)
			}
		})
	}
}
