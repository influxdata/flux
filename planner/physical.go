package planner

// PhysicalProcedureSpec is similar to its logical counterpart but must provide a method to determine cost
type PhysicalProcedureSpec interface {
	Kind() ProcedureKind
	Copy() ProcedureSpec
	Cost(inStats []Statistics) (cost Cost, outStats Statistics)
}

type PhysicalPlanNode struct {
	Edges
	id NodeID
	Spec PhysicalProcedureSpec

	// The attributes required from inputs to this node
	RequiredAttrs []PhysicalAttributes

	// The attributes provided to consumers of this node's output
	OutputAttrs PhysicalAttributes
}

func (ppn *PhysicalPlanNode) ID() NodeID {
	return ppn.id
}

func (ppn *PhysicalPlanNode) ProcedureSpec() ProcedureSpec {
	return ppn.Spec
}

func (ppn *PhysicalPlanNode) Kind() ProcedureKind {
	return ppn.Spec.Kind()
}

func (ppn *PhysicalPlanNode) Cost(inStats []Statistics) (cost Cost, outStats Statistics) {
	return ppn.Spec.Cost(inStats)
}

type PhysicalAttributes struct {
	// Any physical attributes of the result produced by a physical plan node:
	// Collation, etc.
}
