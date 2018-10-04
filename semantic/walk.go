package semantic

func Walk(v Visitor, node Node) {
	walk(v, node)
}

type Visitor interface {
	Visit(node Node) Visitor
	Done()
}

func walk(v Visitor, n Node) {
	defer v.Done()
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
		if w != nil {
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
	case *NativeVariableDeclaration:
		if n != nil {
			w := v.Visit(n)
			if w != nil {
				walk(w, n.Identifier)
				walk(w, n.Init)
			}
		}
	case *ExternalVariableDeclaration:
		if n != nil {
			w := v.Visit(n)
			if w != nil {
				walk(w, n.Identifier)
			}
		}
	case *FunctionExpression:
		w := v.Visit(n)
		if w != nil {
			// Walk defaults first as they are evaluated in
			// the enclosing scope, not the function scope.
			walk(w, n.Defaults)
			walk(w, n.Block)
		}
	case *FunctionDefaults:
		if n != nil {
			w := v.Visit(n)
			if w != nil {
				for _, d := range n.List {
					walk(w, d)
				}
			}
		}
	case *FunctionParameterDefault:
		w := v.Visit(n)
		if w != nil {
			walk(w, n.Key)
			walk(w, n.Value)
		}
	case *FunctionBlock:
		w := v.Visit(n)
		if w != nil {
			walk(w, n.Parameters)
			walk(w, n.Body)
		}
	case *FunctionParameters:
		if n != nil {
			w := v.Visit(n)
			if w != nil {
				for _, p := range n.List {
					walk(w, p)
				}
			}
		}
	case *FunctionParameter:
		w := v.Visit(n)
		if w != nil {
			walk(w, n.Key)
		}
	case *ArrayExpression:
		w := v.Visit(n)
		if w != nil {
			for _, e := range n.Elements {
				walk(w, e)
			}
		}
	case *BinaryExpression:
		w := v.Visit(n)
		if w != nil {
			walk(w, n.Left)
			walk(w, n.Right)
		}
	case *CallExpression:
		w := v.Visit(n)
		if w != nil {
			walk(w, n.Callee)
			walk(w, n.Arguments)
			if n.pipe != nil {
				walk(w, n.pipe)
			}
		}
	case *ConditionalExpression:
		w := v.Visit(n)
		if w != nil {
			walk(w, n.Test)
			walk(w, n.Alternate)
			walk(w, n.Consequent)
		}
	case *IdentifierExpression:
		v.Visit(n)
	case *LogicalExpression:
		w := v.Visit(n)
		if w != nil {
			walk(w, n.Left)
			walk(w, n.Right)
		}
	case *MemberExpression:
		w := v.Visit(n)
		if w != nil {
			walk(w, n.Object)
		}
	case *ObjectExpression:
		w := v.Visit(n)
		if w != nil {
			for _, p := range n.Properties {
				walk(w, p)
			}
		}
	case *UnaryExpression:
		w := v.Visit(n)
		if w != nil {
			walk(w, n.Argument)
		}
	case *Identifier:
		v.Visit(n)
	case *Property:
		w := v.Visit(n)
		if w != nil {
			walk(w, n.Key)
			walk(w, n.Value)
		}
	case *BooleanLiteral:
		v.Visit(n)
	case *DateTimeLiteral:
		v.Visit(n)
	case *DurationLiteral:
		v.Visit(n)
	case *FloatLiteral:
		v.Visit(n)
	case *IntegerLiteral:
		v.Visit(n)
	case *RegexpLiteral:
		v.Visit(n)
	case *StringLiteral:
		v.Visit(n)
	case *UnsignedIntegerLiteral:
		v.Visit(n)
	}
}
