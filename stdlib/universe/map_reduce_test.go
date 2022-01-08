package universe_test

import (
	"context"
	"testing"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/dependencies/dependenciestest"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/stdlib/universe"
	"github.com/influxdata/flux/values"
)

func TestMapReduce_Process(t *testing.T) {
	builtIns := runtime.Prelude()
	testCases := []struct {
		name    string
		spec    *universe.MapReduceProcedureSpec
		data    []flux.Table
		want    []*executetest.Table
		wantErr error
	}{
		{
			name: `overwrite groupkey`,
			spec: &universe.MapReduceProcedureSpec{
				Fn: interpreter.ResolvedFunction{
					Scope: builtIns,
					Fn: executetest.FunctionExpression(t, `
						(r, accumulator) => {
							newSum = accumulator.sum + r.value
							return ({
								row: { r with state: newSum },
								accumulator: { sum: newSum }
							})
						}
					`),
				},
				Identity: values.NewObjectWithValues(map[string]values.Value{"sum": values.NewFloat(0.0)}),
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 1.0},
					{execute.Time(2), 2.0},
					{execute.Time(3), 3.0},
					{execute.Time(4), 4.0},
					{execute.Time(5), 5.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "value", Type: flux.TFloat},
					{Label: "state", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 1.0, 1.0},
					{execute.Time(2), 2.0, 3.0},
					{execute.Time(3), 3.0, 6.0},
					{execute.Time(4), 4.0, 10.0},
					{execute.Time(5), 5.0, 15.0},
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
					f, err := universe.NewMapReduceTransformation(ctx, tc.spec, d, c)
					if err != nil {
						t.Fatal(err)
					}
					return f
				},
			)
		})
	}
}
