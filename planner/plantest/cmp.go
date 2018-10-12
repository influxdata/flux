package plantest

import (
	"fmt"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/planner"
	"github.com/influxdata/flux/semantic/semantictest"
)

// ComparePlans compares two query plans using an arbitrary comparator function f
func ComparePlans(p, q *planner.PlanSpec, f func(p, q planner.PlanNode) error) error {
	var w, v []planner.PlanNode

	p.TopDownWalk(func(node planner.PlanNode) error {
		w = append(w, node)
		return nil
	})

	q.TopDownWalk(func(node planner.PlanNode) error {
		v = append(v, node)
		return nil
	})

	if len(w) != len(v) {
		return fmt.Errorf("plans have %d and %d nodes respectively", len(w), len(v))
	}

	for i := 0; i < len(w); i++ {

		if err := f(w[i], v[i]); err != nil {
			return err
		}
	}

	return nil
}

// CompareLogicalPlanNodes is a comparator fuction for LogicalPlanNodes
func CompareLogicalPlanNodes(p, q planner.PlanNode) error {
	if _, ok := p.(*planner.LogicalPlanNode); !ok {
		return fmt.Errorf("expected %s to be a LogicalPlanNode", p.ID())
	}

	if _, ok := q.(*planner.LogicalPlanNode); !ok {
		return fmt.Errorf("expected %s to be a LogicalPlanNode", q.ID())
	}

	return cmpPlanNode(p, q)
}

// ComparePhysicalPlanNodes is a comparator function for PhysicalPlanNodes
func ComparePhysicalPlanNodes(p, q planner.PlanNode) error {
	var pp, qq *planner.PhysicalPlanNode
	var ok bool

	if pp, ok = p.(*planner.PhysicalPlanNode); !ok {
		return fmt.Errorf("expected %s to be a PhysicalPlanNode", p.ID())
	}

	if qq, ok = q.(*planner.PhysicalPlanNode); !ok {
		return fmt.Errorf("expected %s to be a PhysicalPlanNode", q.ID())
	}

	if err := cmpPlanNode(p, q); err != nil {
		return err
	}

	// Both nodes must consume the same required attributes
	if !cmp.Equal(pp.RequiredAttrs, qq.RequiredAttrs) {
		return fmt.Errorf("required attributes not equal -want(%s)/+got(%s) %s",
			pp.ID(), qq.ID(), cmp.Diff(pp.RequiredAttrs, qq.RequiredAttrs))
	}

	// Both nodes must produce the same physical attributes
	if !cmp.Equal(pp.OutputAttrs, qq.OutputAttrs) {
		return fmt.Errorf("output attributes not equal -want(%s)/+got(%s) %s",
			pp.ID(), qq.ID(), cmp.Diff(pp.OutputAttrs, qq.OutputAttrs))
	}

	return nil
}

func cmpPlanNode(p, q planner.PlanNode) error {
	// Both nodes must have the same ID
	if p.ID() != q.ID() {
		return fmt.Errorf("wanted %s, but got %s", p.ID(), q.ID())
	}

	// Both nodes must be the same kind of procedure
	if p.Kind() != q.Kind() {
		return fmt.Errorf("wanted %s, but got %s", p.Kind(), q.Kind())
	}

	// The specifications of both procedures must be the same
	if !cmp.Equal(p.ProcedureSpec(), q.ProcedureSpec(), semantictest.CmpOptions...) {
		return fmt.Errorf("procedure specs not equal -want(%s)/+got(%s) %s",
			p.ID(), q.ID(), cmp.Diff(p.ProcedureSpec(), q.ProcedureSpec()))
	}

	return nil
}
