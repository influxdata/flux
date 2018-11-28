package ast

import "fmt"

/*
`Walk` recursively visits every children of a given `Node` given a `Visitor`.

It performs a pre-order visit of the AST (visit parent node, then visit children from left to right).
If a call to `Visit` for a node returns a nil visitor, walk stops and doesn't visit the AST rooted at that node,
otherwise it uses the returned visitor to continue walking.
Once Walk has finished visiting a node (the node itself and its children), it invokes `Done` on the node's visitor.

NOTE: `Walk` doesn't visit `nil` nodes.
*/
func Walk(v Visitor, node Node) {
	walk(v, node)
}

/*
A `Visitor` extracts information from a `Node` to build a result and/or have side-effects on it.
The result of `Visit` is a `Visitor` that, in turn, is used by `Walk` to visit the children of the node under exam.
To stop walking, `Visit` must return `nil`.
*/
type Visitor interface {
	Visit(node Node) Visitor
	Done(node Node)
}

func CreateVisitor(f func(Node)) Visitor {
	return &visitor{f: f}
}

type visitor struct {
	f func(Node)
}

func (v *visitor) Visit(node Node) Visitor {
	v.f(node)
	return v
}

func (v *visitor) Done(node Node) {}

func walk(v Visitor, n Node) {
	if n == nil {
		return
	}

	switch n := n.(type) {
	case *Program:
		w := v.Visit(n)
		if w != nil {
			for _, s := range n.Body {
				walk(w, s)
			}
		}
	case *BlockStatement:
		w := v.Visit(n)
		if w != nil {
			for _, s := range n.Body {
				walk(w, s)
			}
		}
	case *OptionStatement:
		w := v.Visit(n)
		if w != nil && n.Declaration != nil {
			walk(w, n.Declaration)
		}
	case *ExpressionStatement:
		w := v.Visit(n)
		if w != nil {
			walk(w, n.Expression)
		}
	case *ReturnStatement:
		w := v.Visit(n)
		if w != nil {
			walk(w, n.Argument)
		}
	case *VariableDeclaration:
		w := v.Visit(n)
		if w != nil {
			for _, s := range n.Declarations {
				walk(w, s)
			}
		}
	case *VariableDeclarator:
		w := v.Visit(n)
		if w != nil {
			walk(w, n.ID)
			walk(w, n.Init)
		}
	case *CallExpression:
		w := v.Visit(n)
		if w != nil {
			walk(w, n.Callee)
			for _, s := range n.Arguments {
				walk(w, s)
			}
		}
	case *PipeExpression:
		w := v.Visit(n)
		if w != nil {
			walk(w, n.Argument)
			walk(w, n.Call)
		}
	case *MemberExpression:
		w := v.Visit(n)
		if w != nil {
			walk(w, n.Object)
			walk(w, n.Property)
		}
	case *IndexExpression:
		w := v.Visit(n)
		if w != nil {
			walk(w, n.Array)
			walk(w, n.Index)
		}
	case *BinaryExpression:
		w := v.Visit(n)
		if w != nil {
			walk(w, n.Left)
			walk(w, n.Right)
		}
	case *UnaryExpression:
		w := v.Visit(n)
		if w != nil {
			walk(w, n.Argument)
		}
	case *LogicalExpression:
		w := v.Visit(n)
		if w != nil {
			walk(w, n.Left)
			walk(w, n.Right)
		}
	case *ObjectExpression:
		w := v.Visit(n)
		if w != nil {
			for _, p := range n.Properties {
				walk(w, p)
			}
		}
	case *ConditionalExpression:
		w := v.Visit(n)
		if w != nil {
			walk(w, n.Test)
			walk(w, n.Alternate)
			walk(w, n.Consequent)
		}
	case *ArrayExpression:
		w := v.Visit(n)
		if w != nil {
			for _, e := range n.Elements {
				walk(w, e)
			}
		}
	case *ArrowFunctionExpression:
		w := v.Visit(n)
		if w != nil {
			for _, e := range n.Params {
				walk(w, e)
			}
			walk(w, n.Body)
		}
	case *Property:
		w := v.Visit(n)
		if w != nil {
			walk(w, n.Key)
			walk(w, n.Value)
		}
	case *Identifier:
		v.Visit(n)
	case *PipeLiteral:
		v.Visit(n)
	case *StringLiteral:
		v.Visit(n)
	case *BooleanLiteral:
		v.Visit(n)
	case *FloatLiteral:
		v.Visit(n)
	case *IntegerLiteral:
		v.Visit(n)
	case *UnsignedIntegerLiteral:
		v.Visit(n)
	case *RegexpLiteral:
		v.Visit(n)
	case *DurationLiteral:
		v.Visit(n)
	case *DateTimeLiteral:
		v.Visit(n)
	default:
		panic(fmt.Errorf("walk not defined for parent %T", n))
	}

	v.Done(n)
}
