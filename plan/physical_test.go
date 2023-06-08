package plan_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/InfluxCommunity/flux/plan"
	"github.com/InfluxCommunity/flux/plan/plantest"
)

func TestPhysicalIntegrityCheckOption(t *testing.T) {
	node0 := plantest.CreatePhysicalMockNode("0")
	node1 := plantest.CreatePhysicalMockNode("1")
	spec := &plantest.PlanSpec{
		Nodes: []plan.Node{
			node0,
			node1,
		},
		Edges: [][2]int{
			{0, 1},
		},
	}

	inputPlan := plantest.CreatePlanSpec(spec)

	intruder := plantest.CreatePhysicalMockNode("intruder")
	// no integrity check enabled, everything should go smoothly
	planner := plan.NewPhysicalPlanner(
		plan.OnlyPhysicalRules(
			plantest.SmashPlanRule{Intruder: intruder, Node: node1},
			plantest.CreateCycleRule{Node: node1},
		),
		plan.DisableValidation(),
	)
	_, err := planner.Plan(context.Background(), inputPlan)
	if err != nil {
		t.Fatalf("unexpected fail: %v", err)
	}

	// let's smash the plan
	planner = plan.NewPhysicalPlanner(
		plan.OnlyPhysicalRules(plantest.SmashPlanRule{Intruder: intruder, Node: node1}))
	_, err = planner.Plan(context.Background(), inputPlan)
	if err == nil {
		t.Fatal("unexpected pass")
	}

	// let's introduce a cycle
	planner = plan.NewPhysicalPlanner(
		plan.OnlyPhysicalRules(plantest.CreateCycleRule{Node: node1}))
	_, err = planner.Plan(context.Background(), inputPlan)
	if err == nil {
		t.Fatal("unexpected pass")
	}
}

// This spec tracks in each node which parallel-test rules have executed
const MockParallelKind = "mock-parallel"

type MockParallelSpec struct {
	plan.DefaultCost
	physical bool
	parallel bool
}

func (MockParallelSpec) Kind() plan.ProcedureKind {
	return MockParallelKind
}

func (ps MockParallelSpec) Copy() plan.ProcedureSpec {
	return MockParallelSpec{physical: ps.physical, parallel: ps.parallel}
}

// Ensure that all physical rules complete before parallel rules are executed.
func TestPhysicalParallelSequence(t *testing.T) {
	node0 := plan.CreatePhysicalNode(plan.NodeID("0"), MockParallelSpec{})
	node1 := plan.CreatePhysicalNode(plan.NodeID("1"), MockParallelSpec{})
	testSpec := &plantest.PlanSpec{
		Nodes: []plan.Node{
			node0,
			node1,
		},
		Edges: [][2]int{
			{0, 1},
		},
	}

	inputPlan := plantest.CreatePlanSpec(testSpec)

	// no integrity check enabled, everything should go smoothly
	planner := plan.NewPhysicalPlanner(
		plan.OnlyPhysicalRules(
			&plantest.FunctionRule{
				RewriteFn: func(ctx context.Context, node plan.Node) (plan.Node, bool, error) {
					ppn := node.(*plan.PhysicalPlanNode)
					spec := ppn.Spec.(MockParallelSpec)

					// Rewrite once only.
					if spec.physical {
						return node, false, nil
					}

					// Ensure the parallel rules have not executed.
					for _, node := range testSpec.Nodes {
						ppn := node.(*plan.PhysicalPlanNode)
						spec := ppn.Spec.(MockParallelSpec)
						if spec.parallel {
							return nil, false, fmt.Errorf("a parallel rule has already executed")
						}
					}

					ppn.Spec = MockParallelSpec{
						physical: true,
						parallel: spec.parallel,
					}
					return node, true, nil
				},
			},
		),
		plan.AddParallelRules(
			&plantest.FunctionRule{
				RewriteFn: func(ctx context.Context, node plan.Node) (plan.Node, bool, error) {
					ppn := node.(*plan.PhysicalPlanNode)
					spec := ppn.Spec.(MockParallelSpec)

					// Rewrite once only.
					if spec.parallel {
						return node, false, nil
					}

					// Ensure the physical rules have executed.
					for _, node := range testSpec.Nodes {
						ppn := node.(*plan.PhysicalPlanNode)
						spec := ppn.Spec.(MockParallelSpec)
						if !spec.physical {
							return nil, false, fmt.Errorf("not all physical rules have run")
						}
					}

					ppn.Spec = MockParallelSpec{
						physical: spec.physical,
						parallel: true,
					}
					return node, true, nil
				},
			},
		),
	)

	_, err := planner.Plan(context.Background(), inputPlan)
	if err != nil {
		t.Fatalf("unexpected fail: %v", err)
	}

	// Ensure all nodes were visited.
	for _, node := range testSpec.Nodes {
		ppn := node.(*plan.PhysicalPlanNode)
		spec := ppn.Spec.(MockParallelSpec)
		if !spec.parallel || !spec.physical {
			t.Fatalf("all nodes must be visited by both physical and parallel passes")
		}
	}
}
