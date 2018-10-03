package planner_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/plan/plantest"
	"github.com/influxdata/flux/planner"
)

// from() |> range() -> FromRange
func TestPlanFromRange(t *testing.T) {
	// Instantiate planner
	physicalPlanner := planner.NewLogicalToPhysicalPlanner([]planner.Rule{
		planner.FromRangeTransformationRule{},
	})

	rangeProcedureSpec := &planner.RangeProcedureSpec{
		Bounds: flux.Bounds{
			Start: flux.Time{
				IsRelative: true,
				Relative:   -1 * time.Hour,
			},
			Stop: flux.Now,
		},
		TimeCol:  "_time",
		StartCol: "_start",
		StopCol:  "_stop",
	}

	fromProcedureSpec := &planner.FromProcedureSpec{
		Bucket: "bucket",
	}

	logicalRangePlan := &planner.LogicalPlanNode{
		Spec: rangeProcedureSpec,
	}

	logicalFromPlan := &planner.LogicalPlanNode{
		Spec: fromProcedureSpec,
	}

	logicalRangePlan.Edges = planner.Edges{
		Pred: []planner.PlanNode{
			logicalFromPlan,
		},
	}

	logicalFromPlan.Edges = planner.Edges{
		Succ: []planner.PlanNode{
			logicalRangePlan,
		},
	}

	// planner should produce the following plan
	want := &planner.PhysicalPlanNode{
		Spec: &planner.FromRangeProcedureSpec{
			Bucket: "bucket",
			Bounds: flux.Bounds{
				Start: flux.Time{
					IsRelative: true,
					Relative:   -1 * time.Hour,
				},
				Stop: flux.Now,
			},
			TimeCol:  "_time",
			StartCol: "_start",
			StopCol:  "_stop",
		},
	}

	got, err := physicalPlanner.Plan(logicalRangePlan)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(got, want, plantest.CmpOptions...) {
		t.Errorf("unexpected physical plan -want/+got:\n%s", cmp.Diff(want, got, plantest.CmpOptions...))
	}
}
