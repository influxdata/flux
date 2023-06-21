package spec

import "github.com/InfluxCommunity/flux/plan"

const MockKind = "mock"

// CreateLogicalMockNode creates a mock plan node that doesn't match any rules
// (other than rules that match any node)
func CreateLogicalMockNode(id string) *plan.LogicalNode {
	return plan.CreateLogicalNode(plan.NodeID(id), MockProcedureSpec{})
}

// CreatePhysicalMockNode creates a mock plan node that doesn't match any rules
// (other than rules that match any node)
func CreatePhysicalMockNode(id string) *plan.PhysicalPlanNode {
	return plan.CreatePhysicalNode(plan.NodeID(id), MockProcedureSpec{})
}

// MockProcedureSpec provides a type that implements ProcedureSpec but does not require
// importing packages which register rules and procedure kinds, which makes it useful for
// unit testing.
type MockProcedureSpec struct {
	plan.DefaultCost
	OutputAttributesFn     func() plan.PhysicalAttributes
	PassThroughAttributeFn func(attrKey string) bool
	RequiredAttributesFn   func() []plan.PhysicalAttributes
	PlanDetailsFn          func() string
}

func (s MockProcedureSpec) PlanDetails() string {
	if s.PlanDetailsFn != nil {
		return s.PlanDetailsFn()
	}
	return ""
}

func (s MockProcedureSpec) OutputAttributes() plan.PhysicalAttributes {
	if s.OutputAttributesFn != nil {
		return s.OutputAttributesFn()
	}
	return nil
}

func (s MockProcedureSpec) PassThroughAttribute(attrKey string) bool {
	if s.PassThroughAttributeFn != nil {
		return s.PassThroughAttributeFn(attrKey)
	}
	return false
}

func (s MockProcedureSpec) RequiredAttributes() []plan.PhysicalAttributes {
	if s.RequiredAttributesFn != nil {
		return s.RequiredAttributesFn()
	}
	return nil
}

func (MockProcedureSpec) Kind() plan.ProcedureKind {
	return MockKind
}

func (MockProcedureSpec) Copy() plan.ProcedureSpec {
	return MockProcedureSpec{}
}
