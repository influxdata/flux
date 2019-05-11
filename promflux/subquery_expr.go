package promflux

import (
	"github.com/influxdata/flux/ast"
	"github.com/prometheus/prometheus/promql"
)

func (t *transpiler) transpileSubqueryExpr(a *promql.SubqueryExpr) (ast.Expression, error) {
	// 1. Create new transpiler with boundaries and step of subquery.
	sqt := &transpiler{
		bucket:     t.bucket,
		start:      t.start.Add(-a.Range - a.Offset),
		end:        t.end.Add(-a.Offset),
		resolution: a.Step,
	}

	// 2. Transpile subexpression with that transpiler.
	subquery, err := sqt.transpileExpr(a.Expr)
	if err != nil {
		return nil, err
	}

	// 3. Window the subexpression data according to the parent query's step and range.
	return buildPipeline(
		subquery,
		// The resolution step evaluation timestamp needs to become the output timestamp.
		call("duplicate", map[string]ast.Expression{
			"column": &ast.StringLiteral{Value: "_stop"},
			"as":     &ast.StringLiteral{Value: "_time"},
		}),
		// At every resolution step, include the specified range of data.
		call("window", map[string]ast.Expression{
			"every":  &ast.DurationLiteral{Values: []ast.Duration{{Magnitude: t.resolution.Nanoseconds(), Unit: "ns"}}},
			"period": &ast.DurationLiteral{Values: []ast.Duration{{Magnitude: a.Range.Nanoseconds(), Unit: "ns"}}},
		}),
		// Remove any windows smaller than the specified range at the edges of the graph range.
		call("filter", map[string]ast.Expression{"fn": windowCutoffFn(t.start.Add(-a.Offset), t.end.Add(-a.Range-a.Offset))}),
		// Apply offsets to make past data look like it's in the present.
		call("timeShift", map[string]ast.Expression{
			"duration": &ast.DurationLiteral{Values: []ast.Duration{{Magnitude: a.Offset.Nanoseconds(), Unit: "ns"}}},
		}),
	), nil
}
