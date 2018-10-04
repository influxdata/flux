package plan_test

import (
	"github.com/influxdata/flux/functions/inputs"
	"testing"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/functions/transformations"
	"github.com/influxdata/flux/plan"
)

func TestPhysicalPlanner_DefaultMemoryLimit(t *testing.T) {
	// Simple logical plan taken from the planner tests.
	lp := &plan.LogicalPlanSpec{
		Now: time.Now(),
		Resources: flux.ResourceManagement{
			ConcurrencyQuota: 1,
		},
		Procedures: map[plan.ProcedureID]*plan.Procedure{
			plan.ProcedureIDFromOperationID("from"): {
				ID: plan.ProcedureIDFromOperationID("from"),
				Spec: &inputs.FromProcedureSpec{
					Bucket: "mydb",
				},
				Parents:  nil,
				Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("range")},
			},
			plan.ProcedureIDFromOperationID("range"): {
				ID: plan.ProcedureIDFromOperationID("range"),
				Spec: &transformations.RangeProcedureSpec{
					Bounds: flux.Bounds{
						Start: flux.Time{
							IsRelative: true,
							Relative:   -1 * time.Hour,
						},
						Stop: flux.Now,
					},
					TimeCol: "_time",
				},
				Parents: []plan.ProcedureID{
					plan.ProcedureIDFromOperationID("from"),
				},
				Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("count")},
			},
			plan.ProcedureIDFromOperationID("count"): {
				ID:   plan.ProcedureIDFromOperationID("count"),
				Spec: &transformations.CountProcedureSpec{},
				Parents: []plan.ProcedureID{
					(plan.ProcedureIDFromOperationID("range")),
				},
				Children: nil,
			},
		},
		Order: []plan.ProcedureID{
			plan.ProcedureIDFromOperationID("from"),
			plan.ProcedureIDFromOperationID("range"),
			plan.ProcedureIDFromOperationID("count"),
		},
	}

	planner := plan.NewPlanner(plan.WithDefaultMemoryLimit(1024))
	spec, err := planner.Plan(lp, nil)
	if err != nil {
		t.Fatal(err)
	}

	// The plan spec should have 1024 set for the memory limits.
	if got, exp := spec.Resources.MemoryBytesQuota, int64(1024); got != exp {
		t.Fatalf("unexpected memory bytes quota: exp=%d got=%d", exp, got)
	}
}
