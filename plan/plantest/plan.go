package plantest

import (
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/planner"
)

// LogicalPlanSpec is a set of nodes and edges of a logical query plan
type LogicalPlanSpec struct {
	Nodes []planner.PlanNode

	// Edges is a list of predecessor-to-successor edges.
	// [1, 3] => Nodes[1] is a predecessor of Nodes[3].
	// Predecessor ordering must be encoded in this list.
	Edges [][2]int

	Resources flux.ResourceManagement

	Now time.Time
}

// PhysicalPlanSpec is a LogicalPlanSpec with a set of result nodes
type PhysicalPlanSpec struct {
	Nodes     []planner.PlanNode
	Edges     [][2]int
	Resources flux.ResourceManagement
	Now       time.Time

	// Results maps a name to a result node.
	// "a": 3 => Nodes[3] is a result node.
	Results map[string]int
}

// CreateLogicalPlanSpec creates a logcial plan from a set of nodes and edges
func CreateLogicalPlanSpec(spec *LogicalPlanSpec) *planner.PlanSpec {
	return createPlanSpec(spec.Nodes, spec.Edges, spec.Resources, spec.Now)
}

// CreatePhysicalPlanSpec creates a physical plan from a set of nodes, edges, and results
func CreatePhysicalPlanSpec(spec *PhysicalPlanSpec) *planner.PlanSpec {
	plan := createPlanSpec(spec.Nodes, spec.Edges, spec.Resources, spec.Now)
	for name, i := range spec.Results {
		plan.Results[name] = spec.Nodes[i]
	}
	return plan
}

func createPlanSpec(nodes []planner.PlanNode, edges [][2]int, resources flux.ResourceManagement, now time.Time) *planner.PlanSpec {
	predecessors := make(map[planner.PlanNode][]planner.PlanNode)
	successors := make(map[planner.PlanNode][]planner.PlanNode)

	// Compute predecessors and successors of each node
	for _, edge := range edges {

		parent := nodes[edge[0]]
		child := nodes[edge[1]]

		successors[parent] = append(successors[parent], child)
		predecessors[child] = append(predecessors[child], parent)
	}

	roots := make([]planner.PlanNode, 0)

	// Construct query plan
	for _, node := range nodes {

		if len(successors[node]) == 0 {
			roots = append(roots, node)
		}

		node.AddPredecessors(predecessors[node]...)
		node.AddSuccessors(successors[node]...)
	}

	plan := planner.NewPlanSpec()

	for _, root := range roots {
		plan.Roots[root] = struct{}{}
	}

	plan.Resources = resources
	plan.Now = now
	return plan
}
