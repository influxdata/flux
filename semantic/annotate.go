package semantic

import (
	"fmt"
	"strconv"
	"strings"
)

// TypeVar is a type variable
type TypeVar struct {
	name string
}

func (tv TypeVar) MonoType() (Type, bool) {
	return nil, false
}

func (tv TypeVar) Substitute(c Constraint) Substitutable {
	if tv == c.left {
		return c.right
	}
	return tv
}

func (tv TypeVar) String() string {
	return tv.name
}

type TypeEnvironment map[Node]TypeVar

func (e TypeEnvironment) String() string {
	var builder strings.Builder
	builder.WriteString("{\n")
	for n, tv := range e {
		fmt.Fprintf(&builder, "%#v=%v,\n", n, tv)
	}
	builder.WriteString("}")
	return builder.String()
}

// TypeAnnotationVisitor visits the nodes in a semantic
// graph and assigns all sub-expressions type variables.
type TypeAnnotationVisitor struct {
	// tenv represents the type environment of a program.
	// It maps a semantic node to a unique type variable.
	tenv TypeEnvironment
	// the next type variable to be assigned. The first
	// type variable is t0. Then t1, t2, t3 and so on.
	next *TypeVarGenerator
}

// Visit assigns a new type variable to a semantic node
func (v *TypeAnnotationVisitor) Visit(node Node) Visitor {
	switch n := node.(type) {
	case Expression:
		v.tenv[n] = v.next.NewTypeVar()
	case *NativeVariableDeclaration:
		v.tenv[n] = v.next.NewTypeVar()
	case *FunctionBody:
		// The function body type annotation corresponds to the return type of the function
		v.tenv[n] = v.next.NewTypeVar()
	case *FunctionParam:
		v.tenv[n] = v.next.NewTypeVar()
	}
	return v
}

// Done is required to implement the Visitor Interface
func (v *TypeAnnotationVisitor) Done() {}

// NewTypeAnnotationVisitor returns a new TypeAnnotationVisitor
func NewTypeAnnotationVisitor() *TypeAnnotationVisitor {
	return &TypeAnnotationVisitor{
		tenv: make(TypeEnvironment, 64),
		next: &TypeVarGenerator{
			prefix: "t",
			suffix: 0,
		},
	}
}

// TypeEnvironment returns the mapping between nodes and type variables
func (v *TypeAnnotationVisitor) TypeEnvironment() TypeEnvironment {
	return v.tenv
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

func Annotate(program *Program) TypeEnvironment {
	visitor := NewTypeAnnotationVisitor()
	Walk(visitor, program)
	return visitor.TypeEnvironment()
}
