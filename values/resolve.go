package values

import (
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/semantic"
)

// Resolver represents a value that can resolve itself.
// Resolving is the action of capturing the scope at function declaration and
// replacing any identifiers with static values from the scope where possible.
// TODO(nathanielc): Improve implementations of scope to only preserve values
// in the scope that are referrenced.
type Resolver interface {
	Resolve() (semantic.Node, error)
	Scope() Scope
}

func ResolveFunction(scope Scope, f *semantic.FunctionExpression) (semantic.Node, error) {
	n := f.Copy()
	localIdentifiers := make([]string, 0, 10)
	node, err := resolveIdentifiers(scope, f, n, &localIdentifiers)
	if err != nil {
		return nil, err
	}
	return node, nil
}

func resolveIdentifiers(scope Scope, origExpr *semantic.FunctionExpression, n semantic.Node, localIdentifiers *[]string) (semantic.Node, error) {
	switch n := n.(type) {
	case *semantic.MemberExpression:
		node, err := resolveIdentifiers(scope, origExpr, n.Object, localIdentifiers)
		if err != nil {
			return nil, err
		}
		// Substitute member expressions with the object properties
		// they point to if possible.
		//
		// TODO(jlapacik): The following is a complete hack
		// and should be replaced with a proper eval/reduction
		// of the semantic graph. It has been added to aid the
		// planner in pushing down as many predicates to storage
		// as possible.
		//
		// The planner will now be able to push down predicates
		// involving member expressions like so:
		//
		//     r.env == v.env
		//
		if obj, ok := node.(*semantic.ObjectExpression); ok {
			for _, prop := range obj.Properties {
				if prop.Key.Key() == n.Property {
					return resolveIdentifiers(scope, origExpr, prop.Value, localIdentifiers)
				}
			}
		}
		n.Object = node.(semantic.Expression)
	case *semantic.IdentifierExpression:
		if origExpr.Parameters != nil {
			for _, p := range origExpr.Parameters.List {
				if n.Name == p.Key.Name {
					// Identifier is a parameter do not resolve
					return n, nil
				}
			}
		}

		// if we are looking at a reference to a locally defined variable,
		// then we can't resolve it because it hasn't been evaluated yet.
		for _, id := range *localIdentifiers {
			if id == n.Name {
				return n, nil
			}
		}

		v, ok := scope.Lookup(n.Name)
		if ok {
			// Attempt to resolve the value if it is possible to inline.
			node, ok, err := resolveValue(v)
			if !ok {
				return n, nil
			}
			return node, err
		}
		return nil, errors.Newf(codes.Invalid, "name %q does not exist in scope", n.Name)
	case *semantic.Block:
		for i, s := range n.Body {
			node, err := resolveIdentifiers(scope, origExpr, s, localIdentifiers)
			if err != nil {
				return nil, err
			}
			n.Body[i] = node.(semantic.Statement)
		}
	case *semantic.OptionStatement:
		node, err := resolveIdentifiers(scope, origExpr, n.Assignment, localIdentifiers)
		if err != nil {
			return nil, err
		}
		n.Assignment = node.(semantic.Assignment)
	case *semantic.ExpressionStatement:
		node, err := resolveIdentifiers(scope, origExpr, n.Expression, localIdentifiers)
		if err != nil {
			return nil, err
		}
		n.Expression = node.(semantic.Expression)
	case *semantic.ReturnStatement:
		node, err := resolveIdentifiers(scope, origExpr, n.Argument, localIdentifiers)
		if err != nil {
			return nil, err
		}
		n.Argument = node.(semantic.Expression)
	case *semantic.NativeVariableAssignment:
		node, err := resolveIdentifiers(scope, origExpr, n.Init, localIdentifiers)
		if err != nil {
			return nil, err
		}
		*localIdentifiers = append(*localIdentifiers, n.Identifier.Name)
		n.Init = node.(semantic.Expression)
	case *semantic.CallExpression:
		node, err := resolveIdentifiers(scope, origExpr, n.Arguments, localIdentifiers)
		if err != nil {
			return nil, err
		}
		// TODO(adam): lookup the function definition, call the function if it's found in scope.
		n.Arguments = node.(*semantic.ObjectExpression)
	case *semantic.FunctionExpression:
		node, err := resolveIdentifiers(scope, origExpr, n.Block, localIdentifiers)
		if err != nil {
			return nil, err
		}
		n.Block = node.(*semantic.Block)
	case *semantic.BinaryExpression:
		node, err := resolveIdentifiers(scope, origExpr, n.Left, localIdentifiers)
		if err != nil {
			return nil, err
		}
		n.Left = node.(semantic.Expression)

		node, err = resolveIdentifiers(scope, origExpr, n.Right, localIdentifiers)
		if err != nil {
			return nil, err
		}
		n.Right = node.(semantic.Expression)
	case *semantic.UnaryExpression:
		node, err := resolveIdentifiers(scope, origExpr, n.Argument, localIdentifiers)
		if err != nil {
			return nil, err
		}
		n.Argument = node.(semantic.Expression)

	case *semantic.LogicalExpression:
		node, err := resolveIdentifiers(scope, origExpr, n.Left, localIdentifiers)
		if err != nil {
			return nil, err
		}
		n.Left = node.(semantic.Expression)
		node, err = resolveIdentifiers(scope, origExpr, n.Right, localIdentifiers)
		if err != nil {
			return nil, err
		}
		n.Right = node.(semantic.Expression)
	case *semantic.ArrayExpression:
		for i, el := range n.Elements {
			node, err := resolveIdentifiers(scope, origExpr, el, localIdentifiers)
			if err != nil {
				return nil, err
			}
			n.Elements[i] = node.(semantic.Expression)
		}
	case *semantic.IndexExpression:
		node, err := resolveIdentifiers(scope, origExpr, n.Array, localIdentifiers)
		if err != nil {
			return nil, err
		}
		n.Array = node.(semantic.Expression)
		node, err = resolveIdentifiers(scope, origExpr, n.Index, localIdentifiers)
		if err != nil {
			return nil, err
		}
		n.Index = node.(semantic.Expression)
	case *semantic.ObjectExpression:
		for i, p := range n.Properties {
			node, err := resolveIdentifiers(scope, origExpr, p, localIdentifiers)
			if err != nil {
				return nil, err
			}
			n.Properties[i] = node.(*semantic.Property)
		}
	case *semantic.ConditionalExpression:
		node, err := resolveIdentifiers(scope, origExpr, n.Test, localIdentifiers)
		if err != nil {
			return nil, err
		}
		n.Test = node.(semantic.Expression)

		node, err = resolveIdentifiers(scope, origExpr, n.Alternate, localIdentifiers)
		if err != nil {
			return nil, err
		}
		n.Alternate = node.(semantic.Expression)

		node, err = resolveIdentifiers(scope, origExpr, n.Consequent, localIdentifiers)
		if err != nil {
			return nil, err
		}
		n.Consequent = node.(semantic.Expression)
	case *semantic.Property:
		node, err := resolveIdentifiers(scope, origExpr, n.Value, localIdentifiers)
		if err != nil {
			return nil, err
		}
		n.Value = node.(semantic.Expression)
	}
	return n, nil
}

func resolveValue(v Value) (semantic.Node, bool, error) {
	switch k := v.Type().Nature(); k {
	case semantic.String:
		return &semantic.StringLiteral{
			Value: v.Str(),
		}, true, nil
	case semantic.Int:
		return &semantic.IntegerLiteral{
			Value: v.Int(),
		}, true, nil
	case semantic.UInt:
		return &semantic.UnsignedIntegerLiteral{
			Value: v.UInt(),
		}, true, nil
	case semantic.Float:
		return &semantic.FloatLiteral{
			Value: v.Float(),
		}, true, nil
	case semantic.Bool:
		return &semantic.BooleanLiteral{
			Value: v.Bool(),
		}, true, nil
	case semantic.Time:
		return &semantic.DateTimeLiteral{
			Value: v.Time().Time(),
		}, true, nil
	case semantic.Regexp:
		return &semantic.RegexpLiteral{
			Value: v.Regexp(),
		}, true, nil
	case semantic.Duration:
		d := v.Duration()
		var node semantic.Expression = &semantic.DurationLiteral{
			Values: d.AsValues(),
		}
		if d.IsNegative() {
			node = &semantic.UnaryExpression{
				Operator: ast.SubtractionOperator,
				Argument: node,
			}
		}
		return node, true, nil
	case semantic.Function:
		resolver, ok := v.Function().(Resolver)
		if ok {
			node, err := resolver.Resolve()
			return node, true, err
		}
		return nil, false, nil
	case semantic.Array:
		arr := v.Array()
		node := new(semantic.ArrayExpression)
		node.Elements = make([]semantic.Expression, arr.Len())
		var (
			err error
			ok  = true
		)
		arr.Range(func(i int, el Value) {
			if err != nil || !ok {
				return
			}
			var n semantic.Node
			n, ok, err = resolveValue(el)
			if err != nil {
				return
			} else if ok {
				node.Elements[i] = n.(semantic.Expression)
			}
		})
		if err != nil || !ok {
			return nil, false, err
		}
		// Determine the element type by looking at the first
		// element. If there are no elements, then we have an
		// array type with an invalid element type.
		elemType := semantic.MonoType{}
		if len(node.Elements) > 0 {
			elemType = node.Elements[0].TypeOf()
		}
		node.Type = semantic.NewArrayType(elemType)
		return node, true, nil
	case semantic.Object:
		obj := v.Object()
		node := new(semantic.ObjectExpression)
		node.Properties = make([]*semantic.Property, 0, obj.Len())
		var (
			err error
			ok  = true
		)
		obj.Range(func(k string, v Value) {
			if err != nil || !ok {
				return
			}
			var n semantic.Node
			n, ok, err = resolveValue(v)
			if err != nil {
				return
			} else if ok {
				node.Properties = append(node.Properties, &semantic.Property{
					Key:   &semantic.Identifier{Name: k},
					Value: n.(semantic.Expression),
				})
			}
		})
		if err != nil || !ok {
			return nil, false, err
		}
		return node, true, nil
	default:
		return nil, false, errors.Newf(codes.Internal, "cannot resolve value of type %v", k)
	}
}
