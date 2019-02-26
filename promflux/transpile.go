package main

import (
	"fmt"
	"regexp"
	"time"

	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/ast/edit"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/promql"
)

func mustEditOptions(n ast.Node, options map[string]edit.OptionFn) {
	for opt, fn := range options {
		changed, err := edit.Option(n, opt, fn)
		if !changed {
			panic(fmt.Errorf("option %q not found", opt))
		}
		if err != nil {
			panic(fmt.Errorf("error editing option %q: %s", opt, err))
		}
	}
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
func windowCutoffFn(maxStart time.Time) *ast.FunctionExpression {
	return &ast.FunctionExpression{
		Params: []*ast.Property{
			{
				Key: &ast.Identifier{
					Name: "r",
				},
			},
		},
		Body: &ast.BinaryExpression{
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
	}
}

func transpileInstantVectorSelector(bucket string, v *promql.VectorSelector, start time.Time, end time.Time, resolution time.Duration) *ast.PipeExpression {
	return buildPipeline(
		// Select all Prometheus data.
		call("from", map[string]ast.Expression{"bucket": &ast.StringLiteral{Value: bucket}}),
		// Query entire graph range.
		call("range", map[string]ast.Expression{
			"start": &ast.DateTimeLiteral{Value: start.Add(-5*time.Minute - v.Offset)},
			"stop":  &ast.DateTimeLiteral{Value: end.Add(-v.Offset)},
		}),
		// Apply label matching filters.
		call("filter", map[string]ast.Expression{"fn": transpileLabelMatchersFn(v.LabelMatchers)}),
		// At every resolution step, load / look back up to 5m of data (PromQL lookback delta).
		call("window", map[string]ast.Expression{
			"every":  &ast.DurationLiteral{Values: []ast.Duration{{Magnitude: resolution.Nanoseconds(), Unit: "ns"}}},
			"period": &ast.DurationLiteral{Values: []ast.Duration{{Magnitude: 5, Unit: "m"}}},
		}),
		// Remove any windows <5m long to act like PromQL.
		call("filter", map[string]ast.Expression{"fn": windowCutoffFn(end.Add(-5*time.Minute - v.Offset))}),
		// Select the last data point after the current evaluation (resolution step) timestamp.
		call("last", nil),
		// The resolution step evaluation timestamp needs to become the output timestamp.
		call("drop", map[string]ast.Expression{"columns": &ast.ArrayExpression{
			Elements: []ast.Expression{&ast.StringLiteral{Value: "_time"}},
		}}),
		call("duplicate", map[string]ast.Expression{
			"column": &ast.StringLiteral{Value: "_stop"},
			"as":     &ast.StringLiteral{Value: "_time"},
		}),
		// Apply offsets to make past data look like it's in the present.
		call("shift", map[string]ast.Expression{
			"shift": &ast.DurationLiteral{Values: []ast.Duration{{Magnitude: v.Offset.Nanoseconds(), Unit: "ns"}}},
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
	dropNonGrouping bool
}

var aggregateFns = map[promql.ItemType]aggregateFn{
	promql.ItemSum:     {name: "sum", dropMeasurement: true, dropNonGrouping: false},
	promql.ItemAvg:     {name: "mean", dropMeasurement: true, dropNonGrouping: false},
	promql.ItemMax:     {name: "max", dropMeasurement: true, dropNonGrouping: true},
	promql.ItemMin:     {name: "min", dropMeasurement: true, dropNonGrouping: true},
	promql.ItemCount:   {name: "count", dropMeasurement: true, dropNonGrouping: false},
	promql.ItemStddev:  {name: "stddev", dropMeasurement: true, dropNonGrouping: false},
	promql.ItemTopK:    {name: "top", dropMeasurement: false, dropNonGrouping: false},
	promql.ItemBottomK: {name: "bottom", dropMeasurement: false, dropNonGrouping: false},
	// TODO: Flux does not yet have a quantile() aggregator.
	promql.ItemQuantile: {name: "quantile", dropMeasurement: true, dropNonGrouping: false},
}

func dropNonGroupingColsCall(groupCols []string, without bool) *ast.CallExpression {
	if without {
		cols := make([]string, len(groupCols)-1)
		// Remove "_value" from list of columns to drop.
		for _, col := range groupCols {
			if col != "_value" {
				cols = append(cols, col)
			}
		}

		// TODO: This errors with non-existent columns. In PromQL, this is a no-op.
		return call("drop", map[string]ast.Expression{"columns": columnList(groupCols...)})
	}

	// We want to keep "_value" and "_time" even if they are not explicitly in the grouping labels.
	cols := append(groupCols, "_value", "_time")
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

func transpileAggregateExpr(bucket string, a *promql.AggregateExpr, start time.Time, end time.Time, resolution time.Duration) (ast.Expression, error) {
	expr, err := transpile(bucket, a.Expr, start, end, resolution)
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
		// TODO: there's no Flux quantile() function yet.
		return nil, fmt.Errorf("quantile aggregator not supported yet")

		// The PromQL already verifies that a.Param is a scalar.
		n, ok := a.Param.(*promql.NumberLiteral)
		if !ok {
			return nil, fmt.Errorf("arbitrary scalar subexpressions not supported yet")
		}
		aggArgs["q"] = &ast.FloatLiteral{Value: n.Val}
	}

	mode := "by"
	dropMeasurement := true
	if a.Without {
		mode = "except"
		groupCols.Elements = append(groupCols.Elements, &ast.StringLiteral{Value: "_value"})
	} else {
		groupCols.Elements = append(groupCols.Elements, &ast.StringLiteral{Value: "_time"})
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
	)
	if aggFn.dropNonGrouping {
		// Drop labels that are not part of the grouping.
		pipeline = buildPipeline(pipeline, dropNonGroupingColsCall(a.Grouping, a.Without))
	}
	if aggFn.dropMeasurement && dropMeasurement {
		pipeline = buildPipeline(
			pipeline,
			call("drop", map[string]ast.Expression{
				"columns": &ast.ArrayExpression{Elements: []ast.Expression{&ast.StringLiteral{Value: "_measurement"}}},
			}),
		)
	}
	return pipeline, nil
}

func transpile(bucket string, n promql.Node, start time.Time, end time.Time, resolution time.Duration) (ast.Expression, error) {
	switch t := n.(type) {
	// case *promql.NumberLiteral:
	// 	// TODO: Do we need to keep the scalar timestamp?
	// 	return &ast.FloatLiteral{Value: t.Val}, nil
	case *promql.VectorSelector:
		return transpileInstantVectorSelector(bucket, t, start, end, resolution), nil
	case *promql.AggregateExpr:
		return transpileAggregateExpr(bucket, t, start, end, resolution)
	default:
		return nil, fmt.Errorf("PromQL node type %T is not supported yet", t)
	}
}
