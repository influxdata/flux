package plantest

import (
	"fmt"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/plan"
)

// Spec is a set of nodes and edges of a logical query plan
type PlanSpec struct {
	Nodes []plan.Node

	// Edges is a list of predecessor-to-successor edges.
	// [1, 3] => Nodes[1] is a predecessor of Nodes[3].
	// Predecessor ordering must be encoded in this list.
	Edges [][2]int

	Resources flux.ResourceManagement

	Now time.Time
}

// CreatePlanSpec creates a logical plan from a set of nodes and edges
func CreatePlanSpec(spec *PlanSpec) *plan.Spec {
	return createPlanSpec(spec.Nodes, spec.Edges, spec.Resources, spec.Now)
}

// Copy makes a copy of a Spec.
func (ps *PlanSpec) Copy() *PlanSpec {
	cps := new(PlanSpec)

	cps.Nodes = make([]plan.Node, len(ps.Nodes))
	for i := range ps.Nodes {
		cps.Nodes[i] = copyNode(ps.Nodes[i])
	}

	cps.Edges = make([][2]int, len(ps.Edges))
	copy(cps.Edges, ps.Edges)
	cps.Resources = ps.Resources
	cps.Now = ps.Now
	return cps
}

func copyNode(n plan.Node) plan.Node {
	var cn plan.Node
	switch n := n.(type) {
	case *plan.LogicalNode:
		cn = plan.CreateLogicalNode(n.ID(), n.ProcedureSpec().Copy())
	case *plan.PhysicalPlanNode:
		cn = plan.CreatePhysicalNode(n.ID(), n.ProcedureSpec().Copy().(plan.PhysicalProcedureSpec))
	}
	return cn
}

func createPlanSpec(nodes []plan.Node, edges [][2]int, resources flux.ResourceManagement, now time.Time) *plan.Spec {
	predecessors := make(map[plan.Node][]plan.Node)
	successors := make(map[plan.Node][]plan.Node)

	// Compute predecessors and successors of each node
	for _, edge := range edges {

		parent := nodes[edge[0]]
		child := nodes[edge[1]]

		successors[parent] = append(successors[parent], child)
		predecessors[child] = append(predecessors[child], parent)
	}

	roots := make([]plan.Node, 0)

	// Construct query plan
	for _, node := range nodes {

		if len(successors[node]) == 0 {
			roots = append(roots, node)
		}

		if len(nodes) > 1 && len(predecessors[node]) == 0 && len(successors[node]) == 0 {
			panic(fmt.Errorf("found disconnected node: %v", node.ID()))
		}

		node.AddPredecessors(predecessors[node]...)
		node.AddSuccessors(successors[node]...)
	}

	plan := plan.NewPlanSpec()

	for _, root := range roots {
		plan.Roots[root] = struct{}{}
	}

	plan.Resources = resources
	plan.Now = now
	return plan
}
