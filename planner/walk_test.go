package planner_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/planner"
	"github.com/influxdata/flux/planner/plantest"
)

type mockProcedureSpec struct {
}

func (m *mockProcedureSpec) Kind() planner.ProcedureKind {
	return "mock"
}

func (m *mockProcedureSpec) Copy() planner.ProcedureSpec {
	return &mockProcedureSpec{}
}

func makeNode(id string) planner.PlanNode {
	return planner.CreateLogicalNode(planner.NodeID(id), &mockProcedureSpec{})
}

func TestPlanSpec_BottomUpWalk(t *testing.T) {
	dag := plantest.DAG{
		//  0 1 2  additional edge (3->2)
		//  |\|\|
		//  3 4 5  additional edge (8->3)
		//  |/|/|
		//  6 7 8
		Nodes: []planner.PlanNode{
			makeNode("0"),
			makeNode("1"),
			makeNode("2"),

			makeNode("3"),
			makeNode("4"),
			makeNode("5"),

			makeNode("6"),
			makeNode("7"),
			makeNode("8"),
		},
		Edges: [][2]int{
			{6, 3},
			{6, 4},
			{7, 4},
			{7, 5},
			{8, 3},
			{8, 5},

			{3, 0},
			{3, 2},
			{4, 0},
			{4, 1},
			{5, 1},
			{5, 2},
		},
	}

	thePlan := plantest.CreatePlanFromDAG(dag)

	got := make([]planner.NodeID, 0, 9)
	thePlan.BottomUpWalk(func(n planner.PlanNode) error {
		got = append(got, n.ID())
		return nil
	})

	want := []planner.NodeID{"6", "8", "3", "7", "4", "0", "5", "1", "2"}
	if !cmp.Equal(want, got) {
		t.Errorf("Did not get expected node traversal, -want/+got:\n%v", cmp.Diff(want, got))
	}
}

func TestPlanSpec_TopDownWalk(t *testing.T) {
	dag := plantest.DAG{
		//  0 1 2  additional edge (3->2)
		//  |\|\|
		//  3 4 5  additional edge (8->3)
		//  |/|/|
		//  6 7 8
		Nodes: []planner.PlanNode{
			makeNode("0"),
			makeNode("1"),
			makeNode("2"),

			makeNode("3"),
			makeNode("4"),
			makeNode("5"),

			makeNode("6"),
			makeNode("7"),
			makeNode("8"),
		},
		Edges: [][2]int{
			{6, 3},
			{6, 4},
			{7, 4},
			{7, 5},
			{8, 3},
			{8, 5},

			{3, 0},
			{3, 2},
			{4, 0},
			{4, 1},
			{5, 1},
			{5, 2},
		},
	}

	thePlan := plantest.CreatePlanFromDAG(dag)

	got := make([]planner.NodeID, 0, 9)
	thePlan.TopDownWalk(func(n planner.PlanNode) error {
		got = append(got, n.ID())
		return nil
	})

	want := []planner.NodeID{"2", "1", "5", "0", "4", "7", "3", "8", "6"}
	if !cmp.Equal(want, got) {
		t.Errorf("Did not get expected node traversal, -want/+got:\n%v", cmp.Diff(want, got))
	}
}
