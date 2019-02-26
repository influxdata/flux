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

func buildPipeline(exprs ...*ast.CallExpression) *ast.PipeExpression {
	switch len(exprs) {
	case 0:
		panic("empty expression list")
	case 1:
		panic("less than two pipeline stages")
	case 2:
		return &ast.PipeExpression{
			Argument: exprs[0],
			Call:     exprs[1],
		}
	default:
		return &ast.PipeExpression{
			Argument: buildPipeline(exprs[0 : len(exprs)-1]...),
			Call:     exprs[len(exprs)-1],
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

func transpile(bucket string, n promql.Node, start time.Time, end time.Time, resolution time.Duration) (ast.Node, error) {
	switch t := n.(type) {
	case *promql.VectorSelector:
		return transpileInstantVectorSelector(bucket, t, start, end, resolution), nil
	default:
		return nil, fmt.Errorf("PromQL node type %T is not supported yet", t)
	}
}
