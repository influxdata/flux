package promflux

import (
	"fmt"
	"strings"
	"time"

	"github.com/influxdata/flux/ast"
	"github.com/prometheus/prometheus/promql"
)

type transpiler struct {
	bucket     string
	start      time.Time
	end        time.Time
	resolution time.Duration
}

func buildPipeline(arg ast.Expression, calls ...*ast.CallExpression) *ast.PipeExpression {
	switch len(calls) {
	case 0:
		panic("empty calls list")
	case 1:
		return &ast.PipeExpression{
			Argument: arg,
			Call:     calls[0],
		}
	default:
		return &ast.PipeExpression{
			Argument: buildPipeline(arg, calls[0:len(calls)-1]...),
			Call:     calls[len(calls)-1],
		}
	}
}

func call(fn string, args map[string]ast.Expression) *ast.CallExpression {
	var callee ast.Expression
	switch components := strings.Split(fn, "."); len(components) {
	case 1:
		callee = &ast.Identifier{Name: fn}
	case 2:
		callee = member(components[0], components[1])
	default:
		panic("invalid number of dot-separated components in function name")
	}

	expr := &ast.CallExpression{
		Callee: callee,
	}
	if len(args) > 0 {
		props := make([]*ast.Property, 0, len(args))
		for k, v := range args {
			props = append(props, &ast.Property{
				Key:   &ast.Identifier{Name: k},
				Value: v,
			})
		}

		expr.Arguments = []ast.Expression{
			&ast.ObjectExpression{
				Properties: props,
			},
		}
	}
	return expr
}

// Function to remove extraneous windows at edges of range.
func windowCutoffFn(minStop time.Time, maxStart time.Time) *ast.FunctionExpression {
	return &ast.FunctionExpression{
		Params: []*ast.Property{
			{
				Key: &ast.Identifier{
					Name: "r",
				},
			},
		},
		Body: &ast.LogicalExpression{
			Operator: ast.AndOperator,
			Left: &ast.BinaryExpression{
				Operator: ast.GreaterThanEqualOperator,
				Left:     member("r", "_stop"),
				Right:    &ast.DateTimeLiteral{Value: minStop},
			},
			Right: &ast.BinaryExpression{
				Operator: ast.LessThanEqualOperator,
				Left:     member("r", "_start"),
				Right:    &ast.DateTimeLiteral{Value: maxStart},
			},
		},
	}
}

// Function to apply a math function to all values in a table and a given float64 operand.
func scalarArithBinaryMathFn(mathFn string, operand ast.Expression, swapped bool) *ast.FunctionExpression {
	val := member("r", "_value")

	var lhs, rhs ast.Expression = val, operand

	if swapped {
		lhs, rhs = rhs, lhs
	}

	// (r) => {"_value": mathFn("x": <lhs>, "y": <rhs>), "_stop": r._stop}
	return &ast.FunctionExpression{
		Params: []*ast.Property{
			{
				Key: &ast.Identifier{
					Name: "r",
				},
			},
		},
		Body: &ast.ObjectExpression{
			Properties: []*ast.Property{
				{
					Key:   &ast.Identifier{Name: "_value"},
					Value: call(mathFn, map[string]ast.Expression{"x": lhs, "y": rhs}),
				},
				{
					Key:   &ast.Identifier{Name: "_stop"},
					Value: member("r", "_stop"),
				},
			},
		},
	}
}

func member(o, p string) *ast.MemberExpression {
	return &ast.MemberExpression{
		Object: &ast.Identifier{
			Name: o,
		},
		Property: &ast.Identifier{
			Name: p,
		},
	}
}

func columnList(strs ...string) *ast.ArrayExpression {
	list := make([]ast.Expression, len(strs))
	for i, str := range strs {
		list[i] = &ast.StringLiteral{Value: str}
	}
	return &ast.ArrayExpression{
		Elements: list,
	}
}

var dropMeasurementCall = call(
	"drop",
	map[string]ast.Expression{
		"columns": &ast.ArrayExpression{
			Elements: []ast.Expression{
				&ast.StringLiteral{Value: "_measurement"},
			},
		},
	},
)

func yieldsFloat(expr promql.Expr) bool {
	switch v := expr.(type) {
	case *promql.NumberLiteral:
		return true
	case *promql.BinaryExpr:
		return yieldsFloat(v.LHS) && yieldsFloat(v.RHS)
	case *promql.UnaryExpr:
		return yieldsFloat(v.Expr)
	case *promql.ParenExpr:
		return yieldsFloat(v.Expr)
	default:
		return false
	}
}

func yieldsTable(expr promql.Expr) bool {
	return !yieldsFloat(expr)
}

func (t *transpiler) transpileExpr(expr promql.Expr) (ast.Expression, error) {
	switch e := expr.(type) {
	case *promql.ParenExpr:
		return t.transpileExpr(e.Expr)
	case *promql.UnaryExpr:
		return t.transpileUnaryExpr(e)
	case *promql.NumberLiteral:
		return &ast.FloatLiteral{Value: e.Val}, nil
	case *promql.StringLiteral:
		return &ast.StringLiteral{Value: e.Val}, nil
	case *promql.VectorSelector:
		return t.transpileInstantVectorSelector(e), nil
	case *promql.MatrixSelector:
		return t.transpileRangeVectorSelector(e), nil
	case *promql.AggregateExpr:
		return t.transpileAggregateExpr(e)
	case *promql.BinaryExpr:
		return t.transpileBinaryExpr(e)
	case *promql.Call:
		return t.transpileCall(e)
	case *promql.SubqueryExpr:
		return t.transpileSubqueryExpr(e)
	default:
		return nil, fmt.Errorf("PromQL node type %T is not supported yet", t)
	}
}

func (t *transpiler) transpile(expr promql.Expr) (*ast.File, error) {
	promql.Walk(labelNameEscaper{}, expr, nil)

	fluxNode, err := t.transpileExpr(expr)
	if err != nil {
		return nil, fmt.Errorf("error transpiling expression: %s", err)
	}


	return &ast.File{
		Imports: []*ast.ImportDeclaration{
			{Path: &ast.StringLiteral{Value: "math"}},
			{Path: &ast.StringLiteral{Value: "promql"}},
		},
		Body: []ast.Statement{
			&ast.ExpressionStatement{
				Expression: buildPipeline(
					fluxNode,
					// The resolution step evaluation timestamp needs to become the output timestamp.
					call("duplicate", map[string]ast.Expression{
						"column": &ast.StringLiteral{Value: "_stop"},
						"as":     &ast.StringLiteral{Value: "_time"},
					}),
				),
			},
		},
	}, nil
}
