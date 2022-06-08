package plan_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/plan/plantest"
	"github.com/stretchr/testify/require"
)

func TestHeuristicPlanner_Plan(t *testing.T) {

	checkVisitedNodes := func(wantNodes []plan.NodeID, rules ...plan.Rule) func(*testing.T, *plan.Spec) {
		return func(t *testing.T, inputSpec *plan.Spec) {
			if len(rules) == 0 {
				rules = append(rules, &plantest.SimpleRule{})
			}
			thePlanner := plan.NewPhysicalPlanner(plan.OnlyPhysicalRules(rules...))
			spec, err := thePlanner.Plan(context.Background(), inputSpec)
			require.NoError(t, err)
			require.NoError(t, spec.CheckIntegrity())
			if simpleRule, ok := rules[0].(*plantest.SimpleRule); ok {
				require.True(t, cmp.Equal(wantNodes, simpleRule.SeenNodes),
					"Traversal didn't match expected, -want/+got:\n%v", cmp.Diff(wantNodes, simpleRule.SeenNodes))
			}
		}
	}

	checkError := func(rule plan.Rule, wantErr string) func(*testing.T, *plan.Spec) {
		return func(t *testing.T, inputSpec *plan.Spec) {
			thePlanner := plan.NewPhysicalPlanner(plan.OnlyPhysicalRules(rule))
			_, err := thePlanner.Plan(context.Background(), inputSpec)
			require.Error(t, err)
			require.True(t, strings.Contains(err.Error(), wantErr))
		}
	}

	testCases := []struct {
		name       string
		plan       plantest.PlanSpec
		validateFn func(*testing.T, *plan.Spec)
	}{
		{
			name: "simple",
			//        0
			plan: plantest.PlanSpec{
				Nodes: []plan.Node{plantest.CreatePhysicalMockNode("0")},
			},
			validateFn: checkVisitedNodes([]plan.NodeID{"0"}),
		},
		{
			name: "simple rule changed",
			//        0
			plan: plantest.PlanSpec{
				Nodes: []plan.Node{plantest.CreatePhysicalMockNode("0")},
			},
			validateFn: checkVisitedNodes([]plan.NodeID{"0"}, &plantest.SimpleRule{ReturnChanged: true}),
		},
		{
			name: "simple rule nil return",
			//        0
			plan: plantest.PlanSpec{
				Nodes: []plan.Node{plantest.CreatePhysicalMockNode("0")},
			},
			validateFn: checkVisitedNodes([]plan.NodeID{"0"}, &plantest.SimpleRule{ReturnNilNode: true}),
		},
		{
			name: "simple rule nil return changed",
			//        0
			plan: plantest.PlanSpec{
				Nodes: []plan.Node{plantest.CreatePhysicalMockNode("0")},
			},
			validateFn: checkError(
				&plantest.SimpleRule{ReturnNilNode: true, ReturnChanged: true},
				"rule \"simple\" returned a nil plan node even though it seems to have changed the plan",
			),
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
			validateFn: checkVisitedNodes([]plan.NodeID{"1", "0"}),
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
			validateFn: checkVisitedNodes([]plan.NodeID{"1", "0", "3", "2"}),
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
			validateFn: checkVisitedNodes([]plan.NodeID{"4", "1", "0", "3", "2"}),
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
			validateFn: checkVisitedNodes([]plan.NodeID{"7", "6", "5", "4", "1", "0", "3", "2"}),
		},
		{
			name: "diamond with rewrite",
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
			validateFn: func(t *testing.T, inputSpec *plan.Spec) {
				var seenNodes []plan.NodeID
				rule := &plantest.FunctionRule{RewriteFn: func(ctx context.Context, node plan.Node) (plan.Node, bool, error) {
					seenNodes = append(seenNodes, node.ID())
					// Replace the central node with a new one
					if len(node.Predecessors()) == 2 && len(node.Successors()) == 2 && node.ID() != "new" {
						// Create a new plan node that will get linked into the plan
						newNode := plantest.CreatePhysicalMockNode("new")
						plan.ReplaceNode(node, newNode)
						return newNode, true, nil
					}
					return node, false, nil
				}}
				thePlanner := plan.NewPhysicalPlanner(plan.OnlyPhysicalRules(rule))
				spec, err := thePlanner.Plan(context.Background(), inputSpec)
				require.NoError(t, err)
				require.NoError(t, spec.CheckIntegrity())
				wantSeenNodes := []plan.NodeID{
					"7", "6", "5", "4", "1", "0", "3", "2", // first pass
					"7", "6", "5", "new", "1", "0", "3", "2", // second pass
				}
				diff := cmp.Diff(wantSeenNodes, seenNodes)
				require.True(t, diff == "", "found difference between -want/+got nodes:\n%v", diff)
			},
		},
		{
			name: "half diamond physical",
			//            2
			//           / \
			//          |   1
			//           \ /
			//            0
			plan: plantest.PlanSpec{
				Nodes: []plan.Node{
					plantest.CreatePhysicalMockNode("0"),
					plantest.CreatePhysicalMockNode("1"),
					plantest.CreatePhysicalMockNode("2"),
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
					{0, 2},
				},
			},
			validateFn: checkVisitedNodes([]plan.NodeID{"2", "1", "0"}),
		},
		{
			name: "half diamond logical",
			//            2
			//           / \
			//          |   1
			//           \ /
			//            0
			plan: plantest.PlanSpec{
				Nodes: []plan.Node{
					plantest.CreateLogicalMockNode("0"),
					plantest.CreateLogicalMockNode("1"),
					plantest.CreateLogicalMockNode("2"),
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
					{0, 2},
				},
			},
			validateFn: checkVisitedNodes([]plan.NodeID{"2", "1", "0"}),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			planSpec := plantest.CreatePlanSpec(&tc.plan)
			tc.validateFn(t, planSpec)
		})
	}
}
