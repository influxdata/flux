package plan_test

import (
	"testing"

	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/plan/plantest"
)

func TestClone(t *testing.T) {
	//
	//        sink0  sink1
	//          |      |
	//          2      3
	//        /  \   /   \
	//       /    \ /     \
	//      5      4       8
	//            / \
	//           /   \
	//          6     7

	spec := &plantest.PlanSpec{
		Nodes: []plan.Node{
			plantest.CreatePhysicalMockNode("sink0"),
			plantest.CreatePhysicalMockNode("sink1"),
			plantest.CreatePhysicalMockNode("2"),
			plantest.CreatePhysicalMockNode("3"),
			plantest.CreatePhysicalMockNode("4"),
			plantest.CreatePhysicalMockNode("5"),
			plantest.CreatePhysicalMockNode("6"),
			plantest.CreatePhysicalMockNode("7"),
			plantest.CreatePhysicalMockNode("8"),
		},
		Edges: [][2]int{
			{6, 4},
			{7, 4},
			{4, 2},
			{4, 3},
			{5, 2},
			{8, 3},
			{2, 0},
			{3, 1},
		},
	}

	inputPlan := plantest.CreatePlanSpec(spec)
	clonedPlan, err := plan.CloneSpec(inputPlan)
	if err != nil {
		t.Fatal(err)
	}

	inNodes := specToSlice(inputPlan)
	outNodes := specToSlice(clonedPlan.(*plan.Spec))
	if want, got := len(inNodes), len(outNodes); want != got {
		t.Fatalf("Clone() produced a graph with wrong number of nodes; want: %v, got %v", want, got)
	}

	for i, in := range inNodes {
		if want, got := in+"_copy", outNodes[i]; want != got {
			t.Errorf("node ID does not match expected at position %v; want: %v, got %v", i, want, got)
		}
	}

}

func specToSlice(s *plan.Spec) []plan.NodeID {
	var nids []plan.NodeID
	if err := s.TopDownWalk(func(node plan.Node) error {
		nids = append(nids, node.ID())
		return nil
	}); err != nil {
		panic(err)
	}
	return nids
}
