package universe_test

import (
	"context"
	"testing"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/dependencies/dependenciestest"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/stdlib/universe"
	"github.com/influxdata/flux/values"
	"github.com/influxdata/flux/values/valuestest"
)

func TestReduce_Process(t *testing.T) {
	testCases := []struct {
		name    string
		spec    *universe.ReduceProcedureSpec
		data    []flux.Table
		want    []*executetest.Table
		wantErr error
	}{
		{
			name: `sum _value`,
			spec: &universe.ReduceProcedureSpec{
				Identity: values.NewObjectWithValues(map[string]values.Value{
					"sum": values.NewFloat(0.0),
				}),
				Fn: interpreter.ResolvedFunction{
					Fn:    executetest.FunctionExpression(t, `(r, accumulator) => ({sum: r._value + accumulator.sum})`),
					Scope: valuestest.Scope(),
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 1.0},
					{execute.Time(2), 6.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "sum", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{7.0},
				},
			}},
		},
		{
			name: `sum+prod _value`,
			spec: &universe.ReduceProcedureSpec{
				Identity: values.NewObjectWithValues(map[string]values.Value{
					"sum":  values.NewFloat(0.0),
					"prod": values.NewFloat(1.0),
				}),
				Fn: interpreter.ResolvedFunction{
					Fn:    executetest.FunctionExpression(t, `(r, accumulator) => ({sum: r._value + accumulator.sum, prod: r._value * accumulator.prod})`),
					Scope: valuestest.Scope(),
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 4.1},
					{execute.Time(2), 6.2},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "prod", Type: flux.TFloat},
					{Label: "sum", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{25.419999999999998, 10.3},
				},
			}},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			executetest.ProcessTestHelper(
				t,
				tc.data,
				tc.want,
				tc.wantErr,
				func(d execute.Dataset, c execute.TableBuilderCache) execute.Transformation {
					ctx := dependenciestest.Default().Inject(context.Background())
					f, err := universe.NewReduceTransformation(ctx, tc.spec, d, c)
					if err != nil {
						t.Fatal(err)
					}
					return f
				},
			)
		})
	}
}
