package main

import (
	"fmt"

	"github.com/influxdata/flux/ast"
	"github.com/prometheus/prometheus/promql"
)

func (t *transpiler) transpileUnaryExpr(ue *promql.UnaryExpr) (ast.Expression, error) {
	expr, err := t.transpileExpr(ue.Expr)
	if err != nil {
		return nil, fmt.Errorf("error transpiling expression in unary expression: %s", err)
	}

	switch ue.Op {
	case promql.ItemADD:
		return expr, nil
	case promql.ItemSUB:
		if yieldsTable(ue.Expr) {
			// Multiply all table _value columns by -1.
			return buildPipeline(
				expr,
				call("map", map[string]ast.Expression{
					"fn": scalarArithBinaryOpFn(ast.MultiplicationOperator, &ast.FloatLiteral{Value: -1}, false)},
				),
				dropMeasurementCall,
			), nil
		}

		// Multiply float expression by -1.
		return &ast.BinaryExpression{
			Operator: ast.MultiplicationOperator,
			Left:     expr,
			Right:    &ast.FloatLiteral{Value: -1},
		}, nil
	default:
		// PromQL fails to parse this, so this should never happen.
		panic("invalid unary expression operator type")
	}
}
