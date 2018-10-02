package semantic

import (
	"fmt"

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

//func (c Constraint) Equal(o Constraint) bool {
//	return c.left == o.left && c.right.Equal(o.left)
//}

// Substitutable represents any type expression containing type variables
type Substitutable interface {
	MonoType() (Type, bool)
}

type arraySignature struct {
	elementType Substitutable
}

func (a arraySignature) MonoType() (Type, bool) {
	elementType, mono := a.elementType.MonoType()
	if !mono {
		return nil, false
	}
	return NewArrayType(elementType), true
}

type objectSignature struct {
	properties map[string]Substitutable
}

func (o objectSignature) MonoType() (Type, bool) {
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

type funcSignature struct {
	params     map[string]Substitutable
	returnType Substitutable
}

func (f funcSignature) MonoType() (Type, bool) {
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

// ConstraintGenerationVisitor visits a semantic graph and generates
// constraints between type variables and type expressions.
type ConstraintGenerationVisitor struct {
	tenv map[Node]TypeVar
	cons []Constraint
}

func NewConstraintGenerationVisitor(tenv map[Node]TypeVar) *ConstraintGenerationVisitor {
	return &ConstraintGenerationVisitor{
		tenv: tenv,
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
	case *FunctionParam:
		key := v.tenv[n.Key]
		def, ok := v.tenv[n.Default]
		if ok {
			v.cons = append(v.cons, Constraint{
				left:  key,
				right: def,
			})
		}
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
		tvar := v.tenv[n.declaration]
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

func GenerateConstraints(program *Program, tenv map[Node]TypeVar) []Constraint {
	dv := NewVariableDeclarationVisitor()
	Walk(dv, program)

	// Generate the rest of the constraints
	constraintVisitor := NewConstraintGenerationVisitor(tenv)
	Walk(constraintVisitor, program)

	return constraintVisitor.Constraints()
}

//type ConstraintSet struct {
//	set []Constraint
//}
//
//func (s *ConstraintSet) Add(c Constraint) {
//	for _, oc := range s.set {
//		if oc.Equal(c) {
//			return
//		}
//	}
//	s.set = append(s.set, c)
//}
