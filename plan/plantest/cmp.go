package plantest

import (
	"fmt"
	"reflect"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic/semantictest"
	"github.com/influxdata/flux/stdlib/kafka"
	"github.com/influxdata/flux/stdlib/universe"
	"github.com/influxdata/flux/values/valuestest"
)

// CmpOptions are the options needed to compare plan.ProcedureSpecs inside plan.Spec.
var CmpOptions = append(
	semantictest.CmpOptions,
	cmp.AllowUnexported(flux.Spec{}),
	cmp.AllowUnexported(universe.JoinOpSpec{}),
	cmpopts.IgnoreUnexported(flux.Spec{}),
	cmpopts.IgnoreUnexported(universe.JoinOpSpec{}),
	cmp.AllowUnexported(kafka.ToKafkaProcedureSpec{}),
	cmpopts.IgnoreUnexported(kafka.ToKafkaProcedureSpec{}),
	valuestest.ScopeComparer,
)

// ComparePlans compares two query plans using an arbitrary comparator function f
func ComparePlans(p, q *plan.Spec, f func(p, q plan.Node) error) error {
	err := compareMetadata(p, q)
	if err != nil {
		return err
	}

	var w, v []plan.Node

	_ = p.TopDownWalk(func(node plan.Node) error {
		w = append(w, node)
		return nil
	})

	_ = q.TopDownWalk(func(node plan.Node) error {
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

// ComparePlansShallow Compares the two specs, but only compares the
// metadata and the types of each node.  Individual fields of procedure specs
// are not compared.
func ComparePlansShallow(p, q *plan.Spec) error {
	if err := ComparePlans(p, q, cmpPlanNodeShallow); err != nil {
		return err
	}
	return nil
}

func compareMetadata(p, q *plan.Spec) error {
	opts := cmpopts.IgnoreFields(plan.Spec{}, "Roots")
	if diff := cmp.Diff(p, q, opts); diff != "" {
		return fmt.Errorf("plan metadata not equal; -want/+got:\n%v", diff)
	}
	return nil
}

// CompareLogicalPlans compares two logical plans.
func CompareLogicalPlans(p, q *plan.Spec) error {
	return ComparePlans(p, q, CompareLogicalPlanNodes)
}

// CompareLogicalPlanNodes is a comparator function for LogicalPlanNodes
func CompareLogicalPlanNodes(p, q plan.Node) error {
	if _, ok := p.(*plan.LogicalNode); !ok {
		return fmt.Errorf("expected %s to be a LogicalNode", p.ID())
	}

	if _, ok := q.(*plan.LogicalNode); !ok {
		return fmt.Errorf("expected %s to be a LogicalNode", q.ID())
	}

	return cmpPlanNode(p, q)
}

// ComparePhysicalPlanNodes is a comparator function for PhysicalPlanNodes
func ComparePhysicalPlanNodes(p, q plan.Node) error {
	var pp, qq *plan.PhysicalPlanNode
	var ok bool

	if pp, ok = p.(*plan.PhysicalPlanNode); !ok {
		return fmt.Errorf("expected %s to be a PhysicalPlanNode", p.ID())
	}

	if qq, ok = q.(*plan.PhysicalPlanNode); !ok {
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

func cmpPlanNode(p, q plan.Node) error {
	// Both nodes must have the same ID
	if p.ID() != q.ID() {
		return fmt.Errorf("wanted %s, but got %s", p.ID(), q.ID())
	}

	// Both nodes must be the same kind of procedure
	if p.Kind() != q.Kind() {
		return fmt.Errorf("wanted %s, but got %s", p.Kind(), q.Kind())
	}

	// Both nodes must have the same time bounds
	if !cmp.Equal(p.Bounds(), q.Bounds()) {
		return fmt.Errorf("plan nodes have different bounds -want(%s)/+got(%s) %s",
			p.ID(), q.ID(), cmp.Diff(p.Bounds(), q.Bounds()))
	}

	// The specifications of both procedures must be the same
	if !cmp.Equal(p.ProcedureSpec(), q.ProcedureSpec(), CmpOptions...) {
		return fmt.Errorf("procedure specs not equal -want(%s)/+got(%s) %s",
			p.ID(), q.ID(), cmp.Diff(p.ProcedureSpec(), q.ProcedureSpec(), CmpOptions...))
	}

	return nil
}

func cmpPlanNodeShallow(p, q plan.Node) error {
	// Just make sure that they have the same type
	pt, qt := reflect.TypeOf(p.ProcedureSpec()), reflect.TypeOf(q.ProcedureSpec())

	_, pIsYield := p.ProcedureSpec().(plan.YieldProcedureSpec)
	_, qIsYield := q.ProcedureSpec().(plan.YieldProcedureSpec)
	if pIsYield && qIsYield {
		// generated yields are produced by the planner but their specs
		// are not public.  So consider any types that implement yield to be equal.
		return nil
	}
	if pIsYield != qIsYield {
		if pIsYield {
			return fmt.Errorf("wanted a yield, but got a %v", qt)
		} else {
			return fmt.Errorf("wanted a %v, but got a yield", pt)
		}
	}

	if pt != qt {
		return fmt.Errorf("plan nodes have different types; -want/+got: -%v/+%v", pt, qt)
	}

	return nil
}
