package plan

const AnyKind = "*** any procedure kind ***"

// Pattern represents an operator tree pattern
// It can match itself against a query plan.
type Pattern interface {
	// Roots returns the list of procedure kinds that may appear
	// at the "root" of the pattern. "Root" here means a node with
	// some predecessors.
	Roots() []ProcedureKind
	// Match returns true when the given node and its predecessors match
	// with this pattern.
	Match(Node) bool
}

// SingleSuccessor returns a pattern that can match a plan node with the given ProcedureKind
// and whose predecessors match the given predecessor patterns.
// The matched node must have exactly one successor.
//
// For example, to construct a pattern that matches a sort node that succeeds a join:
//
//	 sort
//	  |
//	 join   <=>  join(A, B) |> sum()  <=>  MultiSuccessor(SortKind, SingleSuccessor(JoinKind, AnyMultiSuccessor(), AnyMultiSuccessor()))
//	/   \
//
// A     B
func SingleSuccessor(kind ProcedureKind, predecessors ...Pattern) Pattern {
	return SingleSuccessorOneOf([]ProcedureKind{kind}, predecessors...)
}

// MultiSuccessor returns a pattern that can match a plan node with the given ProcedureKind
// and whose predecessors match the given predecessor patterns.
// The matched node may have any number of successors, including zero.
func MultiSuccessor(kind ProcedureKind, predecessors ...Pattern) Pattern {
	return MultiSuccessorOneOf([]ProcedureKind{kind}, predecessors...)
}

// SingleSuccessorOneOf matches any plan node from a given set of ProcedureKind and whose
// predecessors match the given predecessor patterns. This is identical to SingleSuccessor,
// except for matching any pattern root from a set of ProcedureKinds.
func SingleSuccessorOneOf(kinds []ProcedureKind, predecessors ...Pattern) Pattern {
	return &UnionKindPattern{
		kinds:        kinds,
		predecessors: predecessors,
		single:       true,
	}
}

// MultiSuccessorOneOf matches any plan node from a given set of ProcedureKind and whose
// predecessors match the given predecessor patterns. This is identical to MultiSuccessor,
// except for matching any pattern root from a set of ProcedureKinds.
func MultiSuccessorOneOf(kinds []ProcedureKind, predecessors ...Pattern) Pattern {
	return &UnionKindPattern{
		kinds:        kinds,
		predecessors: predecessors,
		single:       false,
	}
}

// PhysPat returns a pattern that matches a physical plan node with the given
// ProcedureKind and whose predecessors match the given predecessor patterns.
func PhysPat(pat Pattern) Pattern {
	return PhysicalOneKindPattern{
		pattern: pat,
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

// UnionKindPattern matches any one of a set of procedures that have a
// specified predecessor pattern.
//
// For example, UnionKindPattern( { Proc1Kind, Proc2Kind }, { Pat1, Pat2 } )
// will match either Proc1Kind { Pat1, Pat2 } or Proc2Kind { Pat1, Pat2 }
//
//	         [ ProcedureKind ]
//	         /       |  ...  \
//	pattern1     pattern2  ... patternK
type UnionKindPattern struct {
	kinds        []ProcedureKind
	predecessors []Pattern
	// single indicates if the matched node must have exactly one successor
	single bool
}

func (ukp UnionKindPattern) Roots() []ProcedureKind {
	return ukp.kinds
}

func (ukp UnionKindPattern) Match(node Node) bool {
	if ukp.single && len(node.Successors()) != 1 {
		return false
	}
	found := false
	for _, kind := range ukp.kinds {
		if node.Kind() == kind || kind == AnyKind {
			found = true
			break
		}
	}
	if !found {
		return false
	}

	if len(ukp.predecessors) > 0 && len(ukp.predecessors) != len(node.Predecessors()) {
		return false
	}

	// Recursively match each predecessor
	for i, pattern := range ukp.predecessors {
		if !pattern.Match(node.Predecessors()[i]) {
			return false
		}
	}
	return true
}

// AnySingleSuccessor returns a pattern that matches any node
// that has a single successor.
func AnySingleSuccessor() Pattern {
	return &UnionKindPattern{
		kinds:  []ProcedureKind{AnyKind},
		single: true,
	}
}

// AnyMultiSuccessor returns a pattern that matches any node
// with any number of successors
func AnyMultiSuccessor() Pattern {
	return &UnionKindPattern{
		kinds:  []ProcedureKind{AnyKind},
		single: false,
	}
}
