package semantic

import "github.com/influxdata/flux/ast"

func MergeArrowFunction(a, b *FunctionExpression) *FunctionExpression {
	fn := a.Copy().(*FunctionExpression)

	aExp, aOK := a.Body.(Expression)
	bExp, bOK := b.Body.(Expression)

	if aOK && bOK {
		fn.Body = &LogicalExpression{
			Operator: ast.AndOperator,
			Left:     aExp,
			Right:    bExp,
		}
		return fn
	}

	// TODO(nathanielc): This code is unreachable while the current PushDownRule Match function is inplace.

	and := &LogicalExpression{
		Operator: ast.AndOperator,
		Left:     aExp,
		Right:    bExp,
	}

	// Create pass through arguments expression
	passThroughArgs := &ObjectExpression{
		Properties: make([]*Property, len(a.Params)),
	}
	for i, p := range a.Params {
		passThroughArgs.Properties[i] = &Property{
			Key: p.Key,
			//TODO(nathanielc): Construct valid IdentifierExpression with Declaration for the value.
			//Value: p.Key,
		}
	}

	if !aOK {
		// Rewrite left expression as a function call.
		and.Left = &CallExpression{
			Callee:    a.Copy().(*FunctionExpression),
			Arguments: passThroughArgs.Copy().(*ObjectExpression),
		}
	}
	if !bOK {
		// Rewrite right expression as a function call.
		and.Right = &CallExpression{
			Callee:    b.Copy().(*FunctionExpression),
			Arguments: passThroughArgs.Copy().(*ObjectExpression),
		}
	}
	return fn
}