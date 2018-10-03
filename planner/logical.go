package planner

// LogicalProcedureSpec is just a ProcedureSpec.
type LogicalProcedureSpec interface {
	Kind() ProcedureKind
	Copy() ProcedureSpec
}

// LogicalPlanNode consists of the input and output edges and a procedure spec
// that describes what the node does.
type LogicalPlanNode struct {
	Edges
	Spec LogicalProcedureSpec
}

func (lpn *LogicalPlanNode) ProcedureSpec() ProcedureSpec {
	return lpn.Spec
}
