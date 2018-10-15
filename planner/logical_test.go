package planner_test

import (
	"context"
	"testing"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/functions/inputs"
	"github.com/influxdata/flux/functions/transformations"
	"github.com/influxdata/flux/planner"
	"github.com/influxdata/flux/planner/plantest"
	"github.com/influxdata/flux/semantic"
)

// Test the translation of Flux query to logical plan
func TestFluxSpecToLogicalPlan(t *testing.T) {
	testcases := []struct {
		// Name of the test
		name string

		// Flux query string to translate
		query string

		// Expected query plan
		plan plantest.DAG
	}{
		{
			name:  `from() |> range()`,
			query: `from(bucket: "my-bucket") |> range(start: -1h)`,
			plan: plantest.DAG{
				Nodes: []planner.PlanNode{
					planner.CreateLogicalNode("from0", &inputs.FromProcedureSpec{
						Bucket: "my-bucket",
					}),
					planner.CreateLogicalNode("range1", &transformations.RangeProcedureSpec{
						Bounds: flux.Bounds{
							Start: flux.Time{
								IsRelative: true,
								Relative:   -1 * time.Hour,
							},
							Stop: flux.Time{
								IsRelative: true,
							},
						},
						TimeCol:  "_time",
						StartCol: "_start",
						StopCol:  "_stop",
					}),
				},
				Edges: [][2]int{
					{0, 1},
				},
			},
		},
		{
			name:  `from() |> range() |> filter()`,
			query: `from(bucket: "my-bucket") |> range(start: -1h) |> filter(fn: (r) => true)`,
			plan: plantest.DAG{
				Nodes: []planner.PlanNode{
					planner.CreateLogicalNode("from0", &inputs.FromProcedureSpec{
						Bucket: "my-bucket",
					}),
					planner.CreateLogicalNode("range1", &transformations.RangeProcedureSpec{
						Bounds: flux.Bounds{
							Start: flux.Time{
								IsRelative: true,
								Relative:   -1 * time.Hour,
							},
							Stop: flux.Time{
								IsRelative: true,
							},
						},
						TimeCol:  "_time",
						StartCol: "_start",
						StopCol:  "_stop",
					}),
					planner.CreateLogicalNode("filter2", &transformations.FilterProcedureSpec{
						Fn: &semantic.FunctionExpression{
							Params: []*semantic.FunctionParam{
								{
									Key: &semantic.Identifier{Name: "r"},
								},
							},
							Body: &semantic.BooleanLiteral{Value: true},
						},
					}),
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
				},
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			now := time.Now().UTC()
			spec, err := flux.Compile(context.Background(), tc.query, now)

			if err != nil {
				t.Fatal(err)
			}

			want := plantest.CreatePlanFromDAG(tc.plan, flux.ResourceManagement{}, time.Time{})

			thePlanner := planner.NewLogicalPlanner()
			got, err := thePlanner.Plan(spec)

			if err != nil {
				t.Fatal(err)
			}

			// Comparator function for LogicalPlanNodes
			f := plantest.CompareLogicalPlanNodes

			if err := plantest.ComparePlans(want, got, f); err != nil {
				t.Fatal(err)
			}
		})
	}
}
