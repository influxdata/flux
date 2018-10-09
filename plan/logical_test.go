package plan_test

import (
	"github.com/influxdata/flux/functions/inputs"
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/functions/transformations"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/plan/plantest"
)

func TestLogicalPlanner_Plan(t *testing.T) {
	testCases := []struct {
		q  *flux.Spec
		ap *plan.LogicalPlanSpec
	}{
		{
			q: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "0",
						Spec: &inputs.FromOpSpec{
							Bucket: "mybucket",
						},
					},
					{
						ID: "1",
						Spec: &transformations.RangeOpSpec{
							Start: flux.Time{Relative: -1 * time.Hour},
							Stop:  flux.Time{},
						},
					},
					{
						ID:   "2",
						Spec: &transformations.CountOpSpec{},
					},
				},
				Edges: []flux.Edge{
					{Parent: "0", Child: "1"},
					{Parent: "1", Child: "2"},
				},
			},
			ap: &plan.LogicalPlanSpec{
				Procedures: map[plan.ProcedureID]*plan.Procedure{
					plan.ProcedureIDFromOperationID("0"): {
						ID: plan.ProcedureIDFromOperationID("0"),
						Spec: &inputs.FromProcedureSpec{
							Bucket: "mybucket",
						},
						Parents:  nil,
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("1")},
					},
					plan.ProcedureIDFromOperationID("1"): {
						ID: plan.ProcedureIDFromOperationID("1"),
						Spec: &transformations.RangeProcedureSpec{
							Bounds: flux.Bounds{
								Start: flux.Time{Relative: -1 * time.Hour},
							},
							TimeCol: "_time",
						},
						Parents: []plan.ProcedureID{
							plan.ProcedureIDFromOperationID("0"),
						},
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("2")},
					},
					plan.ProcedureIDFromOperationID("2"): {
						ID:   plan.ProcedureIDFromOperationID("2"),
						Spec: &transformations.CountProcedureSpec{},
						Parents: []plan.ProcedureID{
							plan.ProcedureIDFromOperationID("1"),
						},
						Children: nil,
					},
				},
				Order: []plan.ProcedureID{
					plan.ProcedureIDFromOperationID("0"),
					plan.ProcedureIDFromOperationID("1"),
					plan.ProcedureIDFromOperationID("2"),
				},
			},
		},
		{
			q: benchmarkQuery,
			ap: &plan.LogicalPlanSpec{
				Procedures: map[plan.ProcedureID]*plan.Procedure{
					plan.ProcedureIDFromOperationID("select0"): {
						ID: plan.ProcedureIDFromOperationID("select0"),
						Spec: &inputs.FromProcedureSpec{
							Bucket: "mybucket",
						},
						Parents:  nil,
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("range0")},
					},
					plan.ProcedureIDFromOperationID("range0"): {
						ID: plan.ProcedureIDFromOperationID("range0"),
						Spec: &transformations.RangeProcedureSpec{
							Bounds: flux.Bounds{
								Start: flux.Time{Relative: -1 * time.Hour},
							},
							TimeCol: "_time",
						},
						Parents: []plan.ProcedureID{
							plan.ProcedureIDFromOperationID("select0"),
						},
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("count0")},
					},
					plan.ProcedureIDFromOperationID("count0"): {
						ID:   plan.ProcedureIDFromOperationID("count0"),
						Spec: &transformations.CountProcedureSpec{},
						Parents: []plan.ProcedureID{
							plan.ProcedureIDFromOperationID("range0"),
						},
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("join")},
					},
					plan.ProcedureIDFromOperationID("select1"): {
						ID: plan.ProcedureIDFromOperationID("select1"),
						Spec: &inputs.FromProcedureSpec{
							Bucket: "mybucket",
						},
						Parents:  nil,
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("range1")},
					},
					plan.ProcedureIDFromOperationID("range1"): {
						ID: plan.ProcedureIDFromOperationID("range1"),
						Spec: &transformations.RangeProcedureSpec{
							Bounds: flux.Bounds{
								Start: flux.Time{Relative: -1 * time.Hour},
							},
							TimeCol: "_time",
						},
						Parents: []plan.ProcedureID{
							plan.ProcedureIDFromOperationID("select1"),
						},
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("sum1")},
					},
					plan.ProcedureIDFromOperationID("sum1"): {
						ID:   plan.ProcedureIDFromOperationID("sum1"),
						Spec: &transformations.SumProcedureSpec{},
						Parents: []plan.ProcedureID{
							plan.ProcedureIDFromOperationID("range1"),
						},
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("join")},
					},
					plan.ProcedureIDFromOperationID("join"): {
						ID: plan.ProcedureIDFromOperationID("join"),
						Spec: &transformations.MergeJoinProcedureSpec{
							TableNames: map[plan.ProcedureID]string{
								plan.ProcedureIDFromOperationID("sum1"):   "sum",
								plan.ProcedureIDFromOperationID("count0"): "count",
							},
						},
						Parents: []plan.ProcedureID{
							plan.ProcedureIDFromOperationID("count0"),
							plan.ProcedureIDFromOperationID("sum1"),
						},
						Children: nil,
					},
				},
				Order: []plan.ProcedureID{
					plan.ProcedureIDFromOperationID("select1"),
					plan.ProcedureIDFromOperationID("range1"),
					plan.ProcedureIDFromOperationID("sum1"),
					plan.ProcedureIDFromOperationID("select0"),
					plan.ProcedureIDFromOperationID("range0"),
					plan.ProcedureIDFromOperationID("count0"),
					plan.ProcedureIDFromOperationID("join"),
				},
			},
		},
	}
	for i, tc := range testCases {
		tc := tc
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			// Set Now time on query spec
			tc.q.Now = time.Now().UTC()
			tc.ap.Now = tc.q.Now

			planner := plan.NewLogicalPlanner()
			got, err := planner.Plan(tc.q)
			if err != nil {
				t.Fatal(err)
			}
			if !cmp.Equal(got, tc.ap, plantest.CmpOptions...) {
				t.Errorf("unexpected logical plan -want/+got %s", cmp.Diff(tc.ap, got, plantest.CmpOptions...))
			}
		})
	}
}

var benchmarkQuery = &flux.Spec{
	Operations: []*flux.Operation{
		{
			ID: "select0",
			Spec: &inputs.FromOpSpec{
				Bucket: "mybucket",
			},
		},
		{
			ID: "range0",
			Spec: &transformations.RangeOpSpec{
				Start: flux.Time{Relative: -1 * time.Hour},
				Stop:  flux.Time{},
			},
		},
		{
			ID:   "count0",
			Spec: &transformations.CountOpSpec{},
		},
		{
			ID: "select1",
			Spec: &inputs.FromOpSpec{
				Bucket: "mybucket",
			},
		},
		{
			ID: "range1",
			Spec: &transformations.RangeOpSpec{
				Start: flux.Time{Relative: -1 * time.Hour},
				Stop:  flux.Time{},
			},
		},
		{
			ID:   "sum1",
			Spec: &transformations.SumOpSpec{},
		},
		{
			ID: "join",
			Spec: &transformations.JoinOpSpec{
				TableNames: map[flux.OperationID]string{
					"count0": "count",
					"sum1":   "sum",
				},
			},
		},
	},
	Edges: []flux.Edge{
		{Parent: "select0", Child: "range0"},
		{Parent: "range0", Child: "count0"},
		{Parent: "select1", Child: "range1"},
		{Parent: "range1", Child: "sum1"},
		{Parent: "count0", Child: "join"},
		{Parent: "sum1", Child: "join"},
	},
}

var benchLogicalPlan *plan.LogicalPlanSpec

func BenchmarkLogicalPlan(b *testing.B) {
	var err error
	planner := plan.NewLogicalPlanner()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		benchLogicalPlan, err = planner.Plan(benchmarkQuery)
		if err != nil {
			b.Fatal(err)
		}
	}
}
