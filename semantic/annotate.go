package semantic

import "strconv"

// TypeVar is a type variable
type TypeVar struct {
	name string
}

// FreeTypeVar returns all the free (not bound) type variables of an expression.
// In the case of a type variable, it just returns the same type variable.
func (tv TypeVar) FreeTypeVar() []TypeVar {
	return []TypeVar{tv}
}

func (tv TypeVar) String() string {
	return tv.name
}

func (tv TypeVar) Equal(sub Substitutable) bool {
	if tvar, ok := sub.(TypeVar); ok {
		return tv.name == tvar.name
	}
	return false
}

// TypeVarGenerator generates type variables
type TypeVarGenerator struct {
	prefix string
	suffix int
}

// NewTypeVar returns the a new type variable from a generator
func (v *TypeVarGenerator) NewTypeVar() TypeVar {
	name := v.prefix + strconv.Itoa(v.suffix)
	v.suffix++
	return TypeVar{name: name}
}

// TypeAnnotationVisitor visits the nodes in a semantic
// graph and assigns all sub-expressions type variables.
type TypeAnnotationVisitor struct {
	// tenv represents the type environment of a program.
	// It maps a semantic node to a unique type variable.
	tenv map[Node]TypeVar
	// the next type variable to be assigned. The first
	// type variable is t0. Then t1, t2, t3 and so on.
	next *TypeVarGenerator
}

// Visit assigns a new type variable to a semantic node
func (v *TypeAnnotationVisitor) Visit(node Node) Visitor {
	if n, ok := node.(Expression); ok {
		v.tenv[n] = v.next.NewTypeVar()
	}
	if n, ok := node.(*Identifier); ok {
		v.tenv[n] = v.next.NewTypeVar()
	}
	return v
}

// Done is required to implement the Visitor Interface
func (v *TypeAnnotationVisitor) Done() {}

// NewTypeAnnotationVisitor returns a new TypeAnnotationVisitor
func NewTypeAnnotationVisitor() *TypeAnnotationVisitor {
	return &TypeAnnotationVisitor{
		tenv: make(map[Node]TypeVar, 64),
		next: &TypeVarGenerator{
			prefix: "t",
			suffix: 0,
		},
	}
}

// TypeEnvironment returns the mapping between nodes and type variables
func (v *TypeAnnotationVisitor) TypeEnvironment() map[Node]TypeVar {
	return v.tenv
}
