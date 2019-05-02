package main

import (
	"fmt"
	"regexp"
	"time"

	"github.com/influxdata/flux/ast"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/promql"
)

// TODO: Temporary hack to work around lack of null support in filter(). Remove this.
const nullReplacement = 123456789

type transpiler struct {
	bucket     string
	start      time.Time
	end        time.Time
	resolution time.Duration
}

var labelMatchOps = map[labels.MatchType]ast.OperatorKind{
	labels.MatchEqual:     ast.EqualOperator,
	labels.MatchNotEqual:  ast.NotEqualOperator,
	labels.MatchRegexp:    ast.RegexpMatchOperator,
	labels.MatchNotRegexp: ast.NotRegexpMatchOperator,
}

func transpileLabelMatchersFn(lms []*labels.Matcher) *ast.FunctionExpression {
	return &ast.FunctionExpression{
		Params: []*ast.Property{
			{
				Key: &ast.Identifier{Name: "r"},
			},
		},
		Body: transpileLabelMatchers(lms),
	}
}

func transpileLabelMatchers(lms []*labels.Matcher) ast.Expression {
	if len(lms) == 0 {
		panic("empty label matchers")
	}
	if len(lms) == 1 {
		return transpileLabelMatcher(lms[0])
	}
	return &ast.LogicalExpression{
		Operator: ast.AndOperator,
		Left:     transpileLabelMatcher(lms[0]),
		// Recurse until we have all label matchers AND-ed together in a right-heavy tree.
		Right: transpileLabelMatchers(lms[1:]),
	}
}

func transpileLabelMatcher(lm *labels.Matcher) *ast.BinaryExpression {
	op, ok := labelMatchOps[lm.Type]
	if !ok {
		panic(fmt.Errorf("invalid label matcher type %v", lm.Type))
	}
	if lm.Name == model.MetricNameLabel {
		lm.Name = "_measurement"
	}
	be := &ast.BinaryExpression{
		Operator: op,
		Left: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "r"},
			Property: &ast.Identifier{Name: lm.Name},
		},
	}
	if op == ast.EqualOperator || op == ast.NotEqualOperator {
		be.Right = &ast.StringLiteral{Value: lm.Value}
	} else {
		// PromQL parsing already validates regexes.
		// PromQL regexes are always full-string matches / fully anchored.
		be.Right = &ast.RegexpLiteral{Value: regexp.MustCompile("^(?:" + lm.Value + ")$")}
	}
	return be
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
	expr := &ast.CallExpression{
		Callee: &ast.Identifier{Name: fn},
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

// Function to remove any windows that are <5m long.
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
				Left: &ast.MemberExpression{
					Object: &ast.Identifier{
						Name: "r",
					},
					Property: &ast.Identifier{
						Name: "_stop",
					},
				},
				Right: &ast.DateTimeLiteral{Value: minStop},
			},
			Right: &ast.BinaryExpression{
				Operator: ast.LessThanEqualOperator,
				Left: &ast.MemberExpression{
					Object: &ast.Identifier{
						Name: "r",
					},
					Property: &ast.Identifier{
						Name: "_start",
					},
				},
				Right: &ast.DateTimeLiteral{Value: maxStart},
			},
		},
	}
}

func (t *transpiler) transpileInstantVectorSelector(v *promql.VectorSelector) *ast.PipeExpression {
	return buildPipeline(
		// Select all Prometheus data.
		call("from", map[string]ast.Expression{"bucket": &ast.StringLiteral{Value: t.bucket}}),
		// Query entire graph range.
		call("range", map[string]ast.Expression{
			"start": &ast.DateTimeLiteral{Value: t.start.Add(-5*time.Minute - v.Offset)},
			"stop":  &ast.DateTimeLiteral{Value: t.end.Add(-v.Offset)},
		}),
		// Apply label matching filters.
		call("filter", map[string]ast.Expression{"fn": transpileLabelMatchersFn(v.LabelMatchers)}),
		// At every resolution step, load / look back up to 5m of data (PromQL lookback delta).
		call("window", map[string]ast.Expression{
			"every":  &ast.DurationLiteral{Values: []ast.Duration{{Magnitude: t.resolution.Nanoseconds(), Unit: "ns"}}},
			"period": &ast.DurationLiteral{Values: []ast.Duration{{Magnitude: 5, Unit: "m"}}},
		}),
		// Remove any windows <5m long at the edges of the graph range to act like PromQL.
		call("filter", map[string]ast.Expression{"fn": windowCutoffFn(t.start.Add(-v.Offset), t.end.Add(-5*time.Minute-v.Offset))}),
		// Select the last data point after the current evaluation (resolution step) timestamp.
		call("last", nil),
		// Apply offsets to make past data look like it's in the present.
		call("timeShift", map[string]ast.Expression{
			"duration": &ast.DurationLiteral{Values: []ast.Duration{{Magnitude: v.Offset.Nanoseconds(), Unit: "ns"}}},
		}),
	)
}

func (t *transpiler) transpileRangeVectorSelector(v *promql.MatrixSelector) *ast.PipeExpression {
	return buildPipeline(
		// Select all Prometheus data.
		call("from", map[string]ast.Expression{"bucket": &ast.StringLiteral{Value: t.bucket}}),
		// Query entire graph range.
		call("range", map[string]ast.Expression{
			"start": &ast.DateTimeLiteral{Value: t.start.Add(-v.Range - v.Offset)},
			"stop":  &ast.DateTimeLiteral{Value: t.end.Add(-v.Offset)},
		}),
		// Apply label matching filters.
		call("filter", map[string]ast.Expression{"fn": transpileLabelMatchersFn(v.LabelMatchers)}),
		// At every resolution step, include the specified range of data.
		call("window", map[string]ast.Expression{
			"every":  &ast.DurationLiteral{Values: []ast.Duration{{Magnitude: t.resolution.Nanoseconds(), Unit: "ns"}}},
			"period": &ast.DurationLiteral{Values: []ast.Duration{{Magnitude: v.Range.Nanoseconds(), Unit: "ns"}}},
		}),
		// Remove any windows smaller than the specified range at the edges of the graph range.
		call("filter", map[string]ast.Expression{"fn": windowCutoffFn(t.start.Add(-v.Offset), t.end.Add(-v.Range-v.Offset))}),
		// Apply offsets to make past data look like it's in the present.
		call("timeShift", map[string]ast.Expression{
			"duration": &ast.DurationLiteral{Values: []ast.Duration{{Magnitude: v.Offset.Nanoseconds(), Unit: "ns"}}},
		}),
	)
}

func columnList(strs ...string) *ast.ArrayExpression {
	list := make([]ast.Expression, len(strs))
	for i, str := range strs {
		if str == model.MetricNameLabel {
			str = "_measurement"
		}
		list[i] = &ast.StringLiteral{Value: str}
	}
	return &ast.ArrayExpression{
		Elements: list,
	}
}

type aggregateFn struct {
	name            string
	dropMeasurement bool
	// All PromQL aggregation operators drop non-grouping labels, but some
	// of the (non-aggregation) Flux counterparts don't. This field indicates
	// that the non-grouping drop needs to be explicitly added to the pipeline.
	dropNonGrouping bool
}

var aggregateFns = map[promql.ItemType]aggregateFn{
	promql.ItemSum:      {name: "sum", dropMeasurement: true, dropNonGrouping: false},
	promql.ItemAvg:      {name: "mean", dropMeasurement: true, dropNonGrouping: false},
	promql.ItemMax:      {name: "max", dropMeasurement: true, dropNonGrouping: true},
	promql.ItemMin:      {name: "min", dropMeasurement: true, dropNonGrouping: true},
	promql.ItemCount:    {name: "count", dropMeasurement: true, dropNonGrouping: false},
	promql.ItemStddev:   {name: "stddev", dropMeasurement: true, dropNonGrouping: false},
	promql.ItemStdvar:   {name: "stddev", dropMeasurement: true, dropNonGrouping: false},
	promql.ItemTopK:     {name: "top", dropMeasurement: false, dropNonGrouping: false},
	promql.ItemBottomK:  {name: "bottom", dropMeasurement: false, dropNonGrouping: false},
	promql.ItemQuantile: {name: "quantile", dropMeasurement: true, dropNonGrouping: false},
}

func dropNonGroupingColsCall(groupCols []string, without bool) *ast.CallExpression {
	if without {
		cols := make([]string, 0, len(groupCols))
		// Remove "_value" from list of columns to drop.
		for _, col := range groupCols {
			if col != "_value" { // TODO: also _start, _stop?
				cols = append(cols, col)
			}
		}

		// TODO: This errors with non-existent columns. In PromQL, this is a no-op.
		return call("drop", map[string]ast.Expression{"columns": columnList(cols...)})
	}

	// We want to keep value and time columns even if they are not explicitly in the grouping labels.
	cols := append(groupCols, "_value", "_start", "_stop")
	return call("keep", map[string]ast.Expression{"columns": columnList(cols...)})
}

// Taken from Prometheus.
const (
	// The largest SampleValue that can be converted to an int64 without overflow.
	maxInt64 = 9223372036854774784
	// The smallest SampleValue that can be converted to an int64 without underflow.
	minInt64 = -9223372036854775808
)

// convertibleToInt64 returns true if v does not over-/underflow an int64.
func convertibleToInt64(v float64) bool {
	return v <= maxInt64 && v >= minInt64
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

func (t *transpiler) transpileAggregateExpr(a *promql.AggregateExpr) (ast.Expression, error) {
	expr, err := t.transpileExpr(a.Expr)
	if err != nil {
		return nil, fmt.Errorf("error transpiling aggregate sub-expression: %s", err)
	}

	aggFn, ok := aggregateFns[a.Op]
	if !ok {
		return nil, fmt.Errorf("unsupported aggregation type %s", a.Op)
	}

	groupCols := columnList(a.Grouping...)
	aggArgs := map[string]ast.Expression{}

	if a.Op == promql.ItemTopK || a.Op == promql.ItemBottomK {
		// The PromQL parser already verifies that a.Param is a scalar.
		n, ok := a.Param.(*promql.NumberLiteral)
		if !ok {
			return nil, fmt.Errorf("arbitrary scalar subexpressions not supported yet")
		}
		if !convertibleToInt64(n.Val) {
			return nil, fmt.Errorf("scalar value %v overflows int64", n)
		}
		aggArgs["n"] = &ast.IntegerLiteral{Value: int64(n.Val)}
	}

	if a.Op == promql.ItemQuantile {
		// The PromQL already verifies that a.Param is a scalar.
		n, ok := a.Param.(*promql.NumberLiteral)
		if !ok {
			return nil, fmt.Errorf("arbitrary scalar subexpressions not supported yet")
		}
		aggArgs["q"] = &ast.FloatLiteral{Value: n.Val}
		aggArgs["method"] = &ast.StringLiteral{Value: "exact_mean"}
	}

	if a.Op == promql.ItemStddev || a.Op == promql.ItemStdvar {
		aggArgs["mode"] = &ast.StringLiteral{Value: "population"}
	}

	mode := "by"
	dropMeasurement := true
	if a.Without {
		mode = "except"
		groupCols.Elements = append(
			groupCols.Elements,
			// "_time" is not always present, but if it is, we don't want to group by it.
			&ast.StringLiteral{Value: "_time"},
			&ast.StringLiteral{Value: "_value"},
		)
	} else {
		groupCols.Elements = append(
			groupCols.Elements,
			&ast.StringLiteral{Value: "_start"},
			&ast.StringLiteral{Value: "_stop"},
		)
		for _, col := range a.Grouping {
			if col == model.MetricNameLabel {
				dropMeasurement = false
			}
		}
	}

	pipeline := buildPipeline(
		// Get the underlying data.
		expr,
		// Group values according to by() / without() clauses.
		call("group", map[string]ast.Expression{
			"columns": groupCols,
			"mode":    &ast.StringLiteral{Value: mode},
		}),
		// Aggregate.
		call(aggFn.name, aggArgs),
		// TODO: Change this in the language to drop empty tables?
		// Remove any windows <5m long at the end of the graph range to act like PromQL.
		// Even if those windows were filtered by a vector selector previously, they might
		// exist as empty tables and then the Flux aggregator functions would records rows with null
		// values for each empty table, which can then confuse further filtering etc. steps.
		//
		// TODO: NOTE: This actually breaks things for windows different than 5m (like a "rate(foo[1m])")
		// and doesn't actually seem to be needed anymore?
		//call("filter", map[string]ast.Expression{"fn": windowCutoffFn(t.start, t.end.Add(-5*time.Minute))}),
	)
	if aggFn.dropNonGrouping {
		// Drop labels that are not part of the grouping.
		pipeline = buildPipeline(pipeline, dropNonGroupingColsCall(a.Grouping, a.Without))
	}
	if aggFn.dropMeasurement && dropMeasurement {
		pipeline = buildPipeline(
			pipeline,
			dropMeasurementCall,
		)
	}
	if a.Op == promql.ItemStdvar {
		pipeline = buildPipeline(
			pipeline,
			call("map", map[string]ast.Expression{
				"fn": scalarArithBinaryMathFn("pow", &ast.FloatLiteral{Value: 2}, false),
			}),
		)
	}
	return pipeline, nil
}

var arithBinOps = map[promql.ItemType]ast.OperatorKind{
	promql.ItemADD: ast.AdditionOperator,
	promql.ItemSUB: ast.SubtractionOperator,
	promql.ItemMUL: ast.MultiplicationOperator,
	promql.ItemDIV: ast.DivisionOperator,
}

var arithBinOpFns = map[promql.ItemType]string{
	promql.ItemPOW: "pow",
	promql.ItemMOD: "mod",
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
	val := &ast.MemberExpression{
		Object: &ast.Identifier{
			Name: "r",
		},
		Property: &ast.Identifier{
			Name: "_value",
		},
	}

	var lhs, rhs ast.Expression = val, operand

	if swapped {
		lhs, rhs = rhs, lhs
	}

	// (r) => {"_value": <lhs> <op> <rhs>, "_stop": r._stop}
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
					Value: &ast.BinaryExpression{
						Operator: op,
						Left:     lhs,
						Right:    rhs,
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

// Function to apply a math function to all values in a table and a given float64 operand.
func scalarArithBinaryMathFn(mathFn string, operand ast.Expression, swapped bool) *ast.FunctionExpression {
	val := &ast.MemberExpression{
		Object: &ast.Identifier{
			Name: "r",
		},
		Property: &ast.Identifier{
			Name: "_value",
		},
	}

	var lhs, rhs ast.Expression = val, operand

	if swapped {
		lhs, rhs = rhs, lhs
	}

	// (r) => {"_value": <lhs> <op> <rhs>, "_stop": r._stop}
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

// Function to apply a comparison binary operator to all values in a table and a given float64 operand.
func scalarCompBinaryOpFn(op ast.OperatorKind, operand ast.Expression, swapped bool) *ast.FunctionExpression {
	val := &ast.MemberExpression{
		Object: &ast.Identifier{
			Name: "r",
		},
		Property: &ast.Identifier{
			Name: "_value",
		},
	}

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
	lhs := &ast.MemberExpression{
		Object: &ast.Identifier{
			Name: "r",
		},
		Property: &ast.Identifier{
			Name: "_value_lhs",
		},
	}
	rhs := &ast.MemberExpression{
		Object: &ast.Identifier{
			Name: "r",
		},
		Property: &ast.Identifier{
			Name: "_value_rhs",
		},
	}

	// (r) => {"_value": <lhs> <op> <rhs>, "_stop": r._stop}
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
					Value: &ast.BinaryExpression{
						Operator: op,
						Left:     lhs,
						Right:    rhs,
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

// Function to apply a binary arithmetic operator math function between values of two joined tables.
func vectorArithBinaryMathFn(mathFn string) *ast.FunctionExpression {
	lhs := &ast.MemberExpression{
		Object: &ast.Identifier{
			Name: "r",
		},
		Property: &ast.Identifier{
			Name: "_value_lhs",
		},
	}
	rhs := &ast.MemberExpression{
		Object: &ast.Identifier{
			Name: "r",
		},
		Property: &ast.Identifier{
			Name: "_value_rhs",
		},
	}

	// (r) => {"_value": mathFn(<lhs>, <rhs>), "_stop": r._stop}
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

// Function to apply a binary comparison operator between values of two joined tables.
func vectorCompBinaryOpFn(op ast.OperatorKind) *ast.FunctionExpression {
	lhs := &ast.MemberExpression{
		Object: &ast.Identifier{
			Name: "r",
		},
		Property: &ast.Identifier{
			Name: "_value_lhs",
		},
	}
	rhs := &ast.MemberExpression{
		Object: &ast.Identifier{
			Name: "r",
		},
		Property: &ast.Identifier{
			Name: "_value_rhs",
		},
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

func (t *transpiler) transpileBinaryExpr(b *promql.BinaryExpr) (ast.Expression, error) {
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
				panic("scalar-to-scalar binary op is missing 'bool' modifier")
			}

			return call("float", map[string]ast.Expression{
				"v": &ast.BinaryExpression{
					Operator: op,
					Left:     lhs,
					Right:    rhs,
				},
			}), nil
		}

		panic(fmt.Errorf("invalid scalar-scalar binary op %q", b.Op))
	case yieldsFloat(b.LHS) && yieldsTable(b.RHS):
		lhs, rhs = rhs, lhs
		swapped = true
		fallthrough
	case yieldsTable(b.LHS) && yieldsFloat(b.RHS):
		if op, ok := arithBinOps[b.Op]; ok {
			return buildPipeline(
				lhs,
				call("map", map[string]ast.Expression{"fn": scalarArithBinaryOpFn(op, rhs, swapped)}),
				dropMeasurementCall,
			), nil
		}

		if opFn, ok := arithBinOpFns[b.Op]; ok {
			return buildPipeline(
				lhs,
				call("map", map[string]ast.Expression{"fn": scalarArithBinaryMathFn(opFn, rhs, swapped)}),
				dropMeasurementCall,
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
					dropMeasurementCall,
				), nil
			}
			return buildPipeline(
				lhs,
				call("filter", map[string]ast.Expression{"fn": scalarCompBinaryOpFn(op, rhs, swapped)}),
			), nil
		}

		panic(fmt.Errorf("invalid scalar-vector binary op %q", b.Op))
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

		dropMeasurement := true
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
				dropMeasurement = false
			}
		} else {
			return nil, fmt.Errorf("vector set operations not supported yet")
		}

		onCols := append(b.VectorMatching.MatchingLabels, "_start", "_stop")

		outputColTransformCalls := []*ast.CallExpression{
			call("keep", map[string]ast.Expression{
				"columns": columnList(append(append(onCols, "_value"), b.VectorMatching.Include...)...),
			}),

			// // Rename x_lhs -> x.
			// call("rename", map[string]ast.Expression{"fn": stripSuffixFn("_lhs")}),
			// // Drop cols RHS cols, except ones we want to copy into the result via a group_x(...) clause.
			// // AH FUCK
			// call("drop", map[string]ast.Expression{"columns": columnList(b.VectorMatching.Include...)}),
			// // Rename x_rhs -> x.
			// call("rename", map[string]ast.Expression{"fn": stripSuffixFn("_rhs")}),
			// // Drop any remaining RHS cols.
			// call("drop", map[string]ast.Expression{"fn": matchRHSSuffixFn}),
		}

		postJoinCalls := append(opCalls, outputColTransformCalls...)
		if dropMeasurement {
			postJoinCalls = append(postJoinCalls, dropMeasurementCall)
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
	default:
		return nil, fmt.Errorf("PromQL function %q is not supported yet", c.Func.Name)
	}
}

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

func (t *transpiler) transpileExpr(node promql.Node) (ast.Expression, error) {
	switch n := node.(type) {
	case *promql.ParenExpr:
		return t.transpileExpr(n.Expr)
	case *promql.UnaryExpr:
		return t.transpileUnaryExpr(n)
	case *promql.NumberLiteral:
		return &ast.FloatLiteral{Value: n.Val}, nil
	case *promql.StringLiteral:
		return &ast.StringLiteral{Value: n.Val}, nil
	case *promql.VectorSelector:
		return t.transpileInstantVectorSelector(n), nil
	case *promql.MatrixSelector:
		return t.transpileRangeVectorSelector(n), nil
	case *promql.AggregateExpr:
		return t.transpileAggregateExpr(n)
	case *promql.BinaryExpr:
		return t.transpileBinaryExpr(n)
	case *promql.Call:
		return t.transpileCall(n)
	default:
		return nil, fmt.Errorf("PromQL node type %T is not supported yet", t)
	}
}

func (t *transpiler) transpile(node promql.Node) (*ast.File, error) {
	fluxNode, err := t.transpileExpr(node)
	if err != nil {
		return nil, fmt.Errorf("error transpiling expression: %s", err)
	}
	return &ast.File{
		Imports: []*ast.ImportDeclaration{
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
