package promflux

import (
	"github.com/influxdata/flux/ast"
	"github.com/prometheus/prometheus/promql"
)

func (t *Transpiler) transpileSubqueryExpr(sq *promql.SubqueryExpr) (ast.Expression, error) {
	// 1. Create new transpiler with boundaries and step of subquery.
	sqt := &Transpiler{
		Bucket:     t.Bucket,
		Start:      t.Start.Add(-sq.Range - sq.Offset),
		End:        t.End.Add(-sq.Offset),
		Resolution: sq.Step,
	}

	// 2. Transpile subexpression with that transpiler.
	subquery, err := sqt.transpileExpr(sq.Expr)
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
			"every":  &ast.DurationLiteral{Values: []ast.Duration{{Magnitude: t.Resolution.Nanoseconds(), Unit: "ns"}}},
			"period": &ast.DurationLiteral{Values: []ast.Duration{{Magnitude: sq.Range.Nanoseconds(), Unit: "ns"}}},
		}),
		// Remove any windows smaller than the specified range at the edges of the graph range.
		call("filter", map[string]ast.Expression{"fn": windowCutoffFn(t.Start.Add(-sq.Offset), t.End.Add(-sq.Range-sq.Offset))}),
		// Apply offsets to make past data look like it's in the present.
		call("timeShift", map[string]ast.Expression{
			"duration": &ast.DurationLiteral{Values: []ast.Duration{{Magnitude: sq.Offset.Nanoseconds(), Unit: "ns"}}},
		}),
	), nil
}
