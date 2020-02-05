package universe_test

import (
	"context"
	"testing"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/dependencies/dependenciestest"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/stdlib/universe"
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
				Identity: map[string]string{"sum": "0.0"},
				ReducerType: semantic.NewObjectType([]semantic.PropertyType{{
					Key:   []byte("sum"),
					Value: semantic.BasicFloat,
				}}),
				Fn: interpreter.ResolvedFunction{
					Fn: &semantic.FunctionExpression{
						Block: &semantic.FunctionBlock{
							Parameters: &semantic.FunctionParameters{
								List: []*semantic.FunctionParameter{{Key: &semantic.Identifier{Name: "r"}}, {Key: &semantic.Identifier{Name: "accumulator"}}},
							},
							Body: &semantic.ObjectExpression{
								Properties: []*semantic.Property{
									{
										Key: &semantic.Identifier{Name: "sum"},
										Value: &semantic.BinaryExpression{
											Operator: ast.AdditionOperator,
											Left: &semantic.MemberExpression{
												Object: &semantic.IdentifierExpression{
													Name: "r",
												},
												Property: "_value",
											},
											Right: &semantic.MemberExpression{
												Object: &semantic.IdentifierExpression{
													Name: "accumulator",
												},
												Property: "sum",
											},
										},
									},
								},
							},
						},
					},
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
				Identity: map[string]string{"sum": "0.0", "prod": "1.0"},
				ReducerType: semantic.NewObjectType([]semantic.PropertyType{
					{
						Key:   []byte("sum"),
						Value: semantic.BasicFloat,
					},
					{
						Key:   []byte("prod"),
						Value: semantic.BasicFloat,
					},
				}),
				Fn: interpreter.ResolvedFunction{
					Fn: &semantic.FunctionExpression{
						Block: &semantic.FunctionBlock{
							Parameters: &semantic.FunctionParameters{
								List: []*semantic.FunctionParameter{{Key: &semantic.Identifier{Name: "r"}}, {Key: &semantic.Identifier{Name: "accumulator"}}},
							},
							Body: &semantic.ObjectExpression{
								Properties: []*semantic.Property{
									{
										Key: &semantic.Identifier{Name: "sum"},
										Value: &semantic.BinaryExpression{
											Operator: ast.AdditionOperator,
											Left: &semantic.MemberExpression{
												Object: &semantic.IdentifierExpression{
													Name: "r",
												},
												Property: "_value",
											},
											Right: &semantic.MemberExpression{
												Object: &semantic.IdentifierExpression{
													Name: "accumulator",
												},
												Property: "sum",
											},
										},
									},
									{
										Key: &semantic.Identifier{Name: "prod"},
										Value: &semantic.BinaryExpression{
											Operator: ast.MultiplicationOperator,
											Left: &semantic.MemberExpression{
												Object: &semantic.IdentifierExpression{
													Name: "r",
												},
												Property: "_value",
											},
											Right: &semantic.MemberExpression{
												Object: &semantic.IdentifierExpression{
													Name: "accumulator",
												},
												Property: "prod",
											},
										},
									},
								},
							},
						},
					},
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
