package planner_test

import (
	"github.com/influxdata/flux/functions/inputs"
	"github.com/influxdata/flux/planner"
	"github.com/influxdata/flux/planner/plantest"
	"math"
	"testing"
)

func TestPhysicalOptions(t *testing.T) {
	configs := [][]planner.PhysicalOption{
		{planner.WithDefaultMemoryLimit(16384)},
		{},
	}

	for _, options := range configs {
		dag := plantest.DAG{
			Nodes: []planner.PlanNode{
				planner.CreateLogicalNode("from0", &inputs.FromProcedureSpec{}),
			},
			Edges: [][2]int{},
		}

		inputPlan := plantest.CreatePlanFromDAG(dag)

		thePlanner := planner.NewPhysicalPlanner(options...)
		outputPlan, err := thePlanner.Plan(inputPlan)
		if err != nil {
			t.Fatalf("Physical planning failed: %v", err)
		}

		// If option was specified, we should have overridden the default memory quota.
		if len(options) > 0 {
			if outputPlan.Resources.MemoryBytesQuota != 16384 {
				t.Errorf("Expected memory quota of 16384 with option specified")
			}
		} else {
			if outputPlan.Resources.MemoryBytesQuota != math.MaxInt64 {
				t.Errorf("Expected memory quota of math.MaxInt64 with no options specified")
			}
		}
	}
}
