package planner

const AnyKind = "*** any procedure kind ***"

// Pattern represents an operator tree pattern
// It can match itself against a query plan
type Pattern interface {
	Root() ProcedureKind
	Match(PlanNode) bool
}

func Pat(kind ProcedureKind, predecessors ...Pattern) Pattern {
	return &OneKindPattern{
		kind:         kind,
		predecessors: predecessors,
	}
}

func Any() Pattern {
	return &AnyPattern{}
}

type OneKindPattern struct {
	kind         ProcedureKind
	predecessors []Pattern
}

func (okp *OneKindPattern) Root() ProcedureKind {
	return okp.kind
}

func (okp *OneKindPattern) Match(node PlanNode) bool {
	if node.Kind() != okp.kind {
		return false
	}

	if len(okp.predecessors) != len(node.Predecessors()) {
		return false
	}

	// Check that each predecessor does not have other successors
	for _, pred := range node.Predecessors() {
		if len(pred.Successors()) != 1 {
			return false
		}
	}

	// Recursively match each predecessor
	for i, pred := range node.Predecessors() {
		if !okp.predecessors[i].Match(pred) {
			return false
		}
	}
	return true
}

type AnyPattern struct {
}

func (ap *AnyPattern) Root() ProcedureKind {
	return AnyKind
}

func (ap *AnyPattern) Match(node PlanNode) bool {
	return true
}
