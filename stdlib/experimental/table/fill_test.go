package table_test

import (
	"testing"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/plan/plantest"
	"github.com/influxdata/flux/stdlib/experimental/table"
	"github.com/influxdata/flux/stdlib/influxdata/influxdb"
	"github.com/influxdata/flux/stdlib/universe"
)

func TestIdempotentTableFill(t *testing.T) {
	windowSpec := universe.WindowProcedureSpec{
		Window: plan.WindowSpec{
			Every:  flux.ConvertDuration(time.Minute),
			Period: flux.ConvertDuration(time.Minute),
		},
		TimeColumn:  execute.DefaultTimeColLabel,
		StartColumn: execute.DefaultStartColLabel,
		StopColumn:  execute.DefaultStopColLabel,
	}

	tc := plantest.RuleTestCase{
		Rules: []plan.Rule{
			table.IdempotentTableFill{},
		},
		Before: &plantest.PlanSpec{
			Nodes: []plan.Node{
				plan.CreateLogicalNode("from", &influxdb.FromProcedureSpec{}),
				plan.CreateLogicalNode("fill0", &table.FillProcedureSpec{}),
				plan.CreateLogicalNode("fill1", &table.FillProcedureSpec{}),
				plan.CreateLogicalNode("window", &windowSpec),
			},
			Edges: [][2]int{
				{0, 1},
				{1, 2},
				{2, 3},
			},
		},
		After: &plantest.PlanSpec{
			Nodes: []plan.Node{
				plan.CreateLogicalNode("from", &influxdb.FromProcedureSpec{}),
				plan.CreateLogicalNode("fill0", &table.FillProcedureSpec{}),
				plan.CreateLogicalNode("window", &windowSpec),
			},
			Edges: [][2]int{
				{0, 1},
				{1, 2},
			},
		},
	}
	plantest.LogicalRuleTestHelper(t, &tc)
}
