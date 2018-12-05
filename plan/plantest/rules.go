package plantest

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic/semantictest"
)

// SimpleRule is a simple rule whose pattern matches any plan node and
// just stores the NodeIDs of nodes it has visited in SeenNodes.
type SimpleRule struct {
	SeenNodes []plan.NodeID
}

func (sr *SimpleRule) Pattern() plan.Pattern {
	return plan.Any()
}

func (sr *SimpleRule) Rewrite(node plan.PlanNode) (plan.PlanNode, bool, error) {
	sr.SeenNodes = append(sr.SeenNodes, node.ID())
	return node, false, nil
}

func (sr *SimpleRule) Name() string {
	return "simple"
}

// RuleTestCase allows for concise creation of test cases that exercise rules
type RuleTestCase struct {
	Name     string
	Rules    []plan.Rule
	Before   *PlanSpec
	After    *PlanSpec
	NoChange bool
}

// RuleTestHelper will run a rule test case.
func RuleTestHelper(t *testing.T, tc *RuleTestCase) {
	t.Helper()

	before := CreatePlanSpec(tc.Before)
	var after *plan.PlanSpec
	if tc.NoChange {
		after = CreatePlanSpec(tc.Before.Copy())
	} else {
		after = CreatePlanSpec(tc.After)
	}

	// Disable validation so that we can avoid having to push a range into every from
	physicalPlanner := plan.NewPhysicalPlanner(
		plan.OnlyPhysicalRules(tc.Rules...),
		plan.DisableValidatation(),
	)

	pp, err := physicalPlanner.Plan(before)
	if err != nil {
		t.Fatal(err)
	}

	type testAttrs struct {
		ID   plan.NodeID
		Spec plan.ProcedureSpec
	}
	want := make([]testAttrs, 0)
	after.BottomUpWalk(func(node plan.PlanNode) error {
		want = append(want, testAttrs{
			ID:   node.ID(),
			Spec: node.ProcedureSpec(),
		})
		return nil
	})

	got := make([]testAttrs, 0)
	pp.BottomUpWalk(func(node plan.PlanNode) error {
		got = append(got, testAttrs{
			ID:   node.ID(),
			Spec: node.ProcedureSpec(),
		})
		return nil
	})

	if !cmp.Equal(want, got, semantictest.CmpOptions...) {
		t.Errorf("transformed plan not as expected, -want/+got:\n%v",
			cmp.Diff(want, got, semantictest.CmpOptions...))
	}
}
