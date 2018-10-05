package plannertest

import (
	"fmt"
	"sort"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/planner"
	"github.com/influxdata/flux/semantic/semantictest"
)

// CompareLogicalPlans compares two logical query plans for equality by value
func CompareLogicalPlans(p, q *planner.QueryPlan) error {
	pRoots := p.Roots()
	qRoots := q.Roots()

	// Sort the roots of the two plans
	sort.Slice(pRoots, func(i, j int) bool {
		return pRoots[i].ID() < pRoots[j].ID()
	})
	sort.Slice(qRoots, func(i, j int) bool {
		return qRoots[i].ID() < qRoots[j].ID()
	})
	return cmpLogicalPlans(pRoots, qRoots, map[cmpPair]bool{}, 0)
}

func cmpLogicalPlans(p, q []planner.PlanNode, visited map[cmpPair]bool, level int) error {
	if len(p) != len(q) {
		return fmt.Errorf("plans do not have same number of nodes at level %d", level)
	}

	for i := 0; i < len(p); i++ {

		if visited[cmpPair{p[i], q[i]}] {
			continue
		}

		// Must be logical plan nodes
		if _, ok := p[i].(*planner.LogicalPlanNode); !ok {
			return fmt.Errorf("encountered a non-logical plan node with spec type %T", p[i].ProcedureSpec())
		}

		if _, ok := q[i].(*planner.LogicalPlanNode); !ok {
			return fmt.Errorf("encountered a non-logical plan node with spec type %T", q[i].ProcedureSpec())
		}

		// Must have the same IDs
		if p[i].ID() != q[i].ID() {
			return fmt.Errorf("expected NodeID %s, but instead got %s", p[i].ID(), q[i].ID())
		}

		// Must be the same kind of procedure
		if p[i].Kind() != q[i].Kind() {
			return fmt.Errorf("expected ProcedureKind %s, but instead got %s", p[i].Kind(), q[i].Kind())
		}

		// The specifications of both procedures must be the same
		if !cmp.Equal(p[i].ProcedureSpec(), q[i].ProcedureSpec(), semantictest.CmpOptions...) {
			return fmt.Errorf("logical plan nodes not equal -want/+got %s", cmp.Diff(
				p[i].ProcedureSpec(), q[i].ProcedureSpec()))
		}

		// Nodes are equal
		visited[cmpPair{p[i], q[i]}] = true

		// Recurse on predecessors
		if err := cmpLogicalPlans(p[i].Predecessors(), q[i].Predecessors(), visited, level+1); err != nil {
			return err
		}
	}
	return nil
}

// ComparePhysicalPlans compares two physical query plans for equality by value
func ComparePhysicalPlans(p, q *planner.QueryPlan) error {
	pRoots := p.Roots()
	qRoots := q.Roots()

	// Sort the roots of the two plans
	sort.Slice(pRoots, func(i, j int) bool {
		return pRoots[i].ID() < pRoots[j].ID()
	})
	sort.Slice(qRoots, func(i, j int) bool {
		return qRoots[i].ID() < qRoots[j].ID()
	})
	return cmpPhysicalPlans(pRoots, qRoots, map[cmpPair]bool{}, 0)
}

func cmpPhysicalPlans(p, q []planner.PlanNode, visited map[cmpPair]bool, level int) error {
	if len(p) != len(q) {
		return fmt.Errorf("plans do not have same number of nodes at level %d", level)
	}

	for i := 0; i < len(p); i++ {

		if visited[cmpPair{p[i], q[i]}] {
			continue
		}

		// Must be physical plan nodes
		if _, ok := p[i].(*planner.PhysicalPlanNode); !ok {
			return fmt.Errorf("encountered a non-physical plan node with spec type %T", p[i].ProcedureSpec())
		}

		if _, ok := q[i].(*planner.PhysicalPlanNode); !ok {
			return fmt.Errorf("encountered a non-physical plan node with spec type %T", q[i].ProcedureSpec())
		}

		left := p[i].(*planner.PhysicalPlanNode)
		right := q[i].(*planner.PhysicalPlanNode)

		// Must have the same IDs
		if p[i].ID() != q[i].ID() {
			return fmt.Errorf("expected NodeID %s, but instead got %s", p[i].ID(), q[i].ID())
		}

		// Must be the same kind of procedure
		if p[i].Kind() != q[i].Kind() {
			return fmt.Errorf("expected ProcedureKind %s, but instead got %s", p[i].Kind(), q[i].Kind())
		}

		// The specifications of both procedures must be the same
		if !cmp.Equal(p[i].ProcedureSpec(), q[i].ProcedureSpec(), semantictest.CmpOptions...) {
			return fmt.Errorf("physical plan nodes not equal -want/+got %s", cmp.Diff(
				p[i].ProcedureSpec(), q[i].ProcedureSpec()))
		}

		// Must have the same number of required attributes
		if len(left.RequiredAttrs) != len(right.RequiredAttrs) {
			return fmt.Errorf("procedures %s and %s do not have the same number of required attributes",
				left.ID(), right.ID())
		}

		// Must have the same required attributes
		for i := 0; i < len(left.RequiredAttrs); i++ {
			if !cmp.Equal(left.RequiredAttrs[i], right.RequiredAttrs[i]) {
				return fmt.Errorf("required attribute not equal -want/+got %s", cmp.Diff(
					left.RequiredAttrs[i], right.RequiredAttrs[i]))
			}
		}

		// Must produce the same attributes
		if !cmp.Equal(left.OutputAttrs, right.OutputAttrs) {
			return fmt.Errorf("output attributes not equal -want/+got %s", cmp.Diff(
				left.OutputAttrs, right.OutputAttrs))
		}

		// Nodes are equal
		visited[cmpPair{p[i], q[i]}] = true

		// Recurse on predecessors
		if err := cmpPhysicalPlans(p[i].Predecessors(), q[i].Predecessors(), visited, level+1); err != nil {
			return err
		}
	}
	return nil
}

type cmpPair struct {
	left, right planner.PlanNode
}
