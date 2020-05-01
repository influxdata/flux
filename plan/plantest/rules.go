package plantest

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/stdlib/universe"
)

// SimpleRule is a simple rule whose pattern matches any plan node and
// just stores the NodeIDs of nodes it has visited in SeenNodes.
type SimpleRule struct {
	SeenNodes []plan.NodeID
}

func (sr *SimpleRule) Pattern() plan.Pattern {
	return plan.Any()
}

func (sr *SimpleRule) Rewrite(ctx context.Context, node plan.Node) (plan.Node, bool, error) {
	for _, nid := range sr.SeenNodes {
		if nid == node.ID() {
			return node, false, nil
		}
	}
	sr.SeenNodes = append(sr.SeenNodes, node.ID())
	return node, false, nil
}

func (sr *SimpleRule) Name() string {
	return "simple"
}

// FunctionRule is a simple rule intended to invoke a Rewrite function.
type FunctionRule struct {
	RewriteFn func(ctx context.Context, node plan.Node) (plan.Node, bool, error)
}

func (fr *FunctionRule) Name() string {
	return "function"
}

func (fr *FunctionRule) Pattern() plan.Pattern {
	return plan.Any()
}

func (fr *FunctionRule) Rewrite(ctx context.Context, node plan.Node) (plan.Node, bool, error) {
	return fr.RewriteFn(ctx, node)
}

// SmashPlanRule adds an `Intruder` as predecessor of the given `Node` without
// marking it as successor of it. It breaks the integrity of the plan.
// If `Kind` is specified, it takes precedence over `Node`, and the rule will use it
// to match.
type SmashPlanRule struct {
	Node     plan.Node
	Intruder plan.Node
	Kind     plan.ProcedureKind
}

func (SmashPlanRule) Name() string {
	return "SmashPlanRule"
}

func (spp SmashPlanRule) Pattern() plan.Pattern {
	var k plan.ProcedureKind
	if len(spp.Kind) > 0 {
		k = spp.Kind
	} else {
		k = spp.Node.Kind()
	}

	return plan.Pat(k, plan.Any())
}

func (spp SmashPlanRule) Rewrite(ctx context.Context, node plan.Node) (plan.Node, bool, error) {
	var changed bool
	if len(spp.Kind) > 0 || node == spp.Node {
		node.AddPredecessors(spp.Intruder)
		changed = true
	}

	// it is not necessary to return a copy of the node, because the rule changes the number
	// of predecessors and it won't be re-triggered again.
	return node, changed, nil
}

// CreateCycleRule creates a cycle between the given `Node` and its predecessor.
// It creates exactly one cycle. After the rule is triggered once, it won't have any effect later.
// This rule breaks the integrity of the plan.
// If `Kind` is specified, it takes precedence over `Node`, and the rule will use it
// to match.
type CreateCycleRule struct {
	Node plan.Node
	Kind plan.ProcedureKind
}

func (CreateCycleRule) Name() string {
	return "CreateCycleRule"
}

func (ccr CreateCycleRule) Pattern() plan.Pattern {
	var k plan.ProcedureKind
	if len(ccr.Kind) > 0 {
		k = ccr.Kind
	} else {
		k = ccr.Node.Kind()
	}

	return plan.Pat(k, plan.Any())
}

func (ccr CreateCycleRule) Rewrite(ctx context.Context, node plan.Node) (plan.Node, bool, error) {
	var changed bool
	if len(ccr.Kind) > 0 || node == ccr.Node {
		node.Predecessors()[0].AddPredecessors(node)
		node.AddSuccessors(node.Predecessors()[0])
		changed = true
	}

	// just return a copy of the node, otherwise the rule will be triggered an infinite number of times
	// (it doesn't change the number of predecessors, indeed).
	return node.ShallowCopy(), changed, nil
}

// MultiRoot matches a set of plan nodes at the root and stores the NodeIDs of
// nodes it has visited in SeenNodes.
type MultiRootRule struct {
	SeenNodes []plan.NodeID
}

func (sr *MultiRootRule) Pattern() plan.Pattern {
	return plan.OneOf(
		[]plan.ProcedureKind{
			universe.MinKind,
			universe.MaxKind,
			universe.MeanKind,
		},
		plan.Any())
}

func (sr *MultiRootRule) Rewrite(ctx context.Context, node plan.Node) (plan.Node, bool, error) {
	sr.SeenNodes = append(sr.SeenNodes, node.ID())
	return node, false, nil
}

func (sr *MultiRootRule) Name() string {
	return "multiroot"
}

// RuleTestCase allows for concise creation of test cases that exercise rules
type RuleTestCase struct {
	Name          string
	Context       context.Context
	Rules         []plan.Rule
	Before        *PlanSpec
	After         *PlanSpec
	NoChange      bool
	ValidateError error
}

// PhysicalRuleTestHelper will run a rule test case.
func PhysicalRuleTestHelper(t *testing.T, tc *RuleTestCase) {
	t.Helper()

	before := CreatePlanSpec(tc.Before)
	var after *plan.Spec
	if tc.NoChange || tc.ValidateError != nil {
		after = CreatePlanSpec(tc.Before.Copy())
	} else {
		after = CreatePlanSpec(tc.After)
	}

	opts := []plan.PhysicalOption{
		plan.OnlyPhysicalRules(tc.Rules...),
	}
	if tc.ValidateError == nil {
		// Disable validation so that we can avoid having to push a range into every from
		opts = append(opts, plan.DisableValidation())
	}
	physicalPlanner := plan.NewPhysicalPlanner(opts...)

	ctx := tc.Context
	if ctx == nil {
		ctx = context.Background()
	}

	pp, err := physicalPlanner.Plan(ctx, before)
	if err != nil {
		if tc.ValidateError != nil {
			if got, want := err, tc.ValidateError; !cmp.Equal(want, got) {
				t.Fatalf("unexpected planner error -want/+got:\n%s", cmp.Diff(want, got))
			}
			return
		}
		t.Fatal(err)
	} else if tc.ValidateError != nil {
		t.Fatal("expected planner error")
	}

	type testAttrs struct {
		ID   plan.NodeID
		Spec plan.PhysicalProcedureSpec
	}
	want := make([]testAttrs, 0)
	after.BottomUpWalk(func(node plan.Node) error {
		want = append(want, testAttrs{
			ID:   node.ID(),
			Spec: node.ProcedureSpec().(plan.PhysicalProcedureSpec),
		})
		return nil
	})

	got := make([]testAttrs, 0)
	pp.BottomUpWalk(func(node plan.Node) error {
		got = append(got, testAttrs{
			ID:   node.ID(),
			Spec: node.ProcedureSpec().(plan.PhysicalProcedureSpec),
		})
		return nil
	})

	if !cmp.Equal(want, got, CmpOptions...) {
		t.Errorf("transformed plan not as expected, -want/+got:\n%v",
			cmp.Diff(want, got, CmpOptions...))
	}
}

// LogicalRuleTestHelper will run a rule test case.
func LogicalRuleTestHelper(t *testing.T, tc *RuleTestCase) {
	t.Helper()

	before := CreatePlanSpec(tc.Before)
	var after *plan.Spec
	if tc.NoChange {
		after = CreatePlanSpec(tc.Before.Copy())
	} else {
		after = CreatePlanSpec(tc.After)
	}

	logicalPlanner := plan.NewLogicalPlanner(
		plan.OnlyLogicalRules(tc.Rules...),
	)

	ctx := tc.Context
	if ctx == nil {
		ctx = context.Background()
	}

	pp, err := logicalPlanner.Plan(ctx, before)
	if err != nil {
		t.Fatal(err)
	}

	type testAttrs struct {
		ID   plan.NodeID
		Spec plan.ProcedureSpec
	}
	want := make([]testAttrs, 0)
	after.BottomUpWalk(func(node plan.Node) error {
		want = append(want, testAttrs{
			ID:   node.ID(),
			Spec: node.ProcedureSpec(),
		})
		return nil
	})

	got := make([]testAttrs, 0)
	pp.BottomUpWalk(func(node plan.Node) error {
		got = append(got, testAttrs{
			ID:   node.ID(),
			Spec: node.ProcedureSpec(),
		})
		return nil
	})

	if !cmp.Equal(want, got, CmpOptions...) {
		t.Errorf("transformed plan not as expected, -want/+got:\n%v",
			cmp.Diff(want, got, CmpOptions...))
	}
}
