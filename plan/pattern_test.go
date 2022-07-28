package plan_test

import (
	"testing"

	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/stdlib/influxdata/influxdb"
	"github.com/influxdata/flux/stdlib/universe"
)

func TestAny(t *testing.T) {
	t.Run("AnyMultiSuccessor", func(t *testing.T) {
		pat := plan.AnyMultiSuccessor()

		node := &plan.LogicalNode{
			Spec: &influxdb.FromProcedureSpec{},
		}

		if !pat.Match(node) {
			t.Fail()
		}
	})
	t.Run("AnySingleSuccessor", func(t *testing.T) {
		pat := plan.AnySingleSuccessor()

		node := &plan.LogicalNode{
			Spec: &influxdb.FromProcedureSpec{},
		}

		succ0 := &plan.LogicalNode{
			Spec: &universe.FilterProcedureSpec{},
		}
		succ0.AddPredecessors(node)
		succ1 := &plan.LogicalNode{
			Spec: &universe.FilterProcedureSpec{},
		}
		succ1.AddPredecessors(node)
		node.AddSuccessors(succ0, succ1)

		if pat.Match(node) {
			t.Fail()
		}
	})
}

func addEdge(pred plan.Node, succ plan.Node) {
	pred.AddSuccessors(succ)
	succ.AddPredecessors(pred)
}

func TestUnionKindPattern(t *testing.T) {

	// Matches
	//     <anything> |> filter(...) |> filter(...)
	filterFilterPat := plan.MultiSuccessor(universe.FilterKind, plan.SingleSuccessor(universe.FilterKind, plan.AnySingleSuccessor()))

	// Matches
	//   from(...) |> filter(...)
	filterFromPat := plan.MultiSuccessor(universe.FilterKind, plan.SingleSuccessor(influxdb.FromKind))

	from := &plan.LogicalNode{
		Spec: &influxdb.FromProcedureSpec{},
	}

	filter1 := &plan.LogicalNode{
		Spec: &universe.FilterProcedureSpec{},
	}

	addEdge(from, filter1)

	if filterFilterPat.Match(filter1) {
		t.Fatalf("Unexpected match")
	}

	if !filterFromPat.Match(filter1) {
		t.Fatalf("Expected match")
	}

	filter2 := &plan.LogicalNode{
		Spec: &universe.FilterProcedureSpec{},
	}

	addEdge(filter1, filter2)

	// Now we have
	//     from |> filter1 |> filter2

	if !filterFilterPat.Match(filter2) {
		t.Fatalf("Expected match")
	}

	if filterFromPat.Match(filter2) {
		t.Fatalf("Unexpected match")
	}

	// Add another successor to filter1.  Thus should break the filter-filter pattern

	filter3 := &plan.LogicalNode{
		Spec: &universe.FilterProcedureSpec{},
	}

	addEdge(filter1, filter3)

	// Now our graph looks like
	//     t = from |> filter1
	//     filter2(t)
	//     filter3(t)

	if filterFilterPat.Match(filter3) || filterFilterPat.Match(filter2) {
		t.Fatalf("Unexpected match")
	}

	if !filterFromPat.Match(filter1) {
		t.Fatalf("Expected match")
	}
}
