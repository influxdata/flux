package planner_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/functions/transformations"
	"github.com/influxdata/flux/functions/inputs"
	"github.com/influxdata/flux/planner"
	"github.com/influxdata/flux/planner/plannertest"
	"github.com/influxdata/flux/semantic"
)

var createFns = map[flux.OperationKind]planner.CreateLogicalProcedureSpec{
	// Take a FromOpSpec and translate it to a FromProcedureSpec
	inputs.FromKind: func(op flux.OperationSpec) (planner.LogicalProcedureSpec, error) {
		spec, ok := op.(*inputs.FromOpSpec)

		if !ok {
			return nil, fmt.Errorf("invalid spec type %T", op)
		}

		return &planner.FromProcedureSpec{
			Bucket:   spec.Bucket,
			BucketID: spec.BucketID,
		}, nil
	},
	// Take a RangeOpSpec and convert it to a RangeProcedureSpec
	transformations.RangeKind: func(op flux.OperationSpec) (planner.LogicalProcedureSpec, error) {
		spec, ok := op.(*transformations.RangeOpSpec)

		if !ok {
			return nil, fmt.Errorf("invalid spec type %T", op)
		}

		if spec.TimeCol == "" {
			spec.TimeCol = execute.DefaultTimeColLabel
		}

		return &planner.RangeProcedureSpec{
			Bounds: flux.Bounds{
				Start: spec.Start,
				Stop:  spec.Stop,
			},
			TimeCol:  spec.TimeCol,
			StartCol: spec.StartCol,
			StopCol:  spec.StopCol,
		}, nil
	},
	// Take a FilterOpSpec and translate it to a FilterProcedureSpec
	transformations.FilterKind: func(op flux.OperationSpec) (planner.LogicalProcedureSpec, error) {
		spec, ok := op.(*transformations.FilterOpSpec)

		if !ok {
			return nil, fmt.Errorf("invalid spec type %T", op)
		}

		return &planner.FilterProcedureSpec{
			Fn: spec.Fn.Copy().(*semantic.FunctionExpression),
		}, nil
	},
}

// Test the translation of Flux query to logical plan
func TestFluxSpecToLogicalPlan(t *testing.T) {
	testcases := []struct {
		// Name of the test
		name string

		// Flux query string to translate
		query string

		// Expected query plan
		plan plannertest.DAG
	}{
		{
			name:  `from() |> range()`,
			query: `from(bucket: "my-bucket") |> range(start: -1h)`,
			plan: plannertest.DAG{
				Nodes: []planner.PlanNode{
					planner.CreateLogicalNode("from0", &planner.FromProcedureSpec{
						Bucket: "my-bucket",
					}),
					planner.CreateLogicalNode("range1", &planner.RangeProcedureSpec{
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
			plan: plannertest.DAG{
				Nodes: []planner.PlanNode{
					planner.CreateLogicalNode("from0", &planner.FromProcedureSpec{
						Bucket: "my-bucket",
					}),
					planner.CreateLogicalNode("range1", &planner.RangeProcedureSpec{
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
					planner.CreateLogicalNode("filter2", &planner.FilterProcedureSpec{
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

			want := plannertest.CreatePlanFromDAG(tc.plan)
			got, err := planner.CreateLogicalPlan(spec, createFns)

			if err != nil {
				t.Fatal(err)
			}

			// Comparator function for LogicalPlanNodes
			f := plannertest.CompareLogicalPlanNodes

			if err := plannertest.ComparePlans(want, got, f); err != nil {
				t.Fatal(err)
			}
		})
	}
}
