package plan_test

import (
	"testing"

	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/stdlib/influxdata/influxdb"
	"github.com/influxdata/flux/stdlib/universe"
)

func TestAny(t *testing.T) {
	pat := plan.Any()

	node := &plan.LogicalNode{
		Spec: &influxdb.FromProcedureSpec{},
	}

	if !pat.Match(node) {
		t.Fail()
	}
}

func addEdge(pred plan.Node, succ plan.Node) {
	pred.AddSuccessors(succ)
	succ.AddPredecessors(pred)
}

func TestPat(t *testing.T) {

	// Matches
	//     <anything> |> filter(...) |> filter(...)
	filterFilterPat := plan.Pat(universe.FilterKind, plan.Pat(universe.FilterKind, plan.Any()))

	// Matches
	//   from(...) |> filter(...)
	filterFromPat := plan.Pat(universe.FilterKind, plan.Pat(influxdb.FromKind))

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
