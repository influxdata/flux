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
func (c Constraint) typeScheme() {}

func (c Constraint) Substitute(o Constraint) Substitutable {
	c.right = c.right.Substitute(o)
	return c
}

func (c Constraint) Vars() []TypeVar {
	return c.right.Vars()
}

// Substitutable represents any type expression containing type variables
type Substitutable interface {
	TypeScheme
	// Var returns a list of all vars in the substitutable expression
	Vars() []TypeVar
	// Substitute returns a new substitutable with the constraint applied
	Substitute(Constraint) Substitutable
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
func (a arrayTypeScheme) typeScheme() {}

func (a arrayTypeScheme) Substitute(c Constraint) Substitutable {
	a.elementType = a.elementType.Substitute(c)
	return a
}

func (a arrayTypeScheme) Vars() []TypeVar {
	return a.elementType.Vars()
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
func (o objectTypeScheme) typeScheme() {}

func (o objectTypeScheme) Substitute(c Constraint) Substitutable {
	no := objectTypeScheme{
		properties: make(map[string]Substitutable, len(o.properties)),
	}
	for k, p := range o.properties {
		no.properties[k] = p.Substitute(c)
	}
	return no
}

func (o objectTypeScheme) Vars() []TypeVar {
	var vars []TypeVar
	for _, p := range o.properties {
		vars = append(vars, p.Vars()...)
	}
	return vars
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
func (f functionTypeScheme) typeScheme() {}

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

func (f functionTypeScheme) Vars() []TypeVar {
	vars := f.returnType.Vars()
	for _, p := range f.params {
		vars = append(vars, p.Vars()...)
	}
	return vars
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
	cons  *[]Constraint
	scope *IdentifierScope

	functionExpr *FunctionExpression
}

func NewConstraintGenerationVisitor(tenv map[Node]TypeVar) *ConstraintGenerationVisitor {
	return &ConstraintGenerationVisitor{
		tenv:  tenv,
		cons:  new([]Constraint),
		scope: NewIdentifierScope(),
	}
}

func (v *ConstraintGenerationVisitor) nestScope() *ConstraintGenerationVisitor {
	return &ConstraintGenerationVisitor{
		tenv:         v.tenv,
		cons:         v.cons,
		scope:        v.scope.Nest(),
		functionExpr: v.functionExpr,
	}
}

func (v *ConstraintGenerationVisitor) nest() *ConstraintGenerationVisitor {
	return &ConstraintGenerationVisitor{
		tenv:         v.tenv,
		cons:         v.cons,
		scope:        v.scope,
		functionExpr: v.functionExpr,
	}
}

func (v *ConstraintGenerationVisitor) addConstraints(cs ...Constraint) {
	*v.cons = append(*v.cons, cs...)
}

func (v *ConstraintGenerationVisitor) Constraints() ConstraintSet {
	return *v.cons
}

func (v *ConstraintGenerationVisitor) TypeEnvironment() map[Node]TypeVar {
	return v.tenv
}

// Visit a semantic node and generate a type constraint
func (v *ConstraintGenerationVisitor) Visit(node Node) Visitor {
	tv := v.tenv[node]
	switch n := node.(type) {
	case *BlockStatement:
		return v.nestScope()
	case *NativeVariableDeclaration:
		// Check scope for previously declared type in local scope
		prev, ok := v.scope.LocalLookup(n.Identifier.Name)
		if ok {
			v.addConstraints(Constraint{
				left:  v.tenv[prev],
				right: v.tenv[n],
			})
		}
		// Update scope with the declaration
		v.scope.Set(n.Identifier.Name, n)

		// Inference Rule: Variable Declaration
		// ------------------------------------
		// x = expression
		//
		// -> typeof(x) = typeof(expression)
		// ------------------------------------
		v.addConstraints(Constraint{
			left:  v.tenv[n],
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

		returnTypeVar := v.tenv[n.Block]
		funcType := functionTypeScheme{
			params:     make(map[string]Substitutable, len(n.Block.Parameters.List)),
			returnType: returnTypeVar,
		}
		for _, param := range n.Block.Parameters.List {
			funcType.params[param.Key.Name] = v.tenv[param]
		}
		v.addConstraints(Constraint{
			left:  tv,
			right: funcType,
		})

		// Remember the context of the function defaults
		nv := v.nest()
		nv.functionExpr = n
		return nv
	case *FunctionBlock:
		// TODO(nathanielc): Handle case were Argument is not annotated because it is a node
		argumentVar := v.tenv[n.Body]
		v.addConstraints(Constraint{
			left:  tv,
			right: argumentVar,
		})
		// new function scope
		return v.nestScope()
	case *FunctionParameter:
		// maintain scope
		v.scope.Set(n.Key.Name, n)
		log.Printf("FunctionParam %p %q", v.scope, n.Key.Name)

		// Find default parameter
		def, ok := v.lookupDefaultParameter(n)
		if ok {
			v.addConstraints(Constraint{
				left:  v.tenv[n],
				right: v.tenv[def],
			})
		}
	case *CallExpression:
		// Find FunctionBody and add constraint that typeof(body) == tv
		fe, err := v.lookupFunctionExpression(n.Callee)
		if err != nil {
			panic(err)
		}
		funcBodyTypeVar := v.tenv[fe.Block]
		v.addConstraints(Constraint{
			left:  tv,
			right: funcBodyTypeVar,
		})
		// Add constraints for call arguments
		for _, a := range n.Arguments.Properties {
			for _, p := range fe.Block.Parameters.List {
				if p.Key.Name == a.Key.Name {
					v.addConstraints(Constraint{
						left:  v.tenv[a.Value],
						right: v.tenv[p],
					})
					break
				}
			}
		}
	case *UnaryExpression:
		// Inference Rule: Unary Expression
		// --------------------------------
		// !(expression) <=> NOT (Boolean)
		// -(expression) <=> MINUS (Int)
		// --------------------------------
		switch n.Operator {
		case ast.NotOperator:
			v.addConstraints(
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
			v.addConstraints(
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
		v.addConstraints(
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
			v.addConstraints(
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
			v.addConstraints(Constraint{
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
			v.addConstraints(
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
		if len(n.Elements) > 0 {
			v.addConstraints(Constraint{
				left: v.tenv[n],
				right: arrayTypeScheme{
					elementType: v.tenv[n.Elements[0]],
				},
			})
			for _, e := range n.Elements {
				v.addConstraints(Constraint{
					left:  v.tenv[n.Elements[0]],
					right: v.tenv[e],
				})
			}
		} else {
			v.addConstraints(Constraint{
				left:  v.tenv[n],
				right: EmptyArrayType,
			})
		}
	case *ObjectExpression:
		// Object expressions generate trivial constraints
		v.addConstraints(Constraint{
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
		node, found := v.scope.Lookup(n.Name)
		if !found {
			log.Println(v.scope)
			panic(fmt.Sprintf("missing identifier %q", n.Name))
		}
		tvar := v.tenv[node]
		v.addConstraints(Constraint{
			left:  tvar,
			right: tv,
		})
	case *BooleanLiteral:
		v.addConstraints(Constraint{
			left:  tv,
			right: Bool,
		})
	case *DateTimeLiteral:
		v.addConstraints(Constraint{
			left:  tv,
			right: Time,
		})
	case *DurationLiteral:
		v.addConstraints(Constraint{
			left:  tv,
			right: Duration,
		})
	case *FloatLiteral:
		v.addConstraints(Constraint{
			left:  tv,
			right: Float,
		})
	case *IntegerLiteral:
		v.addConstraints(Constraint{
			left:  tv,
			right: Int,
		})
	case *RegexpLiteral:
		v.addConstraints(Constraint{
			left:  tv,
			right: Regexp,
		})
	case *StringLiteral:
		v.addConstraints(Constraint{
			left:  tv,
			right: String,
		})
	case *UnsignedIntegerLiteral:
		v.addConstraints(Constraint{
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
		node, found := v.scope.Lookup(n.Name)
		if !found {
			return nil, fmt.Errorf("unknown identifier %q", n.Name)
		}
		decl, ok := node.(*NativeVariableDeclaration)
		if !ok {
			return nil, fmt.Errorf("impossible, identifier does not resolve to a native declaration %T", node)
		}
		fe, ok := decl.Init.(*FunctionExpression)
		if !ok {
			return nil, fmt.Errorf("cannot call non-function %q, got type %T", n.Name, decl.Init)
		}
		return fe, nil
	default:
		return nil, fmt.Errorf("unsupported callee type %T", callee)
	}
}

func (v *ConstraintGenerationVisitor) lookupDefaultParameter(p *FunctionParameter) (*FunctionParameterDefault, bool) {
	if v.functionExpr != nil && v.functionExpr.Defaults != nil {
		for _, dp := range v.functionExpr.Defaults.List {
			if dp.Key.Name == p.Key.Name {
				return dp, true
			}
		}
	}
	return nil, false
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

// Lookup returns the node associated with name in the scope.
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

//LocalLookup returns the node associated with the name in the local scope and does not inspect parent scopes.
func (s *IdentifierScope) LocalLookup(name string) (Node, bool) {
	if s == nil {
		return nil, false
	}
	dec, ok := s.vars[name]
	return dec, ok
}

// Nest returns a new variable scope whose parent is the current scope
func (s *IdentifierScope) Nest() *IdentifierScope {
	return &IdentifierScope{
		parent: s,
		vars:   make(map[string]Node, 8),
	}
}

type ConstraintSet []Constraint

func (cs ConstraintSet) String() string {
	var builder strings.Builder
	builder.WriteString("[")
	if len(cs) > 1 {
		builder.WriteString("\n")
	}
	for i, c := range cs {
		if i != 0 {
			builder.WriteString(",\n")
		}
		fmt.Fprintf(&builder, "%v", c)
	}
	builder.WriteString("]")
	return builder.String()
}

func GenerateConstraints(n Node, tenv map[Node]TypeVar) ConstraintSet {
	// Generate the rest of the constraints
	constraintVisitor := NewConstraintGenerationVisitor(tenv)
	Walk(constraintVisitor, n)

	return constraintVisitor.Constraints()
}
