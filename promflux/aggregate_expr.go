package promflux

import (
	"fmt"

	"github.com/influxdata/flux/ast"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/promql"
)

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
			if col != "_value" && col != "_stop" { // TODO: Handle this systematically instead!
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
		// TODO: Allow any constant scalars here.
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
		// TODO: Allow any constant scalars here.
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
