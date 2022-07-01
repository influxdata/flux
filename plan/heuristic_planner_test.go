package plan_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/plan/plantest"
)

func TestPlanTraversal(t *testing.T) {

	testCases := []struct {
		name    string
		plan    plantest.PlanSpec
		nodeIDs []plan.NodeID
	}{
		{
			name: "simple",
			//        0
			plan: plantest.PlanSpec{
				Nodes: []plan.Node{plantest.CreatePhysicalMockNode("0")},
			},
			nodeIDs: []plan.NodeID{"0"},
		},
		{
			name: "two nodes",
			//        1
			//        |
			//        0
			plan: plantest.PlanSpec{
				Nodes: []plan.Node{
					plantest.CreatePhysicalMockNode("0"),
					plantest.CreatePhysicalMockNode("1"),
				},
				Edges: [][2]int{
					{0, 1},
				},
			},
			nodeIDs: []plan.NodeID{"1", "0"},
		},
		{
			name: "multi-root",
			//        1    3
			//        |    |
			//        0    2
			plan: plantest.PlanSpec{
				Nodes: []plan.Node{
					plantest.CreatePhysicalMockNode("0"),
					plantest.CreatePhysicalMockNode("1"),
					plantest.CreatePhysicalMockNode("2"),
					plantest.CreatePhysicalMockNode("3"),
				},
				Edges: [][2]int{
					{0, 1},
					{2, 3},
				},
			},
			nodeIDs: []plan.NodeID{"1", "0", "3", "2"},
		},
		{
			name: "join",
			//        4
			//       / \
			//      1   3
			//      |   |
			//      0   2
			plan: plantest.PlanSpec{
				Nodes: []plan.Node{
					plantest.CreatePhysicalMockNode("0"),
					plantest.CreatePhysicalMockNode("1"),
					plantest.CreatePhysicalMockNode("2"),
					plantest.CreatePhysicalMockNode("3"),
					plantest.CreatePhysicalMockNode("4"),
				},
				Edges: [][2]int{
					{0, 1},
					{2, 3},
					{1, 4},
					{3, 4},
				},
			},
			nodeIDs: []plan.NodeID{"4", "1", "0", "3", "2"},
		},
		{
			name: "diamond",
			//            7
			//           / \
			//          6   5
			//           \ /
			//            4
			//           / \
			//          1   3
			//          |   |
			//          0   2
			plan: plantest.PlanSpec{
				Nodes: []plan.Node{
					plantest.CreatePhysicalMockNode("0"),
					plantest.CreatePhysicalMockNode("1"),
					plantest.CreatePhysicalMockNode("2"),
					plantest.CreatePhysicalMockNode("3"),
					plantest.CreatePhysicalMockNode("4"),
					plantest.CreatePhysicalMockNode("5"),
					plantest.CreatePhysicalMockNode("6"),
					plantest.CreatePhysicalMockNode("7"),
				},
				Edges: [][2]int{
					{0, 1},
					{2, 3},
					{1, 4},
					{3, 4},
					{4, 6},
					{4, 5},
					{6, 7},
					{5, 7},
				},
			},
			nodeIDs: []plan.NodeID{"7", "6", "4", "1", "0", "3", "2", "5"},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			planSpec := plantest.CreatePlanSpec(&tc.plan)

			simpleRule := plantest.SimpleRule{}
			thePlanner := plan.NewPhysicalPlanner(plan.OnlyPhysicalRules(&simpleRule))
			_, err := thePlanner.Plan(context.Background(), planSpec)
			if err != nil {
				t.Fatalf("Could not plan: %v", err)
			}

			if !cmp.Equal(tc.nodeIDs, simpleRule.SeenNodes) {
				t.Errorf("Traversal didn't match expected, -want/+got:\n%v",
					cmp.Diff(tc.nodeIDs, simpleRule.SeenNodes))
			}
		})
	}
}
