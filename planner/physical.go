package planner

import "github.com/influxdata/flux/plan"

// PhysicalProcedureSpec is similar to its logical counterpart but must provide a method to determine cost
type PhysicalProcedureSpec interface {
	Kind() plan.ProcedureKind
	Copy() plan.ProcedureSpec
	Cost(inStats []Statistics) (cost Cost, outStats Statistics)
}

type PhysicalPlanNode struct {
	Edges
	procedureSpec PhysicalProcedureSpec

	// The attributes required from inputs to this node
	RequiredAttrs []PhysicalAttributes

	// The attributes provided to consumers of this node's output
	OutputAttrs PhysicalAttributes
}

func (ppn *PhysicalPlanNode) ProcedureSpec() plan.ProcedureSpec {
	return ppn.procedureSpec
}

func (ppn *PhysicalPlanNode) Cost(inStats []Statistics) (cost Cost, outStats Statistics) {
	return ppn.procedureSpec.Cost(inStats)
}

type PhysicalAttributes struct {
	// Any physical attributes of the result produced by a physical plan node:
	// Collation, etc.
}
