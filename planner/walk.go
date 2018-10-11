package planner

import "sort"

// Walk traverses a query plan in a depth-first fashion. It calls f on each
// operation exactly once. The function f will be called on a node only after
// all of its successors have already been passed to f.
func Walk(plan *PlanSpec, f func(PlanNode) error) error {
	roots := plan.Roots()

	sort.Slice(roots, func(i, j int) bool {
		return roots[i].ID() < roots[j].ID()
	})

	return walk(roots, f, map[PlanNode]bool{})
}

func walk(nodes []PlanNode, f func(PlanNode) error, visited map[PlanNode]bool) error {
	for _, node := range nodes {

		if visited[node] {
			continue
		}

		if err := f(node); err != nil {
			return err
		}

		walk(node.Predecessors(), f, visited)
	}
	return nil
}
