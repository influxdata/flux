package semantic

import "github.com/influxdata/flux/ast"

// Constraint represents an equality constraint between type expressions
type Constraint struct {
	left  TypeVar
	right Substitutable
}

// Substitutable represents any type expression containing type variables
type Substitutable interface {
	FreeTypeVar() []TypeVar
}

type arrayTypeExpression struct {
	elementType Substitutable
}

func (a arrayTypeExpression) FreeTypeVar() []TypeVar {
	return a.elementType.FreeTypeVar()
}

type funcTypeExpression struct {
	params     map[string]Substitutable
	returnType Substitutable
}

func (f funcTypeExpression) FreeTypeVar() []TypeVar {
	// TODO: Handle duplicates
	freeTypeVars := make([]TypeVar, 0, 8)
	for _, v := range f.params {
		freeTypeVars = append(freeTypeVars, v.FreeTypeVar()...)
	}
	return append(freeTypeVars, f.returnType.FreeTypeVar()...)
}

// ConstraintGenerationVisitor visits a semantic graph and generates
// constraints between type variables and type expressions.
type ConstraintGenerationVisitor struct {
	tenv map[Node]TypeVar
	cons []Constraint
}

// Visit a semantic node and generate a type constraint
func (c *ConstraintGenerationVisitor) Visit(node Node) Visitor {
	var tv TypeVar
	var ok bool

	if tv, ok = c.tenv[node]; !ok {
		return c
	}

	switch n := node.(type) {
	case *NativeVariableDeclaration:
		// Inference Rule: Variable Declaration
		// ------------------------------------
		// x = expression
		//
		// -> typeof(x) = typeof(expression)
		//
		// TODO: variables can change value but not type.
		// This defines a type constraint.
		c.cons = append(c.cons, Constraint{
			left:  c.tenv[n.Identifier],
			right: c.tenv[n.Init],
		})
	case *FunctionExpression:
		// Inference Rule: Function Expression
		// -----------------------------------
		// f = (a, b) => {
		//     Statement
		//     Statement
		//     Return Statement
		// }
		//
		// -> typeof(f) = (typeof(a), typeof(b)) => typeof(Return Statement)
		funcType := funcTypeExpression{
			params:     make(map[string]Substitutable, len(n.Params)),
			returnType: n.returnTypeVar,
		}
		c.cons = append(c.cons, Constraint{
			left:  tv,
			right: funcType,
		})
	case *CallExpression:
		// Inference Rule: Call Expression
		// -------------------------------
		// f(a:1, b:2)
		//
		// -> typeof(f) = (typeof(a), typeof(b)) => typeof(f(a:1, b:2))
		operator := c.tenv[n.Callee]
		operand := make(map[string]Substitutable, len(n.Arguments.Properties))
		for i, prop := range n.Arguments.Properties {
			operand[prop.Key.Name] = c.tenv[n.Arguments.Properties[i].Value]
		}
		funcType := funcTypeExpression{
			params:     operand,
			returnType: tv,
		}
		c.cons = append(c.cons, Constraint{
			left:  operator,
			right: funcType,
		})
	case *UnaryExpression:
		// Inference Rule: Unary Expression
		// --------------------------------
		// x = (op) expression
		//
		// -> typeof(x) = bool
		// -> typeof(expression) = bool
		//
		// Unary expressions yield boolean types
		c.cons = append(c.cons, Constraint{
			left:  tv,
			right: Bool,
		})
		switch n.Operator {
		// Unary operators must act on boolean types
		case ast.NotOperator:
			c.cons = append(c.cons, Constraint{
				left:  c.tenv[n.Argument],
				right: Bool,
			})
		}
	case *LogicalExpression:
		// Inference Rule: Logical Expression
		// ----------------------------------
		// x = left (op) right
		//
		// -> typeof(x) = bool
		// -> typeof(left) = bool
		// -> typeof(right) = bool
		//
		// Logical expressions yield boolean types
		// Logical operators act on boolean types
		c.cons = append(c.cons,
			Constraint{
				left:  tv,
				right: Bool,
			},
			Constraint{
				left:  c.tenv[n.Left],
				right: Bool,
			},
			Constraint{
				left:  c.tenv[n.Right],
				right: Bool,
			},
		)
	case *BinaryExpression:
		switch n.Operator {
		// Inference Rule: Arithmetic Operators
		// ------------------------------------
		// x = a ( + , - , * , / ) b
		//
		// -> typeof(x) = typeof(a) = typeof(b)
		//
		// TODO: Only Int, UInt, Float, and String types supported
		case ast.AdditionOperator:
			c.cons = append(c.cons,
				Constraint{
					left:  tv,
					right: c.tenv[n.Left],
				},
				Constraint{
					left:  tv,
					right: c.tenv[n.Right],
				},
				Constraint{
					left:  c.tenv[n.Right],
					right: c.tenv[n.Left],
				},
			)
		case ast.SubtractionOperator:
			c.cons = append(c.cons,
				Constraint{
					left:  tv,
					right: c.tenv[n.Left],
				},
				Constraint{
					left:  tv,
					right: c.tenv[n.Right],
				},
				Constraint{
					left:  c.tenv[n.Right],
					right: c.tenv[n.Left],
				},
			)
		case ast.MultiplicationOperator:
			c.cons = append(c.cons,
				Constraint{
					left:  tv,
					right: c.tenv[n.Left],
				},
				Constraint{
					left:  tv,
					right: c.tenv[n.Right],
				},
				Constraint{
					left:  c.tenv[n.Right],
					right: c.tenv[n.Left],
				},
			)
		case ast.DivisionOperator:
			c.cons = append(c.cons,
				Constraint{
					left:  tv,
					right: c.tenv[n.Left],
				},
				Constraint{
					left:  tv,
					right: c.tenv[n.Right],
				},
				Constraint{
					left:  c.tenv[n.Right],
					right: c.tenv[n.Left],
				},
			)
		// Inference Rules: Comparison Operators
		// -------------------------------------
		// x = left ( <, <=, >, >=, ==, != ) right
		//
		// -> typeof(x) = Bool
		case ast.LessThanEqualOperator:
			c.cons = append(c.cons,
				Constraint{
					left:  tv,
					right: Bool,
				})
		case ast.LessThanOperator:
			c.cons = append(c.cons,
				Constraint{
					left:  tv,
					right: Bool,
				})
		case ast.GreaterThanEqualOperator:
			c.cons = append(c.cons,
				Constraint{
					left:  tv,
					right: Bool,
				})
		case ast.GreaterThanOperator:
			c.cons = append(c.cons,
				Constraint{
					left:  tv,
					right: Bool,
				})
		case ast.EqualOperator:
			c.cons = append(c.cons,
				Constraint{
					left:  tv,
					right: Bool,
				})
		case ast.NotEqualOperator:
			c.cons = append(c.cons,
				Constraint{
					left:  tv,
					right: Bool,
				})
		case ast.StartsWithOperator:
		case ast.InOperator:
		case ast.RegexpMatchOperator:
		case ast.NotRegexpMatchOperator:
		}
	case *ArrayExpression:
		c.cons = append(c.cons, Constraint{
			left: c.tenv[n],
			right: arrayTypeExpression{
				elementType: c.tenv[n.Elements[0]],
			},
		})
		for _, e := range n.Elements {
			c.cons = append(c.cons, Constraint{
				left:  c.tenv[n.Elements[0]],
				right: c.tenv[e],
			})
		}
	case *MemberExpression:
		// TODO: This is probably the most difficult type
		// inference rule. How to constrain this type?
	case *ConditionalExpression:
	case *IdentifierExpression:
		tvar := c.tenv[n.declaration.ID()]
		c.cons = append(c.cons, Constraint{
			left:  tv,
			right: tvar,
		})
	case *BooleanLiteral:
		c.cons = append(c.cons, Constraint{
			left:  tv,
			right: Bool,
		})
	case *DateTimeLiteral:
		c.cons = append(c.cons, Constraint{
			left:  tv,
			right: Time,
		})
	case *DurationLiteral:
		c.cons = append(c.cons, Constraint{
			left:  tv,
			right: Duration,
		})
	case *FloatLiteral:
		c.cons = append(c.cons, Constraint{
			left:  tv,
			right: Float,
		})
	case *IntegerLiteral:
		c.cons = append(c.cons, Constraint{
			left:  tv,
			right: Int,
		})
	case *RegexpLiteral:
		c.cons = append(c.cons, Constraint{
			left:  tv,
			right: Regexp,
		})
	case *StringLiteral:
		c.cons = append(c.cons, Constraint{
			left:  tv,
			right: String,
		})
	case *UnsignedIntegerLiteral:
		c.cons = append(c.cons, Constraint{
			left:  tv,
			right: UInt,
		})
	}
	return c
}

// Done is used to satisfy the Visitor interface
func (c *ConstraintGenerationVisitor) Done() {}
