package plan

const AnyKind = "*** any procedure kind ***"

// Pattern represents an operator tree pattern
// It can match itself against a query plan
type Pattern interface {
	Roots() []ProcedureKind
	Match(Node) bool
}

// Single returns a pattern that can match a plan node with the given ProcedureKind
// and whose predecessors match the given predecessor patterns.
// The matched node must have exactly one successor.
//
// For example, to construct a pattern that matches a join followed by a sum:
//
//   sum
//    |
//   |X|    <=>  join(A, B) |> sum()  <=>  Pat(SumKind, Pat(JoinKind, Any(), Any()))
//  /   \
// A     B
func Single(kind ProcedureKind, predecessors ...Pattern) Pattern {
	return SingleOneOf([]ProcedureKind{kind}, predecessors...)
}

// Multi returns a pattern that can match a plan node with the given ProcedureKind
// and whose predecessors match the given predecessor patterns.
// The matched node may have any number including zero of successors.
//
// For example, to construct a pattern that matches a join followed by a sum:
//
//   sum
//    |
//   |X|    <=>  join(A, B) |> sum()  <=>  Pat(SumKind, Pat(JoinKind, Any(), Any()))
//  /   \
// A     B
func Multi(kind ProcedureKind, predecessors ...Pattern) Pattern {
	return MultiOneOf([]ProcedureKind{kind}, predecessors...)
}

// SingleOneOf matches any plan node from a given set of ProcedureKind and whose
// predecessors match the given predecessor patterns. This is identical to Pat,
// except for matching any pattern root from a set of ProcedureKinds.
func SingleOneOf(kinds []ProcedureKind, predecessors ...Pattern) Pattern {
	return &UnionKindPattern{
		kinds:        kinds,
		predecessors: predecessors,
		single:       true,
	}
}
func MultiOneOf(kinds []ProcedureKind, predecessors ...Pattern) Pattern {
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
		if node.Kind() == kind {
			found = true
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

// AnyPattern describes (and matches) any plan node
type AnyPattern struct{}

func (AnyPattern) Roots() []ProcedureKind {
	return []ProcedureKind{AnyKind}
}

func (AnyPattern) Match(node Node) bool {
	return true
}
