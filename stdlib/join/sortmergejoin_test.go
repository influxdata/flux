package join_test

import (
	"context"
	"testing"
	"time"

	_ "github.com/influxdata/flux/fluxinit/static"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/plan/plantest"
	"github.com/influxdata/flux/stdlib/influxdata/influxdb"
	"github.com/influxdata/flux/stdlib/join"
	"github.com/influxdata/flux/stdlib/universe"
)

func TestSortMergeJoinPredicateRule(t *testing.T) {
	now := time.Now().UTC()
	testCases := []struct {
		name     string
		flux     string
		wantErr  error
		wantPlan *plantest.PlanSpec
	}{
		{
			name: "single comparison",
			flux: `import "join"
			left = from(bucket: "b1", host: "http://localhost:8086")
				|> filter(fn: (r) => r._measurement == "a")
			right = from(bucket: "b2", host: "http://localhost:8086")
				|> filter(fn: (r) => r._measurement == "b")
			join.join(
				left: left,
				right: right,
				on: (l, r) => l.a == r.b,
				as: (l, r) => ({l with c: r._value}),
				method: "inner",
			)`,
			wantPlan: &plantest.PlanSpec{
				Nodes: []plan.Node{
					plan.CreateLogicalNode("from0", &influxdb.FromProcedureSpec{}),
					plan.CreateLogicalNode("filter1", &universe.FilterProcedureSpec{}),
					plan.CreateLogicalNode("sort2", &universe.SortProcedureSpec{}),
					plan.CreateLogicalNode("from3", &influxdb.FromProcedureSpec{}),
					plan.CreateLogicalNode("filter4", &universe.FilterProcedureSpec{}),
					plan.CreateLogicalNode("sort5", &universe.SortProcedureSpec{}),
					plan.CreateLogicalNode("join.join6", &join.JoinProcedureSpec{}),
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
					{2, 6},
					{3, 4},
					{4, 5},
					{5, 6},
				},
				Now: now,
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fluxSpec, err := compile(tc.flux, now)
			if err != nil {
				t.Fatalf("could not compile flux query: %v", err)
			}

			logicalPlanner := plan.NewLogicalPlanner()
			initPlan, err := logicalPlanner.CreateInitialPlan(fluxSpec)
			if err != nil {
				t.Fatal(err)
			}
			logicalPlan, err := logicalPlanner.Plan(context.Background(), initPlan)
			if err != nil {
				t.Fatal(err)
			}
			physicalPlanner := plan.NewPhysicalPlanner(plan.OnlyPhysicalRules(&join.SortMergeJoinPredicateRule{}))
			physicalPlan, err := physicalPlanner.Plan(context.Background(), logicalPlan)
			if err != nil {
				if tc.wantErr != nil {
					if tc.wantErr.Error() != err.Error() {
						t.Fatalf("expected error: %s - got %s", tc.wantErr, err)
					}
					return
				} else {
					t.Fatalf("got unexpected error: %s", err)
				}
			} else {
				if tc.wantErr != nil {
					t.Fatalf("expected error `%s` - got none", tc.wantErr)
				}
			}

			wantPlan := plantest.CreatePlanSpec(tc.wantPlan)
			if err := plantest.ComparePlansShallow(wantPlan, physicalPlan); err != nil {
				t.Error(err)
			}
		})
	}
}
