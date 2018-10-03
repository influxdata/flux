package planner

import (
	"testing"
)

func TestAny(t *testing.T) {
	pat := Any()

	node := &LogicalPlanNode{
		Spec: &FromProcedureSpec{},
	}

	if ! pat.Match(node) {
		t.Fail()
	}
}

func addEdge(pred PlanNode, succ PlanNode) {
	pred.AddSuccessors(succ)
	succ.AddPredecessors(pred)
}

func TestPat(t *testing.T) {

	// Matches
	//     <anything> |> filter(...) |> filter(...)
	filterFilterPat := Pat(FilterKind, Pat(FilterKind, Any()))

	// Matches
	//   from(...) |> filter(...)
	filterFromPat := Pat(FilterKind, Pat(FromKind))

	from := &LogicalPlanNode{
		Spec: &FromProcedureSpec{},
	}

	filter1 := &LogicalPlanNode{
		Spec: &FilterProcedureSpec{},
	}

	addEdge(from, filter1)

	if filterFilterPat.Match(filter1) {
		t.Fatalf("Unexpected match")
	}

	if ! filterFromPat.Match(filter1) {
		t.Fatalf("Expected match")
	}

	filter2 := &LogicalPlanNode{
		Spec: &FilterProcedureSpec{},
	}

	addEdge(filter1, filter2)

	// Now we have
	//     from |> filter1 |> filter2

	if ! filterFilterPat.Match(filter2) {
		t.Fatalf("Expected match")
	}

	if filterFromPat.Match(filter2) {
		t.Fatalf("Unexpected match")
	}

	// Add another successor to filter1.  Thus should break the filter-filter pattern

	filter3 := &LogicalPlanNode{
		Spec: &FilterProcedureSpec{},
	}

	addEdge(filter1, filter3)

	// Now our graph looks like
	//     t = from |> filter1
	//     filter2(t)
	//     filter3(t)

	if filterFilterPat.Match(filter3) || filterFilterPat.Match(filter2) {
		t.Fatalf("Unexpected match")
	}

	if ! filterFromPat.Match(filter1) {
		t.Fatalf("Expected match")
	}
}
