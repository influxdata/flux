package plantest

import (
	"github.com/influxdata/flux/planner"
)

// DAG is defined by a set of nodes and edges
type DAG struct {
	Nodes []planner.PlanNode

	// Edges is a list of predecessor-to-successor edges.
	// {1, 3} <=> Nodes[1] is a predecessor of Nodes[3].
	// Predecessor ordering must be encoded in this list.
	Edges [][2]int
}

// CreatePlanFromDAG constructs a query plan DAG from a set of nodes and edges
func CreatePlanFromDAG(graph DAG) *planner.PlanSpec {
	predecessors := make(map[planner.PlanNode][]planner.PlanNode)
	successors := make(map[planner.PlanNode][]planner.PlanNode)

	// Compute predecessors and successors of each node
	for _, edge := range graph.Edges {

		parent := graph.Nodes[edge[0]]
		child := graph.Nodes[edge[1]]

		successors[parent] = append(successors[parent], child)
		predecessors[child] = append(predecessors[child], parent)
	}

	roots := []planner.PlanNode{}

	// Construct query plan
	for _, node := range graph.Nodes {

		if len(successors[node]) == 0 {
			roots = append(roots, node)
		}

		node.AddPredecessors(predecessors[node]...)
		node.AddSuccessors(successors[node]...)
	}

	return planner.NewPlanSpec(roots)
}
