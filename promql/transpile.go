package promql

import (
	"fmt"
	"strings"
	"time"

	"github.com/influxdata/flux/ast"
	"github.com/prometheus/prometheus/promql"
)

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
		lastCall := calls[len(calls)-1]
		// Skip nil call arguments.
		//
		// This keeps the code simpler when constructing pipelines where some function calls
		// are optional and thus can remain nil (using a variable that sets them to an actual
		// call if needed), but thus can still be included in the final pipeline assembly code.
		if lastCall == nil {
			return buildPipeline(arg, calls[0:len(calls)-1]...)
		}

		return &ast.PipeExpression{
			Argument: buildPipeline(arg, calls[0:len(calls)-1]...),
			Call:     lastCall,
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

	// (r) => {r with _value: mathFn("x": <lhs>, "y": <rhs>), _stop: r._stop}
	return &ast.FunctionExpression{
		Params: []*ast.Property{
			{
				Key: &ast.Identifier{
					Name: "r",
				},
			},
		},
		Body: &ast.ObjectExpression{
			With: &ast.Identifier{Name: "r"},
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

var dropFieldAndTimeCall = call(
	"drop",
	map[string]ast.Expression{
		"columns": &ast.ArrayExpression{
			Elements: []ast.Expression{
				&ast.StringLiteral{Value: "_field"},
				&ast.StringLiteral{Value: "_time"},
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

// A Transpiler allows transpiling a PromQL expression into a Flux file
// according to a chosen evaluation time range.
type Transpiler struct {
	Bucket     string
	Start      time.Time
	End        time.Time
	Resolution time.Duration
}

// Transpile converts a PromQL expression with the time ranges set in the transpiler
// into a Flux file. The resulting Flux file can be executed and the result needs to
// be transformed using FluxResultToPromQLValue() to get a result value that is fully
// equivalent to the result of a native PromQL execution.
//
// During the transpilation, the transpiler recurisvely translates the PromQL AST into
// equivalent Flux nodes. Each PromQL node translates into one or more Flux
// constructs that as a group (corresponding to the PromQL node) have to
// keep the following invariants:
//
// - The "_field" column contains the PromQL metric name, if any.
// - The "_measurement" column is ignored (always set to constant "prometheus").
// - The "_time" column contains the sample timestamp as long as a raw sample has been
//   selected from storage and not processed further. Otherwise, "_time" will be
//   empty.
// - The "_stop" column contains the stop timestamp of windows that are equivalent to
//   the resolution steps in PromQL. If "_time" is no longer present, "_stop" becomes
//   the output timestamp for a sample.
// - The "_value" column is always of float type and represents the PromQL sample value.
// - Other columns map to PromQL label names, with escaping applied ("_foo" -> "~_foo").
// - Tables should be grouped by all columns except for "_time" and "_value". Each Flux
//   table represents one PromQL series, with potentially multiple samples over time.
func (t *Transpiler) Transpile(expr promql.Expr) (*ast.File, error) {
	promql.Walk(labelNameEscaper{}, expr, nil)

	fluxNode, err := t.transpileExpr(expr)
	if err != nil {
		return nil, fmt.Errorf("error transpiling expression: %s", err)
	}

	// Scalar constants need to be converted to vectors in the final result.
	if yieldsFloat(expr) {
		fluxNode = buildPipeline(
			t.generateZeroWindows(),
			call("map", map[string]ast.Expression{
				"fn": setConstValueFn(fluxNode),
			}),
		)
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

func (t *Transpiler) transpileExpr(expr promql.Expr) (ast.Expression, error) {
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
