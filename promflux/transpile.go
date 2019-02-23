package main

import (
	"fmt"
	"time"

	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/ast/edit"
	"github.com/influxdata/flux/parser"
	"github.com/prometheus/common/model"
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

func transpile(bucket string, n promql.Node, start time.Time, end time.Time, resolution time.Duration) (ast.Node, error) {
	switch t := n.(type) {
	case *promql.VectorSelector:
		if len(t.LabelMatchers) != 1 || t.LabelMatchers[0].Name != model.MetricNameLabel {
			return nil, fmt.Errorf("vector selector label matchers not supported yet")
		}

		script := `
			option queryRangeStart = ""
			option queryRangeEnd = ""
			option queryResolution = ""
			option queryMetricName = ""
			option queryOffset = ""
			option queryWindowCutoff = ""

			from(bucket: "prom")
				// This is the query range with offsets already applied.
				|> range(start: queryRangeStart, stop: queryRangeEnd)

				// Apply metric (and later label) filters.
				|> filter(fn: (r) => r._measurement == queryMetricName)

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
			"queryRangeStart":   edit.OptionValueFn(&ast.DateTimeLiteral{Value: start.Add(-5*time.Minute - t.Offset)}),
			"queryRangeEnd":     edit.OptionValueFn(&ast.DateTimeLiteral{Value: end.Add(-t.Offset)}),
			"queryResolution":   edit.OptionValueFn(&ast.DurationLiteral{Values: []ast.Duration{{Magnitude: resolution.Nanoseconds(), Unit: "ns"}}}),
			"queryMetricName":   edit.OptionValueFn(&ast.StringLiteral{Value: t.Name}),
			"queryOffset":       edit.OptionValueFn(&ast.DurationLiteral{Values: []ast.Duration{{Magnitude: t.Offset.Nanoseconds(), Unit: "ns"}}}),
			"queryWindowCutoff": edit.OptionValueFn(&ast.DateTimeLiteral{Value: end.Add(-5*time.Minute - t.Offset)}),
		}

		mustEditOptions(p, opts)
		return p, nil

	default:
		return nil, fmt.Errorf("PromQL node type %T is not supported yet", t)
	}
}
