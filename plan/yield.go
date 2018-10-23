package plan

const generatedYieldKind = "generatedYield"

type YieldProcedureSpec interface {
	YieldName() string
}

// generatedYieldSpec provides a special planner-generated yield for queries that don't
// have explicit calls to yield().
type generatedYieldProcedureSpec struct {
	name string
}

func (y generatedYieldProcedureSpec) Kind() ProcedureKind {
	return generatedYieldKind
}

func (y generatedYieldProcedureSpec) Copy() ProcedureSpec {
	return generatedYieldProcedureSpec{name: y.name}
}

func (y generatedYieldProcedureSpec) YieldName() string {
	return y.name
}
