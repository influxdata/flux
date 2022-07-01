package promql

import (
	"fmt"

	"github.com/influxdata/flux/ast"
	"github.com/influxdata/promql/v2"
)

func (t *Transpiler) transpileUnaryExpr(ue *promql.UnaryExpr) (ast.Expression, error) {
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
				dropFieldAndTimeCall,
			), nil
		}

		return &ast.UnaryExpression{
			Operator: ast.SubtractionOperator,
			Argument: expr,
		}, nil
	default:
		// PromQL fails to parse unary operators other than +/-, so this should never happen.
		return nil, fmt.Errorf("invalid unary expression operator type (this should never happen)")
	}
}
