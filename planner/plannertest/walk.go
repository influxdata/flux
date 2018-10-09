package plannertest

import (
	"sort"

	"github.com/influxdata/flux/planner"
)

// Visitor visits a query plan node
type Visitor interface {
	Visit(planner.PlanNode) Visitor
}

// Walk traverses a query plan in a depth-first fashion.
// visitor.Visit() is called on each node exactly once.
func Walk(plan *planner.QueryPlan, visitor Visitor) {
	roots := plan.Roots()

	sort.Slice(roots, func(i, j int) bool {
		return roots[i].ID() < roots[j].ID()
	})

	walk(roots, visitor, map[planner.PlanNode]bool{})
}

func walk(nodes []planner.PlanNode, visitor Visitor, visited map[planner.PlanNode]bool) {
	for _, node := range nodes {

		if visited[node] {
			continue
		}

		visitor = visitor.Visit(node)
		walk(node.Predecessors(), visitor, visited)
	}
}
