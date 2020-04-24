package plan

const AnyKind = "*** any procedure kind ***"

// Pattern represents an operator tree pattern
// It can match itself against a query plan
type Pattern interface {
	Roots() []ProcedureKind
	Match(Node) bool
}

// Pat returns a pattern that can match a plan node with the given ProcedureKind
// and whose predecessors match the given predecessor patterns.
//
// For example, to construct a pattern that matches a join followed by a sum:
//
//   sum
//    |
//   |X|    <=>  join(A, B) |> sum()  <=>  Pat(SumKind, Pat(JoinKind, Any(), Any()))
//  /   \
// A     B
func Pat(kind ProcedureKind, predecessors ...Pattern) Pattern {
	return &UnionKindPattern{
		kinds:        []ProcedureKind{kind},
		predecessors: predecessors,
	}
}

// OneOf matches any plan node from a given set of ProcedureKind and whose
// predecessors match the given predecessor patterns. This is identical to Pat,
// except for matching any pattern root from a set of ProcedureKinds.
func OneOf(kinds []ProcedureKind, predecessors ...Pattern) Pattern {
	return &UnionKindPattern{
		kinds:        kinds,
		predecessors: predecessors,
	}
}

// PhysPat returns a pattern that matches a physical plan node with the given
// ProcedureKind and whose predecessors match the given predecessor patterns.
func PhysPat(kind ProcedureKind, predecessors ...Pattern) Pattern {
	return PhysicalOneKindPattern{
		pattern: Pat(kind, predecessors...),
	}
}

// PhysicalOneKindPattern matches a physical operator pattern
type PhysicalOneKindPattern struct {
	pattern Pattern
}

func (p PhysicalOneKindPattern) Roots() []ProcedureKind {
	return p.pattern.Roots()
}

func (p PhysicalOneKindPattern) Match(node Node) bool {
	_, ok := node.(*PhysicalPlanNode)
	return ok && p.pattern.Match(node)
}

// Any returns a pattern that matches anything.
func Any() Pattern {
	return &AnyPattern{}
}

// UnionKindPattern matches any one of a set of procedures that have a
// specified predecessor pattern.
//
// For example, UnionKindPattern( { Proc1Kind, Proc2Kind }, { Pat1, Pat2 } )
// will match either Proc1Kind { Pat1, Pat2 } or Proc2Kind { Pat1, Pat2 }
//
//                 [ ProcedureKind ]
//                 /       |  ...  \
//        pattern1     pattern2  ... patternK
type UnionKindPattern struct {
	kinds        []ProcedureKind
	predecessors []Pattern
}

func (okp UnionKindPattern) Roots() []ProcedureKind {
	return okp.kinds
}

func (okp UnionKindPattern) Match(node Node) bool {
	found := false
	for _, kind := range okp.kinds {
		if node.Kind() == kind {
			found = true
		}
	}
	if !found {
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
	for i, pattern := range okp.predecessors {
		if !pattern.Match(node.Predecessors()[i]) {
			return false
		}
	}
	return true
}

// AnyPattern describes (and matches) any plan node
type AnyPattern struct{}

func (AnyPattern) Roots() []ProcedureKind {
	return []ProcedureKind{AnyKind}
}

func (AnyPattern) Match(node Node) bool {
	return true
}
