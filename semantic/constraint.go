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

// Substitutable represents any type expression containing type variables
type Substitutable interface {
	FreeTypeVar() []TypeVar
	String() string
	Equal(Substitutable) bool
}

type arraySignature struct {
	elementType Substitutable
}

func (a arraySignature) FreeTypeVar() []TypeVar {
	return a.elementType.FreeTypeVar()
}
func (a arraySignature) String() string {
	return "[" + a.elementType.String() + "]"
}
func (a arraySignature) Equal(sub Substitutable) bool {
	if arr, ok := sub.(arraySignature); ok {
		return a.elementType.Equal(arr.elementType)
	}
	return false
}

type objectSignature struct {
	properties map[string]Substitutable
}

func (o objectSignature) FreeTypeVar() []TypeVar {
	// TODO: Handle duplicates
	freeTypeVars := make([]TypeVar, 0, 8)
	for _, v := range o.properties {
		freeTypeVars = append(freeTypeVars, v.FreeTypeVar()...)
	}
	return freeTypeVars
}
func (o objectSignature) String() string {
	builder := strings.Builder{}
	builder.WriteString("{ ")
	for k, v := range o.properties {
		param := fmt.Sprintf("%s:%s, ", k, v.String())
		builder.WriteString(param)
	}
	builder.WriteString("}")
	return builder.String()
}
func (o objectSignature) Equal(sub Substitutable) bool {
	obj, ok := sub.(objectSignature)
	if !ok {
		return false
	}
	for k, v := range o.properties {
		typ, ok := obj.properties[k]
		if !ok || !typ.Equal(v) {
			return false
		}
	}
	return true
}

type funcSignature struct {
	params     map[string]Substitutable
	returnType Substitutable
}

func (f funcSignature) FreeTypeVar() []TypeVar {
	// TODO: Handle duplicates
	freeTypeVars := make([]TypeVar, 0, 8)
	for _, v := range f.params {
		freeTypeVars = append(freeTypeVars, v.FreeTypeVar()...)
	}
	return append(freeTypeVars, f.returnType.FreeTypeVar()...)
}
func (f funcSignature) String() string {
	builder := strings.Builder{}
	builder.WriteString("( ")
	for k, v := range f.params {
		param := fmt.Sprintf("%s:%s, ", k, v.String())
		builder.WriteString(param)
	}
	builder.WriteString(")")
	builder.WriteString(" => ")
	builder.WriteString(f.returnType.String())
	return builder.String()
}
func (f funcSignature) Equal(sub Substitutable) bool {
	fun, ok := sub.(funcSignature)
	if !ok {
		return false
	}
	for k, v := range f.params {
		typ, ok := fun.params[k]
		if !ok || !v.Equal(typ) {
			return false
		}
	}
	return f.returnType.Equal(fun.returnType)
}

// DeclarationConstraintVisitor generates type constraints for variable declarations
type DeclarationConstraintVisitor struct {
	current Node
	tenv    map[Node]TypeVar
	cons    []Constraint
	decs    *VariableScope
}

func NewDeclarationConstraintVisitor(tenv map[Node]TypeVar) *DeclarationConstraintVisitor {
	return &DeclarationConstraintVisitor{
		tenv: tenv,
		cons: make([]Constraint, 0, 10),
		decs: NewVariableScope(),
	}
}

func (v *DeclarationConstraintVisitor) Constraints() []Constraint {
	return v.cons
}

// Visit adds variable re-declaration constraints
func (v *DeclarationConstraintVisitor) Visit(node Node) Visitor {
	v.current = node
	switch n := node.(type) {
	case *BlockStatement:
		v.decs = v.decs.Nest()
	case *NativeVariableDeclaration:
		name := n.Identifier.Name
		dec, ok := v.decs.Lookup(name)
		if !ok {
			v.decs.Set(name, n)
			return v
		}
		// Variables can change value but not type during
		// the course of a program. This defines a constraint.
		v.cons = append(v.cons, Constraint{
			left:  v.tenv[n.Identifier],
			right: v.tenv[dec.ID()],
		})
	case *FunctionExpression:
		v.decs = v.decs.Nest()
	case *FunctionParam:
		v.decs.Set(n.Key.Name, n.declaration)
	}
	return v
}

// Done implements the Visitor interface
func (v *DeclarationConstraintVisitor) Done() {
	switch v.current.(type) {
	case *BlockStatement, *FunctionExpression:
		v.decs = v.decs.Parent()
	}
}

// ConstraintGenerationVisitor visits a semantic graph and generates
// constraints between type variables and type expressions.
type ConstraintGenerationVisitor struct {
	tenv map[Node]TypeVar
	cons []Constraint
}

func NewConstraintGenerationVisitor(tenv map[Node]TypeVar, cons []Constraint) *ConstraintGenerationVisitor {
	return &ConstraintGenerationVisitor{
		tenv: tenv,
		cons: cons,
	}
}

// Format pretty prints the set of constraints generated from the visitor
func (v *ConstraintGenerationVisitor) Format() {
	for _, constraint := range v.cons {
		log.Println(constraint.left.name + " = " + constraint.right.String())
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
	case *NativeVariableDeclaration:
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
		funcType := funcSignature{
			params:     make(map[string]Substitutable, len(n.Params)),
			returnType: n.returnTypeVar,
		}
		for _, param := range n.Params {
			funcType.params[param.Key.Name] = v.tenv[param.Key]
		}
		v.cons = append(v.cons, Constraint{
			left:  tv,
			right: funcType,
		})
	case *CallExpression:
		// Inference Rule: Call Expression
		// ------------------------------------------------------------
		// f(a:1, b:2)
		//
		// -> typeof(f) = (typeof(a), typeof(b)) => typeof(f(a:1, b:2))
		// ------------------------------------------------------------
		operator := v.tenv[n.Callee]
		operand := make(map[string]Substitutable, len(n.Arguments.Properties))
		for i, prop := range n.Arguments.Properties {
			operand[prop.Key.Name] = v.tenv[n.Arguments.Properties[i].Value]
		}
		funcType := funcSignature{
			params:     operand,
			returnType: tv,
		}
		v.cons = append(v.cons, Constraint{
			left:  operator,
			right: funcType,
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
			right: arraySignature{
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
			right: objectSignature{
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
		tvar := v.tenv[n.declaration.ID()]
		v.cons = append(v.cons, Constraint{
			left:  tv,
			right: tvar,
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
