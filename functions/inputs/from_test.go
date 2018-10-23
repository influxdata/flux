package inputs_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/functions/inputs"
	"github.com/influxdata/flux/functions/transformations"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/plan/plantest"
	"github.com/influxdata/flux/semantic/semantictest"
)

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
	)

	tests := []struct {
		name   string
		rules  []plan.Rule
		before *plantest.PlanSpec
		after  *plantest.PlanSpec
	}{
		{
			name: "from range",
			// from -> range  =>  from
			rules: []plan.Rule{&inputs.MergeFromRangeRule{}},
			before: &plantest.PlanSpec{
				Nodes: []plan.PlanNode{
					plan.CreatePhysicalNode("from", from),
					plan.CreatePhysicalNode("range", rangeWithBounds),
				},
				Edges: [][2]int{{0, 1}},
			},
			after: &plantest.PlanSpec{
				Nodes: []plan.PlanNode{
					plan.CreatePhysicalNode("merged_from_range", fromWithBounds),
				},
			},
		},
		{
			name: "from range with successor node",
			// from -> range -> count  =>  from -> count
			rules: []plan.Rule{&inputs.MergeFromRangeRule{}},
			before: &plantest.PlanSpec{
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
			after: &plantest.PlanSpec{
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
			before: &plantest.PlanSpec{
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
			after: &plantest.PlanSpec{
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
			before: &plantest.PlanSpec{
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
			after: &plantest.PlanSpec{
				Nodes: []plan.PlanNode{
					plan.CreatePhysicalNode("merged_from_range", fromWithBounds),
					plan.CreatePhysicalNode("count", count),
					plan.CreatePhysicalNode("yield0", yield("count")),
					plan.CreatePhysicalNode("mean", mean),
					plan.CreatePhysicalNode("yield1", yield("mean")),
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
					{0, 3},
					{3, 4},
				},
			},
		},
		{
			name: "cannot push range into from",
			// range    count                                      range    count
			//     \    /       =>   cannot push range into a   =>     \    /
			//      from           from with multiple sucessors         from
			rules: []plan.Rule{&inputs.MergeFromRangeRule{}},
			before: &plantest.PlanSpec{
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
			after: &plantest.PlanSpec{
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
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			before := plantest.CreatePlanSpec(tc.before)
			after := plantest.CreatePlanSpec(tc.after)

			physicalPlanner := plan.NewPhysicalPlanner(
				plan.WithPhysicalRule(tc.rules...),
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
