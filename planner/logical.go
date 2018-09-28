package planner

import "github.com/influxdata/flux/plan"

// LogicalProcedureSpec is just a ProcedureSpec.
type LogicalProcedureSpec interface {
	Kind() plan.ProcedureKind
	Copy() plan.ProcedureSpec
}

// LogicalPlanNode consists of the input and output edges and a procedure spec
// that describes what the node does.
type LogicalPlanNode struct {
	Edges
	procedureSpec LogicalProcedureSpec
}

func (lpn *LogicalPlanNode) ProcedureSpec() plan.ProcedureSpec {
	return lpn.procedureSpec
}
