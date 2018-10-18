package plantest

import (
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/plan"
)

// LogicalPlanSpec is a set of nodes and edges of a logical query plan
type LogicalPlanSpec struct {
	Nodes []plan.PlanNode

	// Edges is a list of predecessor-to-successor edges.
	// [1, 3] => Nodes[1] is a predecessor of Nodes[3].
	// Predecessor ordering must be encoded in this list.
	Edges [][2]int

	Resources flux.ResourceManagement

	Now time.Time
}

// PhysicalPlanSpec is a LogicalPlanSpec with a set of result nodes
type PhysicalPlanSpec struct {
	Nodes     []plan.PlanNode
	Edges     [][2]int
	Resources flux.ResourceManagement
	Now       time.Time

	// Results maps a name to a result node.
	// "a": 3 => Nodes[3] is a result node.
	Results map[string]int
}

// CreateLogicalPlanSpec creates a logcial plan from a set of nodes and edges
func CreateLogicalPlanSpec(spec *LogicalPlanSpec) *plan.PlanSpec {
	return createPlanSpec(spec.Nodes, spec.Edges, spec.Resources, spec.Now)
}

// CreatePhysicalPlanSpec creates a physical plan from a set of nodes, edges, and results
func CreatePhysicalPlanSpec(spec *PhysicalPlanSpec) *plan.PlanSpec {
	plan := createPlanSpec(spec.Nodes, spec.Edges, spec.Resources, spec.Now)
	for name, i := range spec.Results {
		plan.Results[name] = spec.Nodes[i]
	}
	return plan
}

func createPlanSpec(nodes []plan.PlanNode, edges [][2]int, resources flux.ResourceManagement, now time.Time) *plan.PlanSpec {
	predecessors := make(map[plan.PlanNode][]plan.PlanNode)
	successors := make(map[plan.PlanNode][]plan.PlanNode)

	// Compute predecessors and successors of each node
	for _, edge := range edges {

		parent := nodes[edge[0]]
		child := nodes[edge[1]]

		successors[parent] = append(successors[parent], child)
		predecessors[child] = append(predecessors[child], parent)
	}

	roots := make([]plan.PlanNode, 0)

	// Construct query plan
	for _, node := range nodes {

		if len(successors[node]) == 0 {
			roots = append(roots, node)
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
