package main

import (
	"fmt"
	"time"

	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/ast/edit"
	"github.com/influxdata/flux/parser"
	"github.com/prometheus/prometheus/promql"
)

func editOptions(n ast.Node, options map[string]edit.OptionFn) error {
	for opt, fn := range options {
		changed, err := edit.Option(n, opt, fn)
		if !changed {
			return fmt.Errorf("option %q not found", opt)
		}
		if err != nil {
			return fmt.Errorf("error editing option %q: %s", opt, err)
		}
	}
	return nil
}

func transpile(bucket string, n promql.Node, start time.Time, end time.Time, resolution time.Duration) (ast.Node, error) {
	switch t := n.(type) {
	case *promql.VectorSelector:
		script := `
			option queryRangeStart = ""
			option queryRangeEnd = ""
			option queryResolution = ""
			option queryMetricName = ""
			option queryOffset = ""

		  from(bucket: "prom")
			|> range(start: queryRangeStart, stop: queryRangeEnd)
			|> filter(fn: (r) => r._measurement == queryMetricName)
			|> window(every: queryResolution, period: 5m)
			|> last()
			|> drop(columns: ["_time"])
			|> duplicate(column: "_stop", as: "_time")
			|> shift(shift: queryOffset)
		`
		p := parser.ParseSource(script)
		if ast.Check(p) > 0 {
			return nil, fmt.Errorf("error parsing Flux script: %s", ast.GetError(p))
		}

		opts := map[string]edit.OptionFn{
			"queryRangeStart": edit.OptionValueFn(&ast.DateTimeLiteral{Value: start.Add(-5*time.Minute - t.Offset)}),
			"queryRangeEnd":   edit.OptionValueFn(&ast.DateTimeLiteral{Value: end.Add(-t.Offset)}),
			"queryResolution": edit.OptionValueFn(&ast.DurationLiteral{Values: []ast.Duration{{Magnitude: resolution.Nanoseconds(), Unit: "ns"}}}),
			"queryMetricName": edit.OptionValueFn(&ast.StringLiteral{Value: t.Name}),
			"queryOffset":     edit.OptionValueFn(&ast.DurationLiteral{Values: []ast.Duration{{Magnitude: t.Offset.Nanoseconds(), Unit: "ns"}}}),
		}

		if err := editOptions(p, opts); err != nil {
			return nil, err
		}

		return p, nil

	default:
		return nil, fmt.Errorf("PromQL node type %T is not supported yet", t)
	}
}
