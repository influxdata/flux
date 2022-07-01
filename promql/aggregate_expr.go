package promql

import (
	"fmt"

	"github.com/influxdata/flux/ast"
	"github.com/influxdata/promql/v2"
	"github.com/prometheus/common/model"
)

// Taken from Prometheus.
const (
	// The largest SampleValue that can be converted to an int64 without overflow.
	maxInt64 = 9223372036854774784
	// The smallest SampleValue that can be converted to an int64 without underflow.
	minInt64 = -9223372036854775808
)

type aggregateFn struct {
	name      string
	dropField bool
	// All PromQL aggregation operators drop non-grouping labels, but some
	// of the (non-aggregation) Flux counterparts don't. This field indicates
	// that the non-grouping drop needs to be explicitly added to the pipeline.
	dropNonGrouping bool
}

var aggregateFns = map[promql.ItemType]aggregateFn{
	promql.ItemSum:      {name: "sum", dropField: true, dropNonGrouping: false},
	promql.ItemAvg:      {name: "mean", dropField: true, dropNonGrouping: false},
	promql.ItemMax:      {name: "max", dropField: true, dropNonGrouping: true},
	promql.ItemMin:      {name: "min", dropField: true, dropNonGrouping: true},
	promql.ItemCount:    {name: "count", dropField: true, dropNonGrouping: false},
	promql.ItemStddev:   {name: "stddev", dropField: true, dropNonGrouping: false},
	promql.ItemStdvar:   {name: "stddev", dropField: true, dropNonGrouping: false},
	promql.ItemTopK:     {name: "top", dropField: false, dropNonGrouping: false},
	promql.ItemBottomK:  {name: "bottom", dropField: false, dropNonGrouping: false},
	promql.ItemQuantile: {name: "quantile", dropField: true, dropNonGrouping: false},
}

func dropNonGroupingColsCall(groupCols []string, without bool) *ast.CallExpression {
	if without {
		cols := make([]string, 0, len(groupCols))
		// Remove "_value" and "_stop" from list of columns to drop.
		for _, col := range groupCols {
			if col != "_value" && col != "_stop" {
				cols = append(cols, col)
			}
		}

		return call("drop", map[string]ast.Expression{"columns": columnList(cols...)})
	}

	// We want to keep value and stop columns even if they are not explicitly in the grouping labels.
	cols := append(groupCols, "_value", "_stop")
	// TODO: This errors with non-existent columns. In PromQL, this is a no-op.
	// Blocked on https://github.com/influxdata/flux/issues/1118.
	return call("keep", map[string]ast.Expression{"columns": columnList(cols...)})
}

func (t *Transpiler) transpileAggregateExpr(a *promql.AggregateExpr) (ast.Expression, error) {
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

	switch a.Op {
	case promql.ItemTopK, promql.ItemBottomK:
		// TODO: Allow any constant scalars here.
		// The PromQL parser already verifies that a.Param is a scalar.
		n, ok := a.Param.(*promql.NumberLiteral)
		if !ok {
			return nil, fmt.Errorf("arbitrary scalar subexpressions not supported yet")
		}
		if n.Val > maxInt64 || n.Val < minInt64 {
			return nil, fmt.Errorf("scalar value %v overflows int64", n)
		}
		aggArgs["n"] = &ast.IntegerLiteral{Value: int64(n.Val)}

	case promql.ItemQuantile:
		// TODO: Allow any constant scalars here.
		// The PromQL parser already verifies that a.Param is a scalar.
		n, ok := a.Param.(*promql.NumberLiteral)
		if !ok {
			return nil, fmt.Errorf("arbitrary scalar subexpressions not supported yet")
		}
		aggArgs["q"] = &ast.FloatLiteral{Value: n.Val}
		aggArgs["method"] = &ast.StringLiteral{Value: "exact_mean"}

	case promql.ItemStddev, promql.ItemStdvar:
		aggArgs["mode"] = &ast.StringLiteral{Value: "population"}
	}

	mode := "by"
	dropField := true
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
				dropField = false
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
	if aggFn.name == "count" {
		pipeline = buildPipeline(pipeline, call("toFloat", nil))
	}
	if aggFn.dropNonGrouping {
		// Drop labels that are not part of the grouping.
		pipeline = buildPipeline(pipeline, dropNonGroupingColsCall(a.Grouping, a.Without))
	}
	if aggFn.dropField && dropField {
		pipeline = buildPipeline(
			pipeline,
			dropFieldAndTimeCall,
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
