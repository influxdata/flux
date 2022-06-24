package plantest

import (
	"github.com/mvn-trinhnguyen2-dn/flux/plan"
	"github.com/mvn-trinhnguyen2-dn/flux/plan/plantest/spec"
)

//
// Export the plan/plantest/spec types and functions. This is isolated so it
// can be used in testing code for /execute, where the full plantest
// dependencies cause an import cycle.
//

type PlanSpec = spec.PlanSpec
type MockProcedureSpec = spec.MockProcedureSpec

func CreatePlanSpec(ps *PlanSpec) *plan.Spec {
	return spec.CreatePlanSpec(ps)
}

func CreateLogicalMockNode(id string) *plan.LogicalNode {
	return spec.CreateLogicalMockNode(id)
}

func CreatePhysicalMockNode(id string) *plan.PhysicalPlanNode {
	return spec.CreatePhysicalMockNode(id)
}
