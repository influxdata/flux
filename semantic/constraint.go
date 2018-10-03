package semantic

import (
	"fmt"
	"log"
	"strings"

	"github.com/influxdata/flux/ast"
)

// Constraint represents an equality constraint between type expressions
type Constraint struct {
	left  TypeVar
	right Substitutable
}

func (c Constraint) String() string {
	return fmt.Sprintf("%v ↦ %v", c.left, c.right)
}

func (c Constraint) MonoType() (Type, bool) {
	return c.right.MonoType()
}

func (c Constraint) Substitute(o Constraint) Substitutable {
	c.right = c.right.Substitute(o)
	return c
}

// CallConstraint represents an equality constraint between a type expression and the return type of an callee expression
type CallConstraint struct {
	left   TypeVar
	callee Substitutable
}

func (c CallConstraint) String() string {
	typ, mono := c.MonoType()
	if mono {
		return fmt.Sprintf("%v ↦ %v", c.left, typ)
	}
	return fmt.Sprintf("%v ↦ =>(%v)", c.left, c.callee)
}

func (c CallConstraint) MonoType() (Type, bool) {
	return c.callee.MonoType()
}

func (c CallConstraint) Substitute(o Constraint) Substitutable {
	c.callee = c.callee.Substitute(o)
	ft, mono := c.callee.MonoType()
	if mono && ft.Kind() == Function {
		c.callee = ft.ReturnType()
	}
	return c
}

// Substitutable represents any type expression containing type variables
type Substitutable interface {
	// Substitute returns a new substitutable with the constraint applied
	Substitute(Constraint) Substitutable
	// MonoType returns the concrete type of the substitutable expression if it exists
	MonoType() (Type, bool)
}

type arrayTypeScheme struct {
	elementType Substitutable
}

func (a arrayTypeScheme) MonoType() (Type, bool) {
	elementType, mono := a.elementType.MonoType()
	if !mono {
		return nil, false
	}
	return NewArrayType(elementType), true
}

func (a arrayTypeScheme) Substitute(c Constraint) Substitutable {
	a.elementType = a.elementType.Substitute(c)
	return a
}

func (a arrayTypeScheme) String() string {
	return fmt.Sprintf("[%v]", a.elementType)
}

type objectTypeScheme struct {
	properties map[string]Substitutable
}

func (o objectTypeScheme) MonoType() (Type, bool) {
	types := make(map[string]Type, len(o.properties))
	for k, p := range o.properties {
		t, m := p.MonoType()
		if !m {
			return nil, false
		}
		types[k] = t
	}
	return NewObjectType(types), true
}

func (o objectTypeScheme) Substitute(c Constraint) Substitutable {
	no := objectTypeScheme{
		properties: make(map[string]Substitutable, len(o.properties)),
	}
	for k, p := range o.properties {
		no.properties[k] = p.Substitute(c)
	}
	return no
}

func (o objectTypeScheme) String() string {
	var builder strings.Builder
	builder.WriteString("{")
	for k, p := range o.properties {
		fmt.Fprintf(&builder, "%v", k)
		builder.WriteString("=")
		fmt.Fprintf(&builder, "%v", p)
		builder.WriteString(", ")
	}
	builder.WriteString("}")
	return builder.String()
}

type functionTypeScheme struct {
	params     map[string]Substitutable
	returnType Substitutable
}

func (f functionTypeScheme) MonoType() (Type, bool) {
	rt, mono := f.returnType.MonoType()
	if !mono {
		return nil, false
	}
	types := make(map[string]Type, len(f.params))
	for k, p := range f.params {
		t, m := p.MonoType()
		if !m {
			return nil, false
		}
		types[k] = t
	}
	return NewFunctionType(FunctionSignature{
		Params:     types,
		ReturnType: rt,
	}), true
}

func (f functionTypeScheme) Substitute(c Constraint) Substitutable {
	nf := functionTypeScheme{
		params: make(map[string]Substitutable, len(f.params)),
	}
	for k, p := range f.params {
		nf.params[k] = p.Substitute(c)
	}
	nf.returnType = f.returnType.Substitute(c)
	return nf
}

func (f functionTypeScheme) String() string {
	var builder strings.Builder
	builder.WriteString("(")
	for k, p := range f.params {
		fmt.Fprintf(&builder, "%v", k)
		builder.WriteString("=")
		fmt.Fprintf(&builder, "%v", p)
		builder.WriteString(", ")
	}
	builder.WriteString(") => ")
	fmt.Fprintf(&builder, "%v", f.returnType)
	return builder.String()
}

// ConstraintGenerationVisitor visits a semantic graph and generates
// constraints between type variables and type expressions.
type ConstraintGenerationVisitor struct {
	tenv  map[Node]TypeVar
	cons  []Constraint
	scope *IdentifierScope
}

func NewConstraintGenerationVisitor(tenv map[Node]TypeVar) *ConstraintGenerationVisitor {
	return &ConstraintGenerationVisitor{
		tenv:  tenv,
		scope: NewIdentifierScope(),
	}
}

func (v *ConstraintGenerationVisitor) nest() *ConstraintGenerationVisitor {
	return &ConstraintGenerationVisitor{
		tenv:  v.tenv,
		cons:  v.cons,
		scope: v.scope.Nest(),
	}
}

func (v *ConstraintGenerationVisitor) Constraints() []Constraint {
	return v.cons
}

func (v *ConstraintGenerationVisitor) TypeEnvironment() map[Node]TypeVar {
	return v.tenv
}

// Visit a semantic node and generate a type constraint
func (v *ConstraintGenerationVisitor) Visit(node Node) Visitor {
	tv := v.tenv[node]
	switch n := node.(type) {
	case *BlockStatement:
		return v.nest()
	case *FunctionBody:
		// TODO(nathanielc): Handle case were Argument is not annotated because it is a node
		argumentVar := v.tenv[n.Argument]
		v.cons = append(v.cons, Constraint{
			left:  tv,
			right: argumentVar,
		})
		return v.nest()
	case *NativeVariableDeclaration:
		v.scope.Set(n.Identifier.Name, n)
		// Inference Rule: Variable Declaration
		// ------------------------------------
		// x = expression
		//
		// -> typeof(x) = typeof(expression)
		// ------------------------------------
		v.cons = append(v.cons, Constraint{
			left:  v.tenv[n.Identifier],
			right: v.tenv[n.Init],
		})
	case *FunctionExpression:
		// Inference Rule: Function Expression
		// -----------------------------------------------------------------
		// f = (a, b) => {
		//     Statement
		//     Statement
		//     Return Statement
		// }
		//
		// -> typeof(f) = (typeof(a), typeof(b)) => typeof(Return Statement)
		// -----------------------------------------------------------------
		returnTypeVar := v.tenv[n.Body]
		funcType := functionTypeScheme{
			params:     make(map[string]Substitutable, len(n.Params)),
			returnType: returnTypeVar,
		}
		for _, param := range n.Params {
			funcType.params[param.Key.Name] = v.tenv[param.Key]
		}
		v.cons = append(v.cons, Constraint{
			left:  tv,
			right: funcType,
		})
	case *FunctionParam:
		// maintain scope
		v.scope.Set(n.Key.Name, n.Key)

		key := v.tenv[n.Key]
		def, ok := v.tenv[n.Default]
		if ok {
			v.cons = append(v.cons, Constraint{
				left:  key,
				right: def,
			})
		}
	case *CallExpression:
		// Find FunctionBody and add constraint that typeof(body) == tv
		fe, err := v.lookupFunctionExpression(n.Callee)
		if err != nil {
			log.Println(err)
			return nil
		}
		funcBodyTypeVar := v.tenv[fe.Body]
		v.cons = append(v.cons, Constraint{
			left:  tv,
			right: funcBodyTypeVar,
		})
	case *UnaryExpression:
		// Inference Rule: Unary Expression
		// --------------------------------
		// !(expression) <=> NOT (Boolean)
		// -(expression) <=> MINUS (Int)
		// --------------------------------
		switch n.Operator {
		case ast.NotOperator:
			v.cons = append(v.cons,
				Constraint{
					left:  tv,
					right: Bool,
				},
				Constraint{
					left:  v.tenv[n.Argument],
					right: Bool,
				},
			)
		case ast.SubtractionOperator:
			// TODO: Negation well defined for floats?
			v.cons = append(v.cons,
				Constraint{
					left:  tv,
					right: Int,
				},
				Constraint{
					left:  v.tenv[n.Argument],
					right: Int,
				},
			)
		}
	case *LogicalExpression:
		// Inference Rule: Logical Expression
		// ---------------------------------------
		// Logical expressions yield boolean types
		// Logical operators act on boolean types
		//
		// x = left ( AND, OR ) right
		//
		// -> typeof(x) = bool
		// -> typeof(left) = bool
		// -> typeof(right) = bool
		// ---------------------------------------
		v.cons = append(v.cons,
			Constraint{
				left:  tv,
				right: Bool,
			},
			Constraint{
				left:  v.tenv[n.Left],
				right: Bool,
			},
			Constraint{
				left:  v.tenv[n.Right],
				right: Bool,
			},
		)
	case *BinaryExpression:
		switch n.Operator {
		// Inference Rules: Arithmetic Operators
		// --------------------------------------------
		// Arithmetic operands must be of the same type
		//
		// x = a ( + , - , * , / ) b
		//
		// -> typeof(x) = typeof(a) = typeof(b)
		// --------------------------------------------
		case
			ast.AdditionOperator,
			ast.SubtractionOperator,
			ast.MultiplicationOperator,
			ast.DivisionOperator:
			v.cons = append(v.cons,
				Constraint{
					left:  tv,
					right: v.tenv[n.Left],
				},
				Constraint{
					left:  tv,
					right: v.tenv[n.Right],
				},
				Constraint{
					left:  v.tenv[n.Right],
					right: v.tenv[n.Left],
				},
			)
		// Inference Rules: Comparison Operators
		// ------------------------------------------
		// Comparison operators return boolean values
		//
		// x = left ( <, <=, >, >=, ==, != ) right
		//
		// -> typeof(x) = Bool
		// ------------------------------------------
		case
			ast.GreaterThanEqualOperator,
			ast.LessThanEqualOperator,
			ast.GreaterThanOperator,
			ast.LessThanOperator,
			ast.NotEqualOperator,
			ast.EqualOperator:
			v.cons = append(v.cons, Constraint{
				left:  tv,
				right: Bool,
			})
		// Inference Rules: Regex Operators
		// ---------------------------------------------
		// Regex operators return boolean values
		// The type of the left operand must be a string
		// The type of the right operand must be a regex
		//
		// x = left ( =~, !~ ) right
		//
		// -> typeof(x) = Bool
		// -> typeof(left) = String
		// -> typeof(right) = Regex
		// ---------------------------------------------
		case
			ast.RegexpMatchOperator,
			ast.NotRegexpMatchOperator:
			v.cons = append(v.cons,
				Constraint{
					left:  tv,
					right: Bool,
				},
				Constraint{
					left:  v.tenv[n.Left],
					right: String,
				},
				Constraint{
					left:  v.tenv[n.Right],
					right: Regexp,
				},
			)
		}
	case *ArrayExpression:
		// Inference Rule: Array Expressions
		// -------------------------------------------------
		// All elements of an array must be of the same type
		//
		// x = [a, b, c]
		//
		// -> typeof(a) = typeof(b) = typeof(c)
		// -------------------------------------------------
		v.cons = append(v.cons, Constraint{
			left: v.tenv[n],
			right: arrayTypeScheme{
				elementType: v.tenv[n.Elements[0]],
			},
		})
		for _, e := range n.Elements {
			v.cons = append(v.cons, Constraint{
				left:  v.tenv[n.Elements[0]],
				right: v.tenv[e],
			})
		}
	case *ObjectExpression:
		// Object expressions generate trivial constraints
		v.cons = append(v.cons, Constraint{
			left: tv,
			right: objectTypeScheme{
				properties: func() map[string]Substitutable {
					signature := make(map[string]Substitutable, len(n.Properties))
					for _, prop := range n.Properties {
						signature[prop.Key.Name] = v.tenv[prop.Value]
					}
					return signature
				}(),
			},
		})
	case *MemberExpression:
		// TODO: This is probably the most difficult type
		// inference rule. How to constrain this type?
	case *IdentifierExpression:
		declNode, found := v.scope.Lookup(n.Name)
		if !found {
			log.Printf("missing identifier %q", n.Name)
			return nil
		}
		tvar := v.tenv[declNode]
		v.cons = append(v.cons, Constraint{
			left:  tvar,
			right: tv,
		})
	case *BooleanLiteral:
		v.cons = append(v.cons, Constraint{
			left:  tv,
			right: Bool,
		})
	case *DateTimeLiteral:
		v.cons = append(v.cons, Constraint{
			left:  tv,
			right: Time,
		})
	case *DurationLiteral:
		v.cons = append(v.cons, Constraint{
			left:  tv,
			right: Duration,
		})
	case *FloatLiteral:
		v.cons = append(v.cons, Constraint{
			left:  tv,
			right: Float,
		})
	case *IntegerLiteral:
		v.cons = append(v.cons, Constraint{
			left:  tv,
			right: Int,
		})
	case *RegexpLiteral:
		v.cons = append(v.cons, Constraint{
			left:  tv,
			right: Regexp,
		})
	case *StringLiteral:
		v.cons = append(v.cons, Constraint{
			left:  tv,
			right: String,
		})
	case *UnsignedIntegerLiteral:
		v.cons = append(v.cons, Constraint{
			left:  tv,
			right: UInt,
		})
	}
	return v
}

// Done is used to satisfy the Visitor interface
func (v *ConstraintGenerationVisitor) Done() {}

func (v *ConstraintGenerationVisitor) lookupFunctionExpression(callee Expression) (*FunctionExpression, error) {
	switch n := callee.(type) {
	case *FunctionExpression:
		return n, nil
	case *IdentifierExpression:
		declNode, found := v.scope.Lookup(n.Name)
		if !found {
			return nil, fmt.Errorf("unknown identifier %q", n.Name)
		}
		fe, ok := declNode.(*FunctionExpression)
		if !ok {
			return nil, fmt.Errorf("cannot call non-function %q", n.Name)
		}
		return fe, nil
	default:
		return nil, fmt.Errorf("unsupported callee type %T", callee)
	}
}

// IdentifierScope of the program
type IdentifierScope struct {
	parent *IdentifierScope
	// Identifiers in the current scope
	vars map[string]Node
}

// NewIdentifierScope returns a new variable scope
func NewIdentifierScope() *IdentifierScope {
	return &IdentifierScope{
		vars: make(map[string]Node, 8),
	}
}

// Set adds a new binding to the current scope
func (s *IdentifierScope) Set(name string, node Node) {
	s.vars[name] = node
}

// Lookup returns the variable declaration associated with name in the current scope
func (s *IdentifierScope) Lookup(name string) (Node, bool) {
	if s == nil {
		return nil, false
	}
	dec, ok := s.vars[name]
	if !ok {
		return s.parent.Lookup(name)
	}
	return dec, ok
}

// Nest returns a new variable scope whose parent is the current scope
func (s *IdentifierScope) Nest() *IdentifierScope {
	return &IdentifierScope{
		parent: s,
		vars:   make(map[string]Node, 8),
	}
}

// Parent returns the parent scope of the current scope
func (s *IdentifierScope) Parent() *IdentifierScope {
	return s.parent
}

func GenerateConstraints(program *Program, tenv map[Node]TypeVar) []Substitutable {
	// Generate the rest of the constraints
	constraintVisitor := NewConstraintGenerationVisitor(tenv)
	Walk(constraintVisitor, program)

	return constraintVisitor.Constraints()
}
