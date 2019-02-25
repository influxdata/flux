package main

import (
	"fmt"
	"regexp"
	"time"

	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/ast/edit"
	"github.com/influxdata/flux/parser"
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

func transpile(bucket string, n promql.Node, start time.Time, end time.Time, resolution time.Duration) (ast.Node, error) {
	switch t := n.(type) {
	case *promql.VectorSelector:
		script := `
			option queryRangeStart = ""
			option queryRangeEnd = ""
			option queryResolution = ""
			option queryLabelMatchersFn = ""
			option queryOffset = ""
			option queryWindowCutoff = ""

			from(bucket: "prom")
				// This is the query range with offsets already applied.
				|> range(start: queryRangeStart, stop: queryRangeEnd)

				// Apply metric (and later label) filters.
				|> filter(fn: queryLabelMatchersFn)

				// At every resolution step, look back 5m (PromQL lookback delta).
				|> window(every: queryResolution, period: 5m)

				// Remove any windows <5m long to act like PromQL.
				|> filter(fn: (r) => r._start <= queryWindowCutoff)

				// Select the last data point after the current evaluation (resolution step) timestamp.
				|> last()

				// The resolution step evaluation timestamp needs to become the output timestamp.
				|> drop(columns: ["_time"])
				|> duplicate(column: "_stop", as: "_time")

				// Apply offsets to make past data look like it's in the present.
				|> shift(shift: queryOffset)
		`
		p := parser.ParseSource(script)
		if ast.Check(p) > 0 {
			return nil, fmt.Errorf("error parsing Flux script: %s", ast.GetError(p))
		}

		opts := map[string]edit.OptionFn{
			"queryRangeStart":      edit.OptionValueFn(&ast.DateTimeLiteral{Value: start.Add(-5*time.Minute - t.Offset)}),
			"queryRangeEnd":        edit.OptionValueFn(&ast.DateTimeLiteral{Value: end.Add(-t.Offset)}),
			"queryResolution":      edit.OptionValueFn(&ast.DurationLiteral{Values: []ast.Duration{{Magnitude: resolution.Nanoseconds(), Unit: "ns"}}}),
			"queryLabelMatchersFn": edit.OptionValueFn(transpileLabelMatchersFn(t.LabelMatchers)),
			"queryOffset":          edit.OptionValueFn(&ast.DurationLiteral{Values: []ast.Duration{{Magnitude: t.Offset.Nanoseconds(), Unit: "ns"}}}),
			"queryWindowCutoff":    edit.OptionValueFn(&ast.DateTimeLiteral{Value: end.Add(-5*time.Minute - t.Offset)}),
		}

		mustEditOptions(p, opts)
		return p, nil

	default:
		return nil, fmt.Errorf("PromQL node type %T is not supported yet", t)
	}
}
