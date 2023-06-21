package join_test

import (
	"context"
	"testing"
	"time"

	_ "github.com/InfluxCommunity/flux/fluxinit/static"
	"github.com/InfluxCommunity/flux/plan"
	"github.com/InfluxCommunity/flux/plan/plantest"
	"github.com/InfluxCommunity/flux/stdlib/influxdata/influxdb"
	"github.com/InfluxCommunity/flux/stdlib/join"
	"github.com/InfluxCommunity/flux/stdlib/universe"
	"github.com/google/go-cmp/cmp"
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
			join.tables(
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
					plan.CreateLogicalNode("sort2", &universe.SortProcedureSpec{
						Columns: []string{"a"},
					}),
					plan.CreateLogicalNode("from3", &influxdb.FromProcedureSpec{}),
					plan.CreateLogicalNode("filter4", &universe.FilterProcedureSpec{}),
					plan.CreateLogicalNode("sort5", &universe.SortProcedureSpec{
						Columns: []string{"b"},
					}),
					plan.CreateLogicalNode("join.tables6", &join.SortMergeJoinProcedureSpec{}),
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
			physicalPlanner := plan.NewPhysicalPlanner(plan.OnlyPhysicalRules(
				&join.EquiJoinPredicateRule{},
				&join.SortMergeJoinPredicateRule{},
			))
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
			getSortSpec := func(p plan.Node) *universe.SortProcedureSpec {
				if s, ok := p.ProcedureSpec().(*universe.SortProcedureSpec); ok {
					return s
				}
				return nil
			}
			// compare the sort nodes created by the planner rule
			err = plantest.ComparePlans(wantPlan, physicalPlan, func(p, q plan.Node) error {
				ps, qs := getSortSpec(p), getSortSpec(q)
				if (ps == nil) && (qs == nil) {
					return nil
				}
				if (ps == nil) || (qs == nil) {
					t.Fatalf("wanted a node of type %T but got a node of type %T", p.ProcedureSpec(), q.ProcedureSpec())
				}

				if diff := cmp.Diff(ps, qs); diff != "" {
					t.Fatalf("unexpected sort node (-want/+got):\n%v", diff)
				}
				return nil
			})
			if err != nil {
				t.Error(err)
			}
		})
	}
}
