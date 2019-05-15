package promflux

import (
	"fmt"
	"time"

	"github.com/influxdata/flux/ast"
	"github.com/prometheus/common/model"
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
	"stdvar_over_time":   "stdvar", // TODO: Add stdvar() to Flux stdlib instead of special-casing this below.
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

var dateFunctions = map[string]string{
	"day_of_month":  "promql.dayOfMonth",
	"day_of_week":   "promql.dayOfWeek",
	"days_in_month": "promql.daysInMonth",
	"hour":          "promql.hour",
	"minute":        "promql.minute",
	"month":         "promql.month",
	"year":          "promql.year",
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
				Left:     member("r", "_value"),
				Right:    &ast.FloatLiteral{Value: nullReplacement},
			},
		},
	},
)

// Function to apply a simple one-operand function to all values in a table.
func singleArgFloatFn(fn string, argName string) *ast.FunctionExpression {
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
					Value: call(fn, map[string]ast.Expression{
						argName: member("r", "_value"),
					}),
				},
				{
					Key:   &ast.Identifier{Name: "_stop"},
					Value: member("r", "_stop"),
				},
			},
		},
	}
}

// Function to set all values to a constant.
func setConstValueFn(v ast.Expression) *ast.FunctionExpression {
	// (r) => {"_value": <v>, "_stop": r._stop}
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
					Value: v,
				},
				{
					Key:   &ast.Identifier{Name: "_stop"},
					Value: member("r", "_stop"),
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
				Left:     member("r", "_value"),
				Right:    &ast.FloatLiteral{Value: 0},
			},
		},
	},
)

func (t *transpiler) transpileAggregateOverTimeFunc(fn string, inArgs []ast.Expression) (ast.Expression, error) {
	callFn := fn
	vec := inArgs[0]
	args := map[string]ast.Expression{}
	var nullValue ast.Expression = &ast.FloatLiteral{Value: nullReplacement}

	switch fn {
	case "count":
		// "count" is the only aggregation function that returns an int, so we
		// can only replace its null values with integers, not floats.
		nullValue = &ast.IntegerLiteral{Value: nullReplacement}
	case "quantile":
		vec = inArgs[1]
		args["q"] = inArgs[0]
		args["method"] = &ast.StringLiteral{Value: "exact_mean"}
	case "stddev", "stdvar":
		callFn = "stddev"
		args["mode"] = &ast.StringLiteral{Value: "population"}
	}

	pipelineCalls := []*ast.CallExpression{
		call(callFn, args),
		call("fill", map[string]ast.Expression{
			"column": &ast.StringLiteral{Value: "_value"},
			"value":  nullValue,
		}),
		call("toFloat", nil),
		filterSpecialNullValuesCall,
		dropMeasurementCall,
	}

	switch fn {
	case "count":
		// Count is the only function that produces a 0 instead of null value for an empty table.
		// In PromQL, when we count_over_time() over an empty range, the result is empty, so we need
		// to filter away 0 values here.
		pipelineCalls = append(pipelineCalls, filterWindowsWithZeroValueCall)
	case "stdvar":
		pipelineCalls = append(
			pipelineCalls,
			call("map", map[string]ast.Expression{
				"fn": scalarArithBinaryMathFn("pow", &ast.FloatLiteral{Value: 2}, false),
			}),
		)
	}

	return buildPipeline(
		vec,
		pipelineCalls...,
	), nil
}

func labelJoinFn(srcLabels []*ast.StringLiteral, dst *ast.StringLiteral, sep *ast.StringLiteral) *ast.FunctionExpression {
	// TODO: Deal with empty source labels! Use Flux conditionals to check for existence?
	var dstLabelValue ast.Expression = member("r", srcLabels[0].Value)
	for _, srcLabel := range srcLabels[1:] {
		dstLabelValue = &ast.BinaryExpression{
			Operator: ast.AdditionOperator,
			Left:     dstLabelValue,
			Right: &ast.BinaryExpression{
				Operator: ast.AdditionOperator,
				Left:     sep,
				Right:    member("r", srcLabel.Value),
			},
		}
	}

	// (r) => ({<dst>: <src1><sep><src2>...})
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
					Key:   &ast.Identifier{Name: dst.Value},
					Value: dstLabelValue,
				},
				{
					Key:   &ast.Identifier{Name: "_value"},
					Value: member("r", "_value"),
				},
			},
		},
	}
}

func (t *transpiler) generateZeroWindows() *ast.PipeExpression {
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
	)
}

func (t *transpiler) timeFn() *ast.PipeExpression {
	return buildPipeline(
		t.generateZeroWindows(),
		call("promql.timestamp", nil),
	)
}

func (t *transpiler) transpileCall(c *promql.Call) (ast.Expression, error) {
	// The PromQL parser already verifies argument counts and types, so we don't have to check this here.
	args := make([]ast.Expression, len(c.Args))
	for i, arg := range c.Args {
		tArg, err := t.transpileExpr(arg)
		if err != nil {
			return nil, fmt.Errorf("error transpiling function argument: %s", err)
		}
		args[i] = tArg
	}

	// {count,avg,sum,min,max,...}_over_time()
	if fn, ok := aggregateOverTimeFns[c.Func.Name]; ok {
		return t.transpileAggregateOverTimeFunc(fn, args)
	}

	// abs(), ceil(), round()...
	if fn, ok := vectorMathFunctions[c.Func.Name]; ok {
		return buildPipeline(
			args[0],
			call("map", map[string]ast.Expression{"fn": singleArgFloatFn(fn, "x")}),
			dropMeasurementCall,
		), nil
	}

	// day_of_month(), hour(), etc.
	if fn, ok := dateFunctions[c.Func.Name]; ok {
		var v ast.Expression
		if len(args) == 0 {
			v = t.timeFn()
		} else {
			v = args[0]
		}

		return buildPipeline(
			v,
			call("map", map[string]ast.Expression{"fn": singleArgFloatFn(fn, "timestamp")}),
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

		return buildPipeline(
			args[0],
			call("promql.extrapolatedRate", map[string]ast.Expression{
				"isCounter": &ast.BooleanLiteral{Value: isCounter},
				"isRate":    &ast.BooleanLiteral{Value: isRate},
			}),
			dropMeasurementCall,
		), nil
	case "irate", "idelta":
		isRate := true

		if c.Func.Name == "idelta" {
			isRate = false
		}

		return buildPipeline(
			args[0],
			call("promql.instantRate", map[string]ast.Expression{
				"isRate": &ast.BooleanLiteral{Value: isRate},
			}),
			dropMeasurementCall,
		), nil
	case "deriv":
		return buildPipeline(
			args[0],
			call("promql.linearRegression", nil),
			dropMeasurementCall,
		), nil
	case "predict_linear":
		if yieldsTable(c.Args[1]) {
			return nil, fmt.Errorf("non-const scalar expressions not supported yet")
		}

		return buildPipeline(
			args[0],
			call("promql.linearRegression", map[string]ast.Expression{
				"predict": &ast.BooleanLiteral{Value: true},
				"fromNow": args[1],
			}),
			dropMeasurementCall,
		), nil
	case "timestamp":
		return buildPipeline(
			args[0],
			call("promql.timestamp", nil),
			dropMeasurementCall,
		), nil
	case "time":
		return t.timeFn(), nil
	case "changes", "resets":
		fn := "promql." + c.Func.Name

		return buildPipeline(
			args[0],
			call(fn, nil),
			dropMeasurementCall,
		), nil
	case "clamp_max", "clamp_min":
		fn := "math.mMax"
		if c.Func.Name == "clamp_max" {
			fn = "math.mMin"
		}

		v := args[0]
		clamp := args[1]
		return buildPipeline(
			v,
			call("map", map[string]ast.Expression{
				"fn": scalarArithBinaryMathFn(fn, clamp, false),
			}),
			dropMeasurementCall,
		), nil
	case "label_join":
		v := args[0]

		dst, ok := args[1].(*ast.StringLiteral)
		if !ok {
			return nil, fmt.Errorf("label_join() destination label must be string literal")
		}
		if !model.LabelName(dst.Value).IsValid() {
			return nil, fmt.Errorf("invalid destination label name in label_join(): %s", dst.Value)
		}
		dst.Value = escapeLabelName(dst.Value)

		sep, ok := args[2].(*ast.StringLiteral)
		if !ok {
			return nil, fmt.Errorf("label_join() separator must be string literal")
		}

		srcLabels := make([]*ast.StringLiteral, len(args)-3)
		for i := 3; i < len(args); i++ {
			src, ok := args[i].(*ast.StringLiteral)
			if !ok {
				return nil, fmt.Errorf("label_join() source labels must be string literals")
			}
			if !model.LabelName(src.Value).IsValid() {
				return nil, fmt.Errorf("invalid source label name in label_join(): %s", src.Value)
			}
			src.Value = escapeLabelName(src.Value)
			srcLabels[i-3] = src
		}

		return buildPipeline(
			v,
			call("map", map[string]ast.Expression{"fn": labelJoinFn(srcLabels, dst, sep)}),
		), nil
	case "vector":
		if yieldsTable(c.Args[0]) {
			return args[0], nil
		}
		return buildPipeline(
			t.generateZeroWindows(),
			call("map", map[string]ast.Expression{
				"fn": setConstValueFn(args[0]),
			}),
		), nil
	case "histogram_quantile":
		if yieldsTable(c.Args[0]) {
			return nil, fmt.Errorf("non-const scalar expressions not supported yet")
		}

		return buildPipeline(
			args[1],
			call("group", map[string]ast.Expression{
				"columns": columnList("_time", "_value", "le"),
				"mode":    &ast.StringLiteral{Value: "except"},
			}),
			call("promql.promHistogramQuantile", map[string]ast.Expression{
				"quantile": args[0],
			}),
			dropMeasurementCall,
		), nil
	default:
		return nil, fmt.Errorf("PromQL function %q is not supported yet", c.Func.Name)
	}
}
