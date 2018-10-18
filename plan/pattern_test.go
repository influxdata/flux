package planner_test

import (
	"testing"

	"github.com/influxdata/flux/functions/inputs"
	"github.com/influxdata/flux/functions/transformations"
	"github.com/influxdata/flux/planner"
)

func TestAny(t *testing.T) {
	pat := planner.Any()

	node := &planner.LogicalPlanNode{
		Spec: &inputs.FromProcedureSpec{},
	}

	if !pat.Match(node) {
		t.Fail()
	}
}

func addEdge(pred planner.PlanNode, succ planner.PlanNode) {
	pred.AddSuccessors(succ)
	succ.AddPredecessors(pred)
}

func TestPat(t *testing.T) {

	// Matches
	//     <anything> |> filter(...) |> filter(...)
	filterFilterPat := planner.Pat(transformations.FilterKind, planner.Pat(transformations.FilterKind, planner.Any()))

	// Matches
	//   from(...) |> filter(...)
	filterFromPat := planner.Pat(transformations.FilterKind, planner.Pat(inputs.FromKind))

	from := &planner.LogicalPlanNode{
		Spec: &inputs.FromProcedureSpec{},
	}

	filter1 := &planner.LogicalPlanNode{
		Spec: &transformations.FilterProcedureSpec{},
	}

	addEdge(from, filter1)

	if filterFilterPat.Match(filter1) {
		t.Fatalf("Unexpected match")
	}

	if !filterFromPat.Match(filter1) {
		t.Fatalf("Expected match")
	}

	filter2 := &planner.LogicalPlanNode{
		Spec: &transformations.FilterProcedureSpec{},
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

	filter3 := &planner.LogicalPlanNode{
		Spec: &transformations.FilterProcedureSpec{},
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
