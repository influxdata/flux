package astutil

import "github.com/influxdata/flux/ast"

// A Visitor's Visit method is invoked for each node encountered by Walk.
// If the result visitor w is not nil, Walk visits each of the children
// of node with the visitor w, followed by a call of w.Done(n).
// This will not visit nil nodes.
type Visitor interface {
	Visit(ast.Node) Visitor
	Done(ast.Node)
}

func Walk(v Visitor, n ast.Node) {
	switch n := n.(type) {
	case *ast.ArrayExpression:
		if n == nil {
			return
		}
		if w := v.Visit(n); w != nil {
			for _, element := range n.Elements {
				w.Visit(element)
			}
		}

	case *ast.ArrowFunctionExpression:
		if n == nil {
			return
		}
		if w := v.Visit(n); w != nil {
			for _, param := range n.Params {
				w.Visit(param)
			}
			w.Visit(n.Body)
		}

	case *ast.BinaryExpression:
		if n == nil {
			return
		}
		if w := v.Visit(n); w != nil {
			w.Visit(n.Left)
			w.Visit(n.Right)
		}

	case *ast.BlockStatement:
		if n == nil {
			return
		}
		if w := v.Visit(n); w != nil {
			for _, stmt := range n.Body {
				w.Visit(stmt)
			}
		}

	case *ast.BooleanLiteral:
		if n == nil {
			return
		}
		_ = v.Visit(n)

	case *ast.CallExpression:
		if n == nil {
			return
		}
		if w := v.Visit(n); w != nil {
			w.Visit(n.Callee)
			for _, arg := range n.Arguments {
				w.Visit(arg)
			}
		}

	case *ast.ConditionalExpression:
		// todo(jsternberg): this is probably the wrong order and this
		// expression isn't defined in the parser grammar.
		if n == nil {
			return
		}
		if w := v.Visit(n); w != nil {
			w.Visit(n.Test)
			w.Visit(n.Consequent)
			w.Visit(n.Alternate)
		}

	case *ast.DateTimeLiteral:
		if n == nil {
			return
		}
		_ = v.Visit(n)

	case *ast.DurationLiteral:
		if n == nil {
			return
		}
		_ = v.Visit(n)

	case *ast.ExpressionStatement:
		if n == nil {
			return
		}
		if w := v.Visit(n); w != nil {
			w.Visit(n.Expression)
		}

	case *ast.FloatLiteral:
		if n == nil {
			return
		}
		_ = v.Visit(n)

	case *ast.Identifier:
		if n == nil {
			return
		}
		_ = v.Visit(n)

	case *ast.IndexExpression:
		if n == nil {
			return
		}
		if w := v.Visit(n); w != nil {
			w.Visit(n.Array)
			w.Visit(n.Index)
		}

	case *ast.IntegerLiteral:
		if n == nil {
			return
		}
		_ = v.Visit(n)

	case *ast.LogicalExpression:
		if n == nil {
			return
		}
		if w := v.Visit(n); w != nil {
			w.Visit(n.Left)
			w.Visit(n.Right)
		}

	case *ast.MemberExpression:
		if n == nil {
			return
		}
		if w := v.Visit(n); w != nil {
			w.Visit(n.Object)
			w.Visit(n.Property)
		}

	case *ast.ObjectExpression:
		if n == nil {
			return
		}
		if w := v.Visit(n); w != nil {
			for _, property := range n.Properties {
				w.Visit(property)
			}
		}

	case *ast.OptionStatement:
		if n == nil {
			return
		}
		if w := v.Visit(n); w != nil {
			w.Visit(n.Declaration)
		}

	case *ast.PipeExpression:
		if n == nil {
			return
		}
		if w := v.Visit(n); w != nil {
			w.Visit(n.Argument)
			w.Visit(n.Call)
		}

	case *ast.PipeLiteral:
		if n == nil {
			return
		}
		_ = v.Visit(n)

	case *ast.Program:
		if n == nil {
			return
		}
		if w := v.Visit(n); w != nil {
			for _, stmt := range n.Body {
				w.Visit(stmt)
			}
		}

	case *ast.Property:
		if n == nil {
			return
		}
		if w := v.Visit(n); w != nil {
			w.Visit(n.Key)
			w.Visit(n.Value)
		}

	case *ast.RegexpLiteral:
		if n == nil {
			return
		}
		_ = v.Visit(n)

	case *ast.ReturnStatement:
		if n == nil {
			return
		}
		if w := v.Visit(n); w != nil {
			w.Visit(n.Argument)
		}

	case *ast.StringLiteral:
		if n == nil {
			return
		}
		_ = v.Visit(n)

	case *ast.UnaryExpression:
		if n == nil {
			return
		}
		if w := v.Visit(n); w != nil {
			w.Visit(n.Argument)
		}

	case *ast.UnsignedIntegerLiteral:
		if n == nil {
			return
		}
		_ = v.Visit(n)

	case *ast.VariableDeclaration:
		if n == nil {
			return
		}
		if w := v.Visit(n); w != nil {
			for _, decl := range n.Declarations {
				w.Visit(decl)
			}
		}

	case *ast.VariableDeclarator:
		if n == nil {
			return
		}
		if w := v.Visit(n); w != nil {
			w.Visit(n.ID)
			w.Visit(n.Init)
		}
	}

	v.Done(n)
}
