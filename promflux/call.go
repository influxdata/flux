package main

import (
	"fmt"
	"time"

	"github.com/influxdata/flux/ast"
	"github.com/prometheus/prometheus/promql"
)

// TODO: Temporary hack to work around lack of null support in filter(). Remove this.
const nullReplacement = 123456789

var aggregateOverTimeFns = map[string]string{
	"sum_over_time":      "sum",
	"avg_over_time":      "mean",
	"max_over_time":      "max",
	"min_over_time":      "min",
	"count_over_time":    "count",
	"stddev_over_time":   "stddev",
	"stdvar_over_time":   "stddev", // TODO: Add stdvar() to Flux stdlib instead of special-casing this below.
	"quantile_over_time": "quantile",
}

var vectorMathFunctions = map[string]string{
	"abs":   "math.abs",
	"ceil":  "math.ceil",
	"floor": "math.floor",
	"exp":   "math.exp",
	"sqrt":  "math.sqrt",
	"ln":    "math.log",
	"log2":  "math.log2",
	"log10": "math.log10",
	"round": "math.round",
}

// TODO: Super temporary hack to deal with null values. Remove!
var filterSpecialNullValuesCall = call(
	"filter",
	map[string]ast.Expression{
		"fn": &ast.FunctionExpression{
			Params: []*ast.Property{
				{
					Key: &ast.Identifier{
						Name: "r",
					},
				},
			},
			Body: &ast.BinaryExpression{
				Operator: ast.NotEqualOperator,
				Left: &ast.MemberExpression{
					Object: &ast.Identifier{
						Name: "r",
					},
					Property: &ast.Identifier{
						Name: "_value",
					},
				},
				Right: &ast.FloatLiteral{Value: nullReplacement},
			},
		},
	},
)

// Function to apply a simple one-operand function to all values in a table.
func vectorMathFn(fn string) *ast.FunctionExpression {
	// (r) => {"_value": mathFn(x: r._value), "_stop": r._stop}
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
					Key: &ast.Identifier{Name: "_value"},
					Value: &ast.CallExpression{
						Callee: &ast.Identifier{Name: fn},
						Arguments: []ast.Expression{
							&ast.ObjectExpression{
								Properties: []*ast.Property{
									&ast.Property{
										Key: &ast.Identifier{Name: "x"},
										Value: &ast.MemberExpression{
											Object: &ast.Identifier{
												Name: "r",
											},
											Property: &ast.Identifier{
												Name: "_value",
											},
										},
									},
								},
							},
						},
					},
				},
				{
					Key: &ast.Identifier{Name: "_stop"},
					Value: &ast.MemberExpression{
						Object: &ast.Identifier{
							Name: "r",
						},
						Property: &ast.Identifier{
							Name: "_stop",
						},
					},
				},
			},
		},
	}
}

var filterWindowsWithZeroValueCall = call(
	"filter",
	map[string]ast.Expression{
		"fn": &ast.FunctionExpression{
			Params: []*ast.Property{
				{
					Key: &ast.Identifier{
						Name: "r",
					},
				},
			},
			Body: &ast.BinaryExpression{
				Operator: ast.GreaterThanOperator,
				Left: &ast.MemberExpression{
					Object: &ast.Identifier{
						Name: "r",
					},
					Property: &ast.Identifier{
						Name: "_value",
					},
				},
				Right: &ast.FloatLiteral{Value: 0},
			},
		},
	},
)

func (t *transpiler) transpileCall(c *promql.Call) (ast.Expression, error) {
	if fn, ok := aggregateOverTimeFns[c.Func.Name]; ok {
		v, err := t.transpileExpr(c.Args[0])
		if err != nil {
			return nil, fmt.Errorf("error transpiling function argument: %s", err)
		}

		args := map[string]ast.Expression{}
		if fn == "quantile" {
			args["q"] = v
			args["method"] = &ast.StringLiteral{Value: "exact_mean"}

			v, err = t.transpileExpr(c.Args[1])
			if err != nil {
				return nil, fmt.Errorf("error transpiling function argument")
			}
		}

		if fn == "stddev" {
			args["mode"] = &ast.StringLiteral{Value: "population"}
		}

		// "toFloat" and "filter" can both not deal with null values.
		fillCall := call("fill", map[string]ast.Expression{
			"column": &ast.StringLiteral{Value: "_value"},
			"value":  &ast.FloatLiteral{Value: nullReplacement},
		})
		// "count" is the only aggregation function that returns an int.
		if fn == "count" {
			fillCall = call("fill", map[string]ast.Expression{
				"column": &ast.StringLiteral{Value: "_value"},
				"value":  &ast.IntegerLiteral{Value: nullReplacement},
			})
		}

		pipelineCalls := []*ast.CallExpression{
			call(fn, args),
			fillCall,
			call("toFloat", nil),
			filterSpecialNullValuesCall,
			dropMeasurementCall,
		}
		if fn == "count" {
			// Count is the only function that produces a 0 instead of null value for an empty table.
			// In PromQL, when we count_over_time() over an empty range, the result is empty, so we need
			// to filter away 0 values here.
			pipelineCalls = append(pipelineCalls, filterWindowsWithZeroValueCall)
		}
		if c.Func.Name == "stdvar_over_time" {
			pipelineCalls = append(
				pipelineCalls,
				call("map", map[string]ast.Expression{
					"fn": scalarArithBinaryMathFn("pow", &ast.FloatLiteral{Value: 2}, false),
				}),
			)
		}

		return buildPipeline(
			v,
			pipelineCalls...,
		), nil
	}

	if fn, ok := vectorMathFunctions[c.Func.Name]; ok {
		v, err := t.transpileExpr(c.Args[0])
		if err != nil {
			return nil, fmt.Errorf("error transpiling function argument")
		}

		return buildPipeline(
			v,
			call("map", map[string]ast.Expression{"fn": vectorMathFn(fn)}),
			dropMeasurementCall,
		), nil
	}

	switch c.Func.Name {
	case "rate", "delta", "increase":
		isCounter := true
		isRate := true

		if c.Func.Name == "delta" {
			isCounter = false
			isRate = false
		}
		if c.Func.Name == "increase" {
			isRate = false
		}

		v, err := t.transpileExpr(c.Args[0])
		if err != nil {
			return nil, fmt.Errorf("error transpiling function argument")
		}
		return buildPipeline(
			v,
			call("promql.extrapolatedRate", map[string]ast.Expression{
				"isCounter": &ast.BooleanLiteral{Value: isCounter},
				"isRate":    &ast.BooleanLiteral{Value: isRate},
			}),
			dropMeasurementCall,
		), nil
	case "timestamp":
		v, err := t.transpileExpr(c.Args[0])
		if err != nil {
			return nil, fmt.Errorf("error transpiling function argument")
		}
		return buildPipeline(
			v,
			call("promql.timestamp", nil),
			dropMeasurementCall,
		), nil
	case "time":
		return buildPipeline(
			call("promql.emptyTable", nil),
			call("range", map[string]ast.Expression{
				"start": &ast.DateTimeLiteral{Value: t.start.Add(-5 * time.Minute)},
				"stop":  &ast.DateTimeLiteral{Value: t.end},
			}),
			call("window", map[string]ast.Expression{
				"every":       &ast.DurationLiteral{Values: []ast.Duration{{Magnitude: t.resolution.Nanoseconds(), Unit: "ns"}}},
				"period":      &ast.DurationLiteral{Values: []ast.Duration{{Magnitude: 5, Unit: "m"}}},
				"createEmpty": &ast.BooleanLiteral{Value: true},
			}),
			call("sum", nil),
			// Remove any windows <5m long at the edges of the graph range to act like PromQL.
			call("filter", map[string]ast.Expression{"fn": windowCutoffFn(t.start, t.end.Add(-5*time.Minute))}),
			call("promql.timestamp", nil),
		), nil
	case "clamp_max", "clamp_min":
		fn := "math.mMax"
		if c.Func.Name == "clamp_max" {
			fn = "math.mMin"
		}

		v, err := t.transpileExpr(c.Args[0])
		if err != nil {
			return nil, fmt.Errorf("error transpiling function argument")
		}
		clamp, err := t.transpileExpr(c.Args[1])
		if err != nil {
			return nil, fmt.Errorf("error transpiling function argument")
		}
		return buildPipeline(
			v,
			call("map", map[string]ast.Expression{
				"fn": scalarArithBinaryMathFn(fn, clamp, false),
			}),
			dropMeasurementCall,
		), nil
	default:
		return nil, fmt.Errorf("PromQL function %q is not supported yet", c.Func.Name)
	}
}
