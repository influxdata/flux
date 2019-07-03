package promflux

import (
	"fmt"
	"regexp"

	"github.com/influxdata/flux/ast"
	"github.com/prometheus/prometheus/promql"
)

var arithBinOps = map[promql.ItemType]ast.OperatorKind{
	promql.ItemADD: ast.AdditionOperator,
	promql.ItemSUB: ast.SubtractionOperator,
	promql.ItemMUL: ast.MultiplicationOperator,
	promql.ItemDIV: ast.DivisionOperator,
}

var arithBinOpFns = map[promql.ItemType]string{
	promql.ItemPOW: "math.pow",
	promql.ItemMOD: "math.mod",
}

var compBinOps = map[promql.ItemType]ast.OperatorKind{
	promql.ItemEQL: ast.EqualOperator,
	promql.ItemNEQ: ast.NotEqualOperator,
	promql.ItemGTR: ast.GreaterThanOperator,
	promql.ItemLSS: ast.LessThanOperator,
	promql.ItemGTE: ast.GreaterThanEqualOperator,
	promql.ItemLTE: ast.LessThanEqualOperator,
}

// Function to apply an arithmetic binary operator to all values in a table and a given float64 operand.
func scalarArithBinaryOpFn(op ast.OperatorKind, operand ast.Expression, swapped bool) *ast.FunctionExpression {
	val := member("r", "_value")

	var lhs, rhs ast.Expression = val, operand

	if swapped {
		lhs, rhs = rhs, lhs
	}

	// (r) => {r with _value: <lhs> <op> <rhs>, _stop: r._stop}
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
					Key: &ast.Identifier{Name: "_value"},
					Value: &ast.BinaryExpression{
						Operator: op,
						Left:     lhs,
						Right:    rhs,
					},
				},
				{
					Key:   &ast.Identifier{Name: "_stop"},
					Value: member("r", "_stop"),
				},
			},
		},
	}
}

// Function to apply a comparison binary operator to all values in a table and a given float64 operand.
func scalarCompBinaryOpFn(op ast.OperatorKind, operand ast.Expression, swapped bool) *ast.FunctionExpression {
	val := member("r", "_value")

	var lhs, rhs ast.Expression = val, operand

	if swapped {
		lhs, rhs = rhs, lhs
	}

	// (r) => <lhs> <op> <rhs>
	return &ast.FunctionExpression{
		Params: []*ast.Property{
			{
				Key: &ast.Identifier{
					Name: "r",
				},
			},
		},
		Body: &ast.BinaryExpression{
			Operator: op,
			Left:     lhs,
			Right:    rhs,
		},
	}
}

// Function to apply a binary arithmetic operator between values of two joined tables.
func vectorArithBinaryOpFn(op ast.OperatorKind) *ast.FunctionExpression {
	lhs := member("r", "_value_lhs")
	rhs := member("r", "_value_rhs")

	// (r) => {r with _value: <lhs> <op> <rhs>, _stop: r._stop}
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
					Key: &ast.Identifier{Name: "_value"},
					Value: &ast.BinaryExpression{
						Operator: op,
						Left:     lhs,
						Right:    rhs,
					},
				},
				{
					Key:   &ast.Identifier{Name: "_stop"},
					Value: member("r", "_stop"),
				},
			},
		},
	}
}

// Function to apply a binary arithmetic operator math function between values of two joined tables.
func vectorArithBinaryMathFn(mathFn string) *ast.FunctionExpression {
	lhs := member("r", "_value_lhs")
	rhs := member("r", "_value_rhs")

	// (r) => {r with _value: mathFn(<lhs>, <rhs>), _stop: r._stop}
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

// Function to apply a binary comparison operator between values of two joined tables.
func vectorCompBinaryOpFn(op ast.OperatorKind) *ast.FunctionExpression {
	// (r) => <lhs> <op> <rhs>
	return &ast.FunctionExpression{
		Params: []*ast.Property{
			{
				Key: &ast.Identifier{
					Name: "r",
				},
			},
		},
		Body: &ast.BinaryExpression{
			Operator: op,
			Left:     member("r", "_value_lhs"),
			Right:    member("r", "_value_rhs"),
		},
	}
}

// Function to rename post-join LHS/RHS columns.
func stripSuffixFn(suffix string) *ast.FunctionExpression {
	return &ast.FunctionExpression{
		Params: []*ast.Property{
			{
				Key: &ast.Identifier{
					Name: "column",
				},
			},
		},
		Body: call("strings.trimSuffix", map[string]ast.Expression{
			"v":      &ast.Identifier{Name: "column"},
			"suffix": &ast.StringLiteral{Value: suffix},
		}),
	}
}

// Function to match post-join RHS columns.
var matchRHSSuffixFn = &ast.FunctionExpression{
	Params: []*ast.Property{
		{
			Key: &ast.Identifier{
				Name: "column",
			},
		},
	},
	Body: &ast.BinaryExpression{
		Operator: ast.RegexpMatchOperator,
		Left:     &ast.Identifier{Name: "column"},
		Right:    &ast.RegexpLiteral{Value: regexp.MustCompile("_rhs$")},
	},
}

func (t *Transpiler) transpileBinaryExpr(b *promql.BinaryExpr) (ast.Expression, error) {
	lhs, err := t.transpileExpr(b.LHS)
	if err != nil {
		return nil, fmt.Errorf("unable to transpile left-hand side of binary operation: %s", err)
	}
	rhs, err := t.transpileExpr(b.RHS)
	if err != nil {
		return nil, fmt.Errorf("unable to transpile right-hand side of binary operation: %s", err)
	}

	swapped := false

	switch {
	case yieldsFloat(b.LHS) && yieldsFloat(b.RHS):
		if op, ok := arithBinOps[b.Op]; ok {
			return &ast.BinaryExpression{
				Operator: op,
				Left:     lhs,
				Right:    rhs,
			}, nil
		}

		if opFn, ok := arithBinOpFns[b.Op]; ok {
			return call(opFn, map[string]ast.Expression{"x": lhs, "y": rhs}), nil
		}

		if op, ok := compBinOps[b.Op]; ok {
			if !b.ReturnBool {
				// This is already caught by the PromQL parser.
				return nil, fmt.Errorf("scalar-to-scalar binary op is missing 'bool' modifier (this should never happen)")
			}

			return call("float", map[string]ast.Expression{
				"v": &ast.BinaryExpression{
					Operator: op,
					Left:     lhs,
					Right:    rhs,
				},
			}), nil
		}

		return nil, fmt.Errorf("invalid scalar-scalar binary op %q (this should never happen)", b.Op)
	case yieldsFloat(b.LHS) && yieldsTable(b.RHS):
		lhs, rhs = rhs, lhs
		swapped = true
		fallthrough
	case yieldsTable(b.LHS) && yieldsFloat(b.RHS):
		if op, ok := arithBinOps[b.Op]; ok {
			return buildPipeline(
				lhs,
				call("map", map[string]ast.Expression{"fn": scalarArithBinaryOpFn(op, rhs, swapped)}),
				dropFieldAndTimeCall,
			), nil
		}

		if opFn, ok := arithBinOpFns[b.Op]; ok {
			return buildPipeline(
				lhs,
				call("map", map[string]ast.Expression{"fn": scalarArithBinaryMathFn(opFn, rhs, swapped)}),
				dropFieldAndTimeCall,
			), nil
		}

		if op, ok := compBinOps[b.Op]; ok {
			if b.ReturnBool {
				return buildPipeline(
					lhs,
					call("map", map[string]ast.Expression{
						"fn": scalarArithBinaryOpFn(op, rhs, swapped),
					}),
					call("toFloat", nil),
					dropFieldAndTimeCall,
				), nil
			}
			return buildPipeline(
				lhs,
				call("filter", map[string]ast.Expression{"fn": scalarCompBinaryOpFn(op, rhs, swapped)}),
			), nil
		}

		return nil, fmt.Errorf("invalid scalar-vector binary op %q (this should never happen)", b.Op)
	default:
		if b.VectorMatching == nil {
			// We end up in this branch for non-const scalar-typed PromQL nodes,
			// which don't have VectorMatching initialized.
			b.VectorMatching = &promql.VectorMatching{
				On: true,
			}
		} else if !b.VectorMatching.On || len(b.VectorMatching.MatchingLabels) == 0 {
			return nil, fmt.Errorf("vector-to-vector binary expressions without on() clause not supported yet")
		}

		dropField := true
		var opCalls []*ast.CallExpression

		if op, ok := arithBinOps[b.Op]; ok {
			opCalls = []*ast.CallExpression{
				call("map", map[string]ast.Expression{"fn": vectorArithBinaryOpFn(op)}),
			}
		} else if opFn, ok := arithBinOpFns[b.Op]; ok {
			opCalls = []*ast.CallExpression{
				call("map", map[string]ast.Expression{"fn": vectorArithBinaryMathFn(opFn)}),
			}
		} else if op, ok := compBinOps[b.Op]; ok {
			if b.ReturnBool {
				opCalls = []*ast.CallExpression{
					call("map", map[string]ast.Expression{"fn": vectorArithBinaryOpFn(op)}),
					call("toFloat", nil),
				}
			} else {
				opCalls = []*ast.CallExpression{
					call("filter", map[string]ast.Expression{"fn": vectorCompBinaryOpFn(op)}),
				}
				if b.LHS.Type() == promql.ValueTypeScalar {
					// For <scalar> <comp-op> <vector> filter expressions, we always want to
					// return the sample value from the vector, not the scalar.
					opCalls = append(
						opCalls,
						call("duplicate", map[string]ast.Expression{
							"column": &ast.StringLiteral{Value: "_value_rhs"},
							"as":     &ast.StringLiteral{Value: "_value_lhs"},
						}),
					)
				}
				dropField = false
			}
		} else {
			return nil, fmt.Errorf("vector set operations not supported yet")
		}

		onCols := append(b.VectorMatching.MatchingLabels, "_start", "_stop")

		outputColTransformCalls := []*ast.CallExpression{
			call("keep", map[string]ast.Expression{
				"columns": columnList(append(append(onCols, "_value"), b.VectorMatching.Include...)...),
			}),

			// TODO: Fix binary operations once new join implementation exists.
			//
			// // Rename x_lhs -> x.
			// call("rename", map[string]ast.Expression{"fn": stripSuffixFn("_lhs")}),
			// // Drop cols RHS cols, except ones we want to copy into the result via a group_x(...) clause.
			// call("drop", map[string]ast.Expression{"columns": columnList(b.VectorMatching.Include...)}),
			// // Rename x_rhs -> x.
			// call("rename", map[string]ast.Expression{"fn": stripSuffixFn("_rhs")}),
			// // Drop any remaining RHS cols.
			// call("drop", map[string]ast.Expression{"fn": matchRHSSuffixFn}),
		}

		postJoinCalls := append(opCalls, outputColTransformCalls...)
		if dropField {
			postJoinCalls = append(postJoinCalls, dropFieldAndTimeCall)
		}

		return buildPipeline(
			call("join", map[string]ast.Expression{
				"tables": &ast.ObjectExpression{
					Properties: []*ast.Property{
						{
							Key:   &ast.Identifier{Name: "lhs"},
							Value: lhs,
						},
						{
							Key:   &ast.Identifier{Name: "rhs"},
							Value: rhs,
						},
					},
				},
				"on": columnList(onCols...),
			}),
			postJoinCalls...,
		), nil
	}
}
