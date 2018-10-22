package inputs_test

import (
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/semantic"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/functions/inputs"
	"github.com/influxdata/flux/functions/transformations"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/plan/plantest"
	"github.com/influxdata/flux/querytest"
	"github.com/influxdata/flux/semantic/semantictest"
	"github.com/influxdata/platform"
	pquerytest "github.com/influxdata/platform/query/querytest"
)

func TestFrom_NewQuery(t *testing.T) {
	t.Skip()
	tests := []querytest.NewQueryTestCase{
		{
			Name:    "from no args",
			Raw:     `from()`,
			WantErr: true,
		},
		{
			Name:    "from conflicting args",
			Raw:     `from(bucket:"d", bucket:"b")`,
			WantErr: true,
		},
		{
			Name:    "from repeat arg",
			Raw:     `from(bucket:"telegraf", bucket:"oops")`,
			WantErr: true,
		},
		{
			Name:    "from",
			Raw:     `from(bucket:"telegraf", chicken:"what is this?")`,
			WantErr: true,
		},
		{
			Name:    "from bucket invalid ID",
			Raw:     `from(bucketID:"invalid")`,
			WantErr: true,
		},
		{
			Name: "from bucket ID",
			Raw:  `from(bucketID:"aaaaaaaa")`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &inputs.FromOpSpec{
							BucketID: platform.ID{170, 170, 170, 170},
						},
					},
				},
			},
		},
		{
			Name: "from with database",
			Raw:  `from(bucket:"mybucket") |> range(start:-4h, stop:-2h) |> sum()`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &inputs.FromOpSpec{
							Bucket: "mybucket",
						},
					},
					{
						ID: "range1",
						Spec: &transformations.RangeOpSpec{
							Start: flux.Time{
								Relative:   -4 * time.Hour,
								IsRelative: true,
							},
							Stop: flux.Time{
								Relative:   -2 * time.Hour,
								IsRelative: true,
							},
							TimeCol:  "_time",
							StartCol: "_start",
							StopCol:  "_stop",
						},
					},
					{
						ID: "sum2",
						Spec: &transformations.SumOpSpec{
							AggregateConfig: execute.DefaultAggregateConfig,
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "range1"},
					{Parent: "range1", Child: "sum2"},
				},
			},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			querytest.NewQueryTestHelper(t, tc)
		})
	}
}

func TestFromOperation_Marshaling(t *testing.T) {
	t.Skip()
	data := []byte(`{"id":"from","kind":"from","spec":{"bucket":"mybucket"}}`)
	op := &flux.Operation{
		ID: "from",
		Spec: &inputs.FromOpSpec{
			Bucket: "mybucket",
		},
	}
	querytest.OperationMarshalingTestHelper(t, data, op)
}

func TestFromOpSpec_BucketsAccessed(t *testing.T) {
	bucketName := "my_bucket"
	bucketID, _ := platform.IDFromString("deadbeef")
	tests := []pquerytest.BucketAwareQueryTestCase{
		{
			Name:             "From with bucket",
			Raw:              `from(bucket:"my_bucket")`,
			WantReadBuckets:  &[]platform.BucketFilter{{Name: &bucketName}},
			WantWriteBuckets: &[]platform.BucketFilter{},
		},
		{
			Name:             "From with bucketID",
			Raw:              `from(bucketID:"deadbeef")`,
			WantReadBuckets:  &[]platform.BucketFilter{{ID: bucketID}},
			WantWriteBuckets: &[]platform.BucketFilter{},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			pquerytest.BucketAwareQueryTestHelper(t, tc)
		})
	}
}

type fromTestAttrs struct {
	ID   plan.NodeID
	Spec plan.ProcedureSpec
}

func yield(name string) *transformations.YieldProcedureSpec {
	return &transformations.YieldProcedureSpec{Name: name}
}

func fluxTime(t int64) flux.Time {
	return flux.Time{
		Absolute: time.Unix(0, t).UTC(),
	}
}

func TestFrom_PlannerTransformationRules(t *testing.T) {
	var (
		fromWithBounds = &inputs.FromProcedureSpec{
			BoundsSet: true,
			Bounds: flux.Bounds{
				Start: fluxTime(5),
				Stop:  fluxTime(10),
			},
		}
		fromWithIntersectedBounds = &inputs.FromProcedureSpec{
			BoundsSet: true,
			Bounds: flux.Bounds{
				Start: fluxTime(9),
				Stop:  fluxTime(10),
			},
		}
		rangeWithBounds = &transformations.RangeProcedureSpec{
			Bounds: flux.Bounds{
				Start: fluxTime(5),
				Stop:  fluxTime(10),
			},
		}
		rangeWithDifferentBounds = &transformations.RangeProcedureSpec{
			Bounds: flux.Bounds{
				Start: fluxTime(9),
				Stop:  fluxTime(14),
			},
		}
		from  = &inputs.FromProcedureSpec{}
		mean  = &transformations.MeanProcedureSpec{}
		count = &transformations.CountProcedureSpec{}

		filterFn1 = &semantic.FunctionExpression{
			Params: []*semantic.FunctionParam{{Key: &semantic.Identifier{Name: "r"}}},
			Body: &semantic.BinaryExpression{Operator: ast.LessThanOperator,
				Left:  &semantic.MemberExpression{Object: &semantic.IdentifierExpression{Name: "r"}, Property: "_value"},
				Right: &semantic.FloatLiteral{Value: 10}},
		}

		filterFn2 = &semantic.FunctionExpression{
			Params: []*semantic.FunctionParam{{Key: &semantic.Identifier{Name: "r"}}},
			Body: &semantic.BinaryExpression{Operator: ast.GreaterThanOperator,
				Left:  &semantic.MemberExpression{Object: &semantic.IdentifierExpression{Name: "r"}, Property: "_value"},
				Right: &semantic.FloatLiteral{Value: 5}},
		}

		filterFnBoth = &semantic.FunctionExpression{
			Params: []*semantic.FunctionParam{{Key: &semantic.Identifier{Name: "r"}}},
			Body: &semantic.LogicalExpression{Operator: ast.AndOperator,
				Left: &semantic.BinaryExpression{Operator: ast.LessThanOperator,
					Left:  &semantic.MemberExpression{Object: &semantic.IdentifierExpression{Name: "r"}, Property: "_value"},
					Right: &semantic.FloatLiteral{Value: 10}},
				Right: &semantic.BinaryExpression{Operator: ast.GreaterThanOperator,
					Left:  &semantic.MemberExpression{Object: &semantic.IdentifierExpression{Name: "r"}, Property: "_value"},
					Right: &semantic.FloatLiteral{Value: 5}},
			},
		}
		fromWithBoundsAndFilter = &inputs.FromProcedureSpec{
			BoundsSet: true,
			Bounds: flux.Bounds{
				Start: fluxTime(5),
				Stop:  fluxTime(10),
				Now:   now,
			},
			FilterSet: true,
			Filter: filterFn1,
		}
	)

	tests := []struct {
		name   string
		rules  []plan.Rule
		before *plantest.PhysicalPlanSpec
		after  *plantest.PhysicalPlanSpec
	}{
		{
			name: "from range",
			// from -> range  =>  from
			rules: []plan.Rule{&inputs.MergeFromRangeRule{}},
			before: &plantest.PhysicalPlanSpec{
				Nodes: []plan.PlanNode{
					plan.CreatePhysicalNode("from", from),
					plan.CreatePhysicalNode("range", rangeWithBounds),
				},
				Edges: [][2]int{{0, 1}},
			},
			after: &plantest.PhysicalPlanSpec{
				Nodes: []plan.PlanNode{
					plan.CreatePhysicalNode("merged_from_range", fromWithBounds),
				},
			},
		},
		{
			name: "from range with successor node",
			// from -> range -> count  =>  from -> count
			rules: []plan.Rule{&inputs.MergeFromRangeRule{}},
			before: &plantest.PhysicalPlanSpec{
				Nodes: []plan.PlanNode{
					plan.CreatePhysicalNode("from", from),
					plan.CreatePhysicalNode("range", rangeWithBounds),
					plan.CreatePhysicalNode("count", count),
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
				},
			},
			after: &plantest.PhysicalPlanSpec{
				Nodes: []plan.PlanNode{
					plan.CreatePhysicalNode("merged_from_range", fromWithBounds),
					plan.CreatePhysicalNode("count", count),
				},
				Edges: [][2]int{{0, 1}},
			},
		},
		{
			name: "from with multiple ranges",
			// from -> range -> range  =>  from
			rules: []plan.Rule{&inputs.MergeFromRangeRule{}},
			before: &plantest.PhysicalPlanSpec{
				Nodes: []plan.PlanNode{
					plan.CreatePhysicalNode("from", from),
					plan.CreatePhysicalNode("range0", rangeWithBounds),
					plan.CreatePhysicalNode("range1", rangeWithDifferentBounds),
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
				},
			},
			after: &plantest.PhysicalPlanSpec{
				Nodes: []plan.PlanNode{
					plan.CreatePhysicalNode("merged_merged_from_range0_range1", fromWithIntersectedBounds),
				},
			},
		},
		{
			name: "from range with multiple successor node",
			// count      mean
			//     \     /          count     mean
			//      range       =>      \    /
			//        |                  from
			//       from
			rules: []plan.Rule{&inputs.MergeFromRangeRule{}},
			before: &plantest.PhysicalPlanSpec{
				Nodes: []plan.PlanNode{
					plan.CreatePhysicalNode("from", from),
					plan.CreatePhysicalNode("range", rangeWithBounds),
					plan.CreatePhysicalNode("count", count),
					plan.CreatePhysicalNode("yield0", yield("count")),
					plan.CreatePhysicalNode("mean", mean),
					plan.CreatePhysicalNode("yield1", yield("mean")),
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
					{2, 3},
					{1, 4},
					{4, 5},
				},
			},
			after: &plantest.PhysicalPlanSpec{
				Nodes: []plan.PlanNode{
					plan.CreatePhysicalNode("merged_from_range", fromWithBounds),
					plan.CreatePhysicalNode("count", count),
					plan.CreatePhysicalNode("mean", mean),
				},
				Edges: [][2]int{
					{0, 1},
					{0, 2},
				},
			},
		},
		{
			name: "cannot push range into from",
			// range    count                                      range    count
			//     \    /       =>   cannot push range into a   =>     \    /
			//      from           from with multiple sucessors         from
			rules: []plan.Rule{&inputs.MergeFromRangeRule{}},
			before: &plantest.PhysicalPlanSpec{
				Nodes: []plan.PlanNode{
					plan.CreatePhysicalNode("from", from),
					plan.CreatePhysicalNode("range", rangeWithBounds),
					plan.CreatePhysicalNode("yield0", yield("range")),
					plan.CreatePhysicalNode("count", count),
					plan.CreatePhysicalNode("yield1", yield("count")),
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
					{0, 3},
					{3, 4},
				},
			},
			after: &plantest.PhysicalPlanSpec{
				Nodes: []plan.PlanNode{
					plan.CreatePhysicalNode("from", from),
					plan.CreatePhysicalNode("range", rangeWithBounds),
					plan.CreatePhysicalNode("count", count),
				},
				Edges: [][2]int{
					{0, 1},
					{0, 2},
				},
			},
		},
		{
			name: "from filter",
			// from -> filter  =>  from
			rules: []plan.Rule{inputs.FromFilterMergeRule{}},
			before: &plantest.PhysicalPlanSpec{
				Nodes: []plan.PlanNode{
					plan.CreatePhysicalNode("from", from),
					plan.CreatePhysicalNode("filter", &transformations.FilterProcedureSpec{Fn: filterFn1}),
				},
				Edges: [][2]int {
					{0, 1},
				},
			},
			after: &plantest.PhysicalPlanSpec{
				Nodes: []plan.PlanNode{
					plan.CreatePhysicalNode("merged_from_filter", &inputs.FromProcedureSpec{FilterSet: true, Filter: filterFn1}),
				},
			},
		},
		{
			name: "from filter filter",
			// from -> filter -> filter  =>  from    (rule applied twice)
			rules: []plan.Rule{inputs.FromFilterMergeRule{}},
			before: &plantest.PhysicalPlanSpec{
				Nodes: []plan.PlanNode{
					plan.CreatePhysicalNode("from", from),
					plan.CreatePhysicalNode("filter1", &transformations.FilterProcedureSpec{Fn: filterFn1}),
					plan.CreatePhysicalNode("filter2", &transformations.FilterProcedureSpec{Fn: filterFn2}),
				},
				Edges: [][2]int {
					{0, 1},
					{1, 2},
				},
			},
			after: &plantest.PhysicalPlanSpec{
				Nodes: []plan.PlanNode{
					plan.CreatePhysicalNode("merged_merged_from_filter1_filter2",
						&inputs.FromProcedureSpec{FilterSet: true, Filter: filterFnBoth}),
				},
			},
		},
		{
			name: "from range filter",
			// from -> range -> filter  =>  from
			rules: []plan.Rule{inputs.FromFilterMergeRule{}, inputs.FromRangeTransformationRule{}},
			before: &plantest.PhysicalPlanSpec{
				Nodes: []plan.PlanNode{
					plan.CreatePhysicalNode("from", from),
					plan.CreatePhysicalNode("range", rangeWithBounds),
					plan.CreatePhysicalNode("filter", &transformations.FilterProcedureSpec{Fn: filterFn1}),
				},
				Edges: [][2]int {
					{0, 1},
					{1, 2},
				},
			},
			after: &plantest.PhysicalPlanSpec{
				Nodes: []plan.PlanNode{
					plan.CreatePhysicalNode("merged_merged_from_range_filter", fromWithBoundsAndFilter),
				},
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			before := plantest.CreatePhysicalPlanSpec(tc.before)
			after := plantest.CreatePhysicalPlanSpec(tc.after)

			physicalPlanner := plan.NewPhysicalPlanner(
				plan.OnlyPhysicalRules(tc.rules...),
			)

			pp, err := physicalPlanner.Plan(before)
			if err != nil {
				t.Fatal(err)
			}

			want := make([]fromTestAttrs, 0)
			after.BottomUpWalk(func(node plan.PlanNode) error {
				want = append(want, fromTestAttrs{
					ID:   node.ID(),
					Spec: node.ProcedureSpec(),
				})
				return nil
			})

			got := make([]fromTestAttrs, 0)
			pp.BottomUpWalk(func(node plan.PlanNode) error {
				got = append(got, fromTestAttrs{
					ID:   node.ID(),
					Spec: node.ProcedureSpec(),
				})
				return nil
			})

			if !cmp.Equal(want, got, semantictest.CmpOptions...) {
				t.Errorf("transformed plan not as expected, -want/+got:\n%v",
					cmp.Diff(want, got, semantictest.CmpOptions...))
			}
		})
	}
}
