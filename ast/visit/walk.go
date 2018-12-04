package visit

import (
	"fmt"

	"github.com/influxdata/flux/ast"
)

/*
`Walk` recursively visits every children of a given `ast.Node` given a `Visitor`.
It performs a pre-order visit of the AST (visit parent node, then visit children from left to right).
If a call to `Visit` for a node returns a nil visitor, walk stops and doesn't visit the AST rooted at that node,
otherwise it uses the returned visitor to continue walking.
Once Walk has finished visiting a node (the node itself and its children), it invokes `Done` on the node's visitor.
NOTE: `Walk` doesn't visit `nil` nodes.
*/
func Walk(v Visitor, node ast.Node) {
	walk(v, node)
}

/*
A `Visitor` extracts information from a `ast.Node` to build a result and/or have side-effects on it.
The result of `Visit` is a `Visitor` that, in turn, is used by `Walk` to visit the children of the node under exam.
To stop walking, `Visit` must return `nil`.
*/
type Visitor interface {
	Visit(node ast.Node) Visitor
	Done(node ast.Node)
}

func CreateVisitor(f func(ast.Node)) Visitor {
	return &visitor{f: f}
}

type visitor struct {
	f func(ast.Node)
}

func (v *visitor) Visit(node ast.Node) Visitor {
	v.f(node)
	return v
}

func (v *visitor) Done(node ast.Node) {}

func walk(v Visitor, n ast.Node) {
	if n == nil {
		return
	}

	switch n := n.(type) {
	case *ast.Program:
		w := v.Visit(n)
		if w != nil {
			for _, s := range n.Body {
				walk(w, s)
			}
		}
	case *ast.BlockStatement:
		w := v.Visit(n)
		if w != nil {
			for _, s := range n.Body {
				walk(w, s)
			}
		}
	case *ast.OptionStatement:
		w := v.Visit(n)
		if w != nil && n.Declaration != nil {
			walk(w, n.Declaration)
		}
	case *ast.ExpressionStatement:
		w := v.Visit(n)
		if w != nil {
			walk(w, n.Expression)
		}
	case *ast.ReturnStatement:
		w := v.Visit(n)
		if w != nil {
			walk(w, n.Argument)
		}
	case *ast.VariableDeclaration:
		w := v.Visit(n)
		if w != nil {
			for _, s := range n.Declarations {
				walk(w, s)
			}
		}
	case *ast.VariableDeclarator:
		w := v.Visit(n)
		if w != nil {
			walk(w, n.ID)
			walk(w, n.Init)
		}
	case *ast.CallExpression:
		w := v.Visit(n)
		if w != nil {
			walk(w, n.Callee)
			for _, s := range n.Arguments {
				walk(w, s)
			}
		}
	case *ast.PipeExpression:
		w := v.Visit(n)
		if w != nil {
			walk(w, n.Argument)
			walk(w, n.Call)
		}
	case *ast.MemberExpression:
		w := v.Visit(n)
		if w != nil {
			walk(w, n.Object)
			walk(w, n.Property)
		}
	case *ast.IndexExpression:
		w := v.Visit(n)
		if w != nil {
			walk(w, n.Array)
			walk(w, n.Index)
		}
	case *ast.BinaryExpression:
		w := v.Visit(n)
		if w != nil {
			walk(w, n.Left)
			walk(w, n.Right)
		}
	case *ast.UnaryExpression:
		w := v.Visit(n)
		if w != nil {
			walk(w, n.Argument)
		}
	case *ast.LogicalExpression:
		w := v.Visit(n)
		if w != nil {
			walk(w, n.Left)
			walk(w, n.Right)
		}
	case *ast.ObjectExpression:
		w := v.Visit(n)
		if w != nil {
			for _, p := range n.Properties {
				walk(w, p)
			}
		}
	case *ast.ConditionalExpression:
		w := v.Visit(n)
		if w != nil {
			walk(w, n.Test)
			walk(w, n.Alternate)
			walk(w, n.Consequent)
		}
	case *ast.ArrayExpression:
		w := v.Visit(n)
		if w != nil {
			for _, e := range n.Elements {
				walk(w, e)
			}
		}
	case *ast.ArrowFunctionExpression:
		w := v.Visit(n)
		if w != nil {
			for _, e := range n.Params {
				walk(w, e)
			}
			walk(w, n.Body)
		}
	case *ast.Property:
		w := v.Visit(n)
		if w != nil {
			walk(w, n.Key)
			walk(w, n.Value)
		}
	case *ast.Identifier:
		v.Visit(n)
	case *ast.PipeLiteral:
		v.Visit(n)
	case *ast.StringLiteral:
		v.Visit(n)
	case *ast.BooleanLiteral:
		v.Visit(n)
	case *ast.FloatLiteral:
		v.Visit(n)
	case *ast.IntegerLiteral:
		v.Visit(n)
	case *ast.UnsignedIntegerLiteral:
		v.Visit(n)
	case *ast.RegexpLiteral:
		v.Visit(n)
	case *ast.DurationLiteral:
		v.Visit(n)
	case *ast.DateTimeLiteral:
		v.Visit(n)
	default:
		panic(fmt.Errorf("walk not defined for node %T", n))
	}

	v.Done(n)
}
