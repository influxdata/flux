package promql

import (
	"fmt"
	"regexp"
	"time"

	"github.com/influxdata/flux/ast"
	"github.com/influxdata/promql/v2"
	"github.com/influxdata/promql/v2/pkg/labels"
)

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
	be := &ast.BinaryExpression{
		Operator: op,
		Left:     member("r", lm.Name),
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

func (t *Transpiler) transpileInstantVectorSelector(v *promql.VectorSelector) *ast.PipeExpression {
	var windowCall *ast.CallExpression
	var windowFilterCall *ast.CallExpression
	if t.Resolution > 0 {
		// For range queries:
		// At every resolution step, load / look back up to 5m of data (PromQL lookback delta).
		windowCall = call("window", map[string]ast.Expression{
			"every":  &ast.DurationLiteral{Values: []ast.Duration{{Magnitude: t.Resolution.Nanoseconds(), Unit: "ns"}}},
			"period": &ast.DurationLiteral{Values: []ast.Duration{{Magnitude: 5, Unit: "m"}}},
			"offset": &ast.DurationLiteral{Values: []ast.Duration{{Magnitude: t.Start.Add(-v.Offset).UnixNano() % t.Resolution.Nanoseconds(), Unit: "ns"}}},
		})

		// Remove any windows <5m long at the edges of the graph range to act like PromQL.
		windowFilterCall = call("filter", map[string]ast.Expression{"fn": windowCutoffFn(t.Start.Add(-v.Offset), t.End.Add(-5*time.Minute-v.Offset))})
	}

	return buildPipeline(
		// Select all Prometheus data.
		call("from", map[string]ast.Expression{"bucket": &ast.StringLiteral{Value: t.Bucket}}),
		// Query entire graph range.
		call("range", map[string]ast.Expression{
			"start": &ast.DateTimeLiteral{Value: t.Start.Add(-5*time.Minute - v.Offset)},
			"stop":  &ast.DateTimeLiteral{Value: t.End.Add(-v.Offset)},
		}),
		// Apply label matching filters.
		call("filter", map[string]ast.Expression{"fn": transpileLabelMatchersFn(v.LabelMatchers)}),
		windowCall,
		windowFilterCall,
		// Select the last data point after the current evaluation (resolution step) timestamp.
		call("last", nil),
		// Apply offsets to make past data look like it's in the present.
		call("timeShift", map[string]ast.Expression{
			"duration": &ast.DurationLiteral{Values: []ast.Duration{{Magnitude: v.Offset.Nanoseconds(), Unit: "ns"}}},
		}),
		dropMeasurementCall,
	)
}

func (t *Transpiler) transpileRangeVectorSelector(v *promql.MatrixSelector) *ast.PipeExpression {
	var windowCall *ast.CallExpression
	var windowFilterCall *ast.CallExpression
	if t.Resolution > 0 {
		// For range queries:
		// At every resolution step, include the specified range of data.
		windowCall = call("window", map[string]ast.Expression{
			"every":  &ast.DurationLiteral{Values: []ast.Duration{{Magnitude: t.Resolution.Nanoseconds(), Unit: "ns"}}},
			"period": &ast.DurationLiteral{Values: []ast.Duration{{Magnitude: v.Range.Nanoseconds(), Unit: "ns"}}},
			"offset": &ast.DurationLiteral{Values: []ast.Duration{{Magnitude: t.Start.UnixNano() % t.Resolution.Nanoseconds(), Unit: "ns"}}},
		})

		// Remove any windows smaller than the specified range at the edges of the graph range.
		windowFilterCall = call("filter", map[string]ast.Expression{"fn": windowCutoffFn(t.Start.Add(-v.Offset), t.End.Add(-v.Range-v.Offset))})
	}

	return buildPipeline(
		// Select all Prometheus data.
		call("from", map[string]ast.Expression{"bucket": &ast.StringLiteral{Value: t.Bucket}}),
		// Query entire graph range.
		call("range", map[string]ast.Expression{
			"start": &ast.DateTimeLiteral{Value: t.Start.Add(-v.Range - v.Offset)},
			"stop":  &ast.DateTimeLiteral{Value: t.End.Add(-v.Offset)},
		}),
		// Apply label matching filters.
		call("filter", map[string]ast.Expression{"fn": transpileLabelMatchersFn(v.LabelMatchers)}),
		windowCall,
		windowFilterCall,
		// Apply offsets to make past data look like it's in the present.
		call("timeShift", map[string]ast.Expression{
			"duration": &ast.DurationLiteral{Values: []ast.Duration{{Magnitude: v.Offset.Nanoseconds(), Unit: "ns"}}},
		}),
		dropMeasurementCall,
	)
}
