package promflux

import (
	"fmt"
	"time"

	"github.com/influxdata/flux/ast"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/promql"
)

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

var filterNullValuesCall = call(
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
			Body: &ast.UnaryExpression{
				Operator: ast.ExistsOperator,
				Argument: member("r", "_value"),
			},
		},
	},
)

// Function to apply a simple one-operand function to all values in a table.
func singleArgFloatFn(fn string, argName string) *ast.FunctionExpression {
	// (r) => {r with _value: mathFn(x: r._value), _stop: r._stop}
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
	// (r) => {r with _value: <v>, _stop: r._stop}
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

func (t *Transpiler) transpileAggregateOverTimeFunc(fn string, inArgs []ast.Expression) (ast.Expression, error) {
	callFn := fn
	vec := inArgs[0]
	args := map[string]ast.Expression{}

	switch fn {
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
		filterNullValuesCall,
		call("toFloat", nil),
		dropFieldAndTimeCall,
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

	// (r) => ({r with <dst>: <src1><sep><src2>...})
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

func (t *Transpiler) generateZeroWindows() *ast.PipeExpression {
	var windowCall *ast.CallExpression
	var windowFilterCall *ast.CallExpression
	if t.Resolution > 0 {
		// For range queries:
		// At every resolution step, load / look back up to 5m of data (PromQL lookback delta).
		windowCall = call("window", map[string]ast.Expression{
			"every":       &ast.DurationLiteral{Values: []ast.Duration{{Magnitude: t.Resolution.Nanoseconds(), Unit: "ns"}}},
			"period":      &ast.DurationLiteral{Values: []ast.Duration{{Magnitude: 5, Unit: "m"}}},
			"offset":      &ast.DurationLiteral{Values: []ast.Duration{{Magnitude: t.Start.UnixNano() % t.Resolution.Nanoseconds(), Unit: "ns"}}},
			"createEmpty": &ast.BooleanLiteral{Value: true},
		})

		// Remove any windows <5m long at the edges of the graph range to act like PromQL.
		windowFilterCall = call("filter", map[string]ast.Expression{"fn": windowCutoffFn(t.Start, t.End.Add(-5*time.Minute))})
	}

	return buildPipeline(
		call("promql.emptyTable", nil),
		call("range", map[string]ast.Expression{
			"start": &ast.DateTimeLiteral{Value: t.Start.Add(-5 * time.Minute)},
			"stop":  &ast.DateTimeLiteral{Value: t.End},
		}),
		windowCall,
		call("sum", nil),
		windowFilterCall,
	)
}

func (t *Transpiler) timeFn() *ast.PipeExpression {
	return buildPipeline(
		t.generateZeroWindows(),
		call("promql.timestamp", nil),
	)
}

func (t *Transpiler) transpileCall(c *promql.Call) (ast.Expression, error) {
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
			dropFieldAndTimeCall,
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
			dropFieldAndTimeCall,
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
			dropFieldAndTimeCall,
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
			dropFieldAndTimeCall,
		), nil
	case "deriv":
		return buildPipeline(
			args[0],
			call("promql.linearRegression", nil),
			dropFieldAndTimeCall,
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
			dropFieldAndTimeCall,
		), nil
	case "holt_winters":
		if yieldsTable(c.Args[1]) || yieldsTable(c.Args[2]) {
			return nil, fmt.Errorf("non-const scalar expressions not supported yet")
		}

		return buildPipeline(
			args[0],
			call("promql.holtWinters", map[string]ast.Expression{
				"smoothingFactor": args[1],
				"trendFactor":     args[2],
			}),
			dropFieldAndTimeCall,
		), nil
	case "timestamp":
		return buildPipeline(
			args[0],
			call("promql.timestamp", nil),
			dropFieldAndTimeCall,
		), nil
	case "time":
		return t.timeFn(), nil
	case "changes", "resets":
		fn := "promql." + c.Func.Name

		return buildPipeline(
			args[0],
			call(fn, nil),
			dropFieldAndTimeCall,
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
			dropFieldAndTimeCall,
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
	case "label_replace":
		for _, arg := range args[1:] {
			if _, ok := arg.(*ast.StringLiteral); !ok {
				return nil, fmt.Errorf("non-literal string arguments not supported yet in label_replace()")
			}
		}

		return buildPipeline(
			args[0],
			call("promql.labelReplace", map[string]ast.Expression{
				"destination": args[1],
				"replacement": args[2],
				"source":      args[3],
				"regex":       args[4],
			}),
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
	case "scalar":
		// TODO: Need to insert NaN values at time steps where there is no value in the vector.
		// This requires new outer join support.
		return buildPipeline(
			args[0],
			call("keep", map[string]ast.Expression{
				"columns": columnList("_stop", "_value"),
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
			dropFieldAndTimeCall,
		), nil
	default:
		return nil, fmt.Errorf("PromQL function %q is not supported yet", c.Func.Name)
	}
}
