package builtin

import (
	"fmt"

	"github.com/influxdata/flux/functions/inputs"
	"github.com/influxdata/flux/functions/transformations"
	"github.com/influxdata/flux/plan"
)

// Registers testing-specific rules to convert a logical `from` to a physical `mockFrom`
// and to bound it with a `range` operation. It must be invoked manually by tests when needed.
func init() {
	plan.RegisterPhysicalRules(MockFromConversionRule{})
	plan.RegisterPhysicalRules(MergeMockFromRangeRule{})
}

// This procedure spec is used in flux tests to represent the physical
// (yet storage-agnostic) counterpart of the logical `from` operation.
// A physical (yet mocked) representation is necessary to make physical planning succeed.
const mockFromKind = "mockFrom"

type MockFromProcedureSpec struct {
	*inputs.FromProcedureSpec
	plan.DefaultCost

	Bounded bool
}

func (MockFromProcedureSpec) Kind() plan.ProcedureKind {
	return mockFromKind
}

func (s *MockFromProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(MockFromProcedureSpec)
	ns.FromProcedureSpec = s.FromProcedureSpec.Copy().(*inputs.FromProcedureSpec)
	ns.Bounded = s.Bounded
	return ns
}

func (s MockFromProcedureSpec) PostPhysicalValidate(id plan.NodeID) error {
	if !s.Bounded {
		var bucket string
		if len(s.Bucket) > 0 {
			bucket = s.Bucket
		} else {
			bucket = s.BucketID
		}
		return fmt.Errorf(`%s: results from "%s" must be bounded`, id, bucket)
	}
	return nil
}

// MockFromConversionRule converts a logical `from` node into a physical `mockFrom` node for testing purposes.
type MockFromConversionRule struct {
}

func (MockFromConversionRule) Name() string {
	return "MockFromConversionRule"
}

func (MockFromConversionRule) Pattern() plan.Pattern {
	return plan.Pat(inputs.FromKind)
}

func (mfcr MockFromConversionRule) Rewrite(pn plan.PlanNode) (plan.PlanNode, bool, error) {
	logicalFromSpec := pn.ProcedureSpec().(*inputs.FromProcedureSpec)
	newNode := plan.CreatePhysicalNode(pn.ID(), &MockFromProcedureSpec{
		FromProcedureSpec: logicalFromSpec.Copy().(*inputs.FromProcedureSpec),
	})

	plan.ReplaceNode(pn, newNode)
	return newNode, true, nil
}

// MergeFromRangeRule pushes a `range` into a `mockFrom`.
// It doesn't compute any bound, but it makes the `mockFrom` bounded to pass physical validation.
// This rule is registered on module import. A `mockFrom` operation should exists only for testing
// purpose and if `MockFromConversionRule` has been registered.
type MergeMockFromRangeRule struct{}

func (rule MergeMockFromRangeRule) Name() string {
	return "MergeMockFromRangeRule"
}

func (rule MergeMockFromRangeRule) Pattern() plan.Pattern {
	return plan.Pat(transformations.RangeKind, plan.Pat(mockFromKind))
}

func (rule MergeMockFromRangeRule) Rewrite(node plan.PlanNode) (plan.PlanNode, bool, error) {
	from := node.Predecessors()[0]
	fromSpec := from.ProcedureSpec().(*MockFromProcedureSpec)
	fromRange := fromSpec.Copy().(*MockFromProcedureSpec)

	fromRange.Bounded = true

	// merge nodes into single operation
	merged, err := plan.MergeToPhysicalPlanNode(node, from, fromRange)
	if err != nil {
		return nil, false, err
	}

	return merged, true, nil
}
