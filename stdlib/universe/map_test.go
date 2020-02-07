package universe_test

import (
	"context"
	"errors"
	"testing"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/dependencies/dependenciestest"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/querytest"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/stdlib/influxdata/influxdb"
	"github.com/influxdata/flux/stdlib/universe"
	"github.com/influxdata/flux/values"
	"github.com/influxdata/flux/values/valuestest"
)

func TestMap_NewQuery(t *testing.T) {
	tests := []querytest.NewQueryTestCase{
		{
			Name: "simple static map",
			Raw:  `from(bucket:"mybucket") |> map(fn: (r) => r._value + 1)`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: "mybucket",
						},
					},
					{
						ID: "map1",
						Spec: &universe.MapOpSpec{
							Fn: interpreter.ResolvedFunction{
								Fn: &semantic.FunctionExpression{
									Block: &semantic.FunctionBlock{
										Parameters: &semantic.FunctionParameters{
											List: []*semantic.FunctionParameter{{Key: &semantic.Identifier{Name: "r"}}},
										},
										Body: &semantic.BinaryExpression{
											Operator: ast.AdditionOperator,
											Left: &semantic.MemberExpression{
												Object: &semantic.IdentifierExpression{
													Name: "r",
												},
												Property: "_value",
											},
											Right: &semantic.IntegerLiteral{Value: 1},
										},
									},
								},
								Scope: valuestest.Scope(),
							},
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "map1"},
				},
			},
		},
		{
			Name: "simple static map mergeKey=true",
			Raw:  `from(bucket:"mybucket") |> map(fn: (r) => r._value + 1, mergeKey:true)`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: "mybucket",
						},
					},
					{
						ID: "map1",
						Spec: &universe.MapOpSpec{
							MergeKey: true,
							Fn: interpreter.ResolvedFunction{
								Fn: &semantic.FunctionExpression{
									Block: &semantic.FunctionBlock{
										Parameters: &semantic.FunctionParameters{
											List: []*semantic.FunctionParameter{{Key: &semantic.Identifier{Name: "r"}}},
										},
										Body: &semantic.BinaryExpression{
											Operator: ast.AdditionOperator,
											Left: &semantic.MemberExpression{
												Object: &semantic.IdentifierExpression{
													Name: "r",
												},
												Property: "_value",
											},
											Right: &semantic.IntegerLiteral{Value: 1},
										},
									},
								},
								Scope: valuestest.Scope(),
							},
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "map1"},
				},
			},
		},
		{
			Name: "resolve map",
			Raw:  `x = 2 from(bucket:"mybucket") |> map(fn: (r) => r._value + x)`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: "mybucket",
						},
					},
					{
						ID: "map1",
						Spec: &universe.MapOpSpec{
							Fn: interpreter.ResolvedFunction{
								Fn: &semantic.FunctionExpression{
									Block: &semantic.FunctionBlock{
										Parameters: &semantic.FunctionParameters{
											List: []*semantic.FunctionParameter{{Key: &semantic.Identifier{Name: "r"}}},
										},
										Body: &semantic.BinaryExpression{
											Operator: ast.AdditionOperator,
											Left: &semantic.MemberExpression{
												Object: &semantic.IdentifierExpression{
													Name: "r",
												},
												Property: "_value",
											},
											Right: &semantic.IntegerLiteral{Value: 2},
										},
									},
								},
								Scope: func() values.Scope {
									scope := valuestest.Scope()
									scope.Set("x", values.NewInt(2))
									return scope
								}(),
							},
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "map1"},
				},
			},
		},
	}
	for _, tc := range tests {
		t.Skip("https://github.com/influxdata/flux/issues/2494")
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			querytest.NewQueryTestHelper(t, tc)
		})
	}
}

func TestMapOperation_Marshaling(t *testing.T) {
	t.Skip("https://github.com/influxdata/flux/issues/2492")
	data := []byte(`{
		"id":"map",
		"kind":"map",
		"spec":{
			"fn":{
				"fn":{
					"type": "FunctionExpression",
					"block":{
						"type":"FunctionBlock",
						"parameters": {
							"type":"FunctionParameters",
							"list": [
								{"type":"FunctionParam","key":{"type":"Identifier","name":"r"}}
							]
						},
						"body":{
							"type":"BinaryExpression",
							"operator": "-",
							"left":{
								"type":"MemberExpression",
								"object": {
									"type": "IdentifierExpression",
									"name":"r"
								},
								"property": "_value"
							},
							"right":{
								"type":"FloatLiteral",
								"value": 5.6
							}
						}
					}
				}
			}
		}
	}`)
	op := &flux.Operation{
		ID: "map",
		Spec: &universe.MapOpSpec{
			Fn: interpreter.ResolvedFunction{
				Fn: &semantic.FunctionExpression{
					Block: &semantic.FunctionBlock{
						Parameters: &semantic.FunctionParameters{
							List: []*semantic.FunctionParameter{{Key: &semantic.Identifier{Name: "r"}}},
						},
						Body: &semantic.BinaryExpression{
							Operator: ast.SubtractionOperator,
							Left: &semantic.MemberExpression{
								Object: &semantic.IdentifierExpression{
									Name: "r",
								},
								Property: "_value",
							},
							Right: &semantic.FloatLiteral{Value: 5.6},
						},
					},
				},
				Scope: nil,
			},
		},
	}
	querytest.OperationMarshalingTestHelper(t, data, op)
}
func TestMap_Process(t *testing.T) {
	builtIns := flux.Prelude()
	testCases := []struct {
		name    string
		spec    *universe.MapProcedureSpec
		data    []flux.Table
		want    []*executetest.Table
		wantErr error
	}{
		{
			name: `_value+5`,
			spec: &universe.MapProcedureSpec{
				Fn: interpreter.ResolvedFunction{
					Scope: builtIns,
					Fn: &semantic.FunctionExpression{
						Block: &semantic.FunctionBlock{
							Parameters: &semantic.FunctionParameters{
								List: []*semantic.FunctionParameter{{Key: &semantic.Identifier{Name: "r"}}},
							},
							Body: &semantic.ObjectExpression{
								Properties: []*semantic.Property{
									{
										Key: &semantic.Identifier{Name: "_time"},
										Value: &semantic.MemberExpression{
											Object: &semantic.IdentifierExpression{
												Name: "r",
											},
											Property: "_time",
										},
									},
									{
										Key: &semantic.Identifier{Name: "_value"},
										Value: &semantic.BinaryExpression{
											Operator: ast.AdditionOperator,
											Left: &semantic.MemberExpression{
												Object: &semantic.IdentifierExpression{
													Name: "r",
												},
												Property: "_value",
											},
											Right: &semantic.FloatLiteral{
												Value: 5,
											},
										},
									},
								},
							},
						},
					},
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
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 6.0},
					{execute.Time(2), 11.0},
				},
			}},
		},
		{
			name: `_value+5 custom scope`,
			spec: &universe.MapProcedureSpec{
				Fn: interpreter.ResolvedFunction{
					Scope: func() values.Scope {
						scope := builtIns.Nest(nil)
						scope.Set("x", values.NewFloat(5))
						return scope
					}(),
					Fn: &semantic.FunctionExpression{
						Block: &semantic.FunctionBlock{
							Parameters: &semantic.FunctionParameters{
								List: []*semantic.FunctionParameter{{Key: &semantic.Identifier{Name: "r"}}},
							},
							Body: &semantic.ObjectExpression{
								Properties: []*semantic.Property{
									{
										Key: &semantic.Identifier{Name: "_time"},
										Value: &semantic.MemberExpression{
											Object: &semantic.IdentifierExpression{
												Name: "r",
											},
											Property: "_time",
										},
									},
									{
										Key: &semantic.Identifier{Name: "_value"},
										Value: &semantic.BinaryExpression{
											Operator: ast.AdditionOperator,
											Left: &semantic.MemberExpression{
												Object: &semantic.IdentifierExpression{
													Name: "r",
												},
												Property: "_value",
											},
											Right: &semantic.IdentifierExpression{Name: "x"},
										},
									},
								},
							},
						},
					},
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
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 6.0},
					{execute.Time(2), 11.0},
				},
			}},
		},
		{
			name: `_value+5 drop cols`,
			spec: &universe.MapProcedureSpec{
				Fn: interpreter.ResolvedFunction{
					Scope: builtIns,
					Fn: &semantic.FunctionExpression{
						Block: &semantic.FunctionBlock{
							Parameters: &semantic.FunctionParameters{
								List: []*semantic.FunctionParameter{{Key: &semantic.Identifier{Name: "r"}}},
							},
							Body: &semantic.ObjectExpression{
								Properties: []*semantic.Property{
									{
										Key: &semantic.Identifier{Name: "_time"},
										Value: &semantic.MemberExpression{
											Object: &semantic.IdentifierExpression{
												Name: "r",
											},
											Property: "_time",
										},
									},
									{
										Key: &semantic.Identifier{Name: "_value"},
										Value: &semantic.BinaryExpression{
											Operator: ast.AdditionOperator,
											Left: &semantic.MemberExpression{
												Object: &semantic.IdentifierExpression{
													Name: "r",
												},
												Property: "_value",
											},
											Right: &semantic.FloatLiteral{
												Value: 5,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"_measurement", "host"},
				ColMeta: []flux.ColMeta{
					{Label: "_measurement", Type: flux.TString},
					{Label: "host", Type: flux.TString},
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{"m", "A", execute.Time(1), 1.0},
					{"m", "A", execute.Time(2), 6.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 6.0},
					{execute.Time(2), 11.0},
				},
			}},
		},
		{
			name: `_value+5 drop cols regroup`,
			spec: &universe.MapProcedureSpec{
				Fn: interpreter.ResolvedFunction{
					Scope: builtIns,
					Fn: &semantic.FunctionExpression{
						Block: &semantic.FunctionBlock{
							Parameters: &semantic.FunctionParameters{
								List: []*semantic.FunctionParameter{{Key: &semantic.Identifier{Name: "r"}}},
							},
							Body: &semantic.ObjectExpression{
								Properties: []*semantic.Property{
									{
										Key: &semantic.Identifier{Name: "_time"},
										Value: &semantic.MemberExpression{
											Object: &semantic.IdentifierExpression{
												Name: "r",
											},
											Property: "_time",
										},
									},
									{
										Key: &semantic.Identifier{Name: "_value"},
										Value: &semantic.BinaryExpression{
											Operator: ast.AdditionOperator,
											Left: &semantic.MemberExpression{
												Object: &semantic.IdentifierExpression{
													Name: "r",
												},
												Property: "_value",
											},
											Right: &semantic.FloatLiteral{
												Value: 5,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			data: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"_measurement", "host"},
					ColMeta: []flux.ColMeta{
						{Label: "_measurement", Type: flux.TString},
						{Label: "host", Type: flux.TString},
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{"m", "A", execute.Time(1), 1.0},
						{"m", "A", execute.Time(2), 6.0},
					},
				},
				&executetest.Table{
					KeyCols: []string{"_measurement", "host"},
					ColMeta: []flux.ColMeta{
						{Label: "_measurement", Type: flux.TString},
						{Label: "host", Type: flux.TString},
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{"m", "B", execute.Time(1), 1.5},
						{"m", "B", execute.Time(2), 6.5},
					},
				},
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 6.0},
					{execute.Time(2), 11.0},
					{execute.Time(1), 6.5},
					{execute.Time(2), 11.5},
				},
			}},
		},
		{
			name: `with _value+5 regroup`,
			spec: &universe.MapProcedureSpec{
				Fn: interpreter.ResolvedFunction{
					Scope: builtIns,
					Fn: &semantic.FunctionExpression{
						Block: &semantic.FunctionBlock{
							Parameters: &semantic.FunctionParameters{
								List: []*semantic.FunctionParameter{{Key: &semantic.Identifier{Name: "r"}}},
							},
							Body: &semantic.ObjectExpression{
								With: &semantic.IdentifierExpression{Name: "r"},
								Properties: []*semantic.Property{
									{
										Key: &semantic.Identifier{Name: "host"},
										Value: &semantic.BinaryExpression{
											Operator: ast.AdditionOperator,
											Left: &semantic.MemberExpression{
												Object: &semantic.IdentifierExpression{
													Name: "r",
												},
												Property: "host",
											},
											Right: &semantic.StringLiteral{Value: ".local"},
										},
									},
									{
										Key: &semantic.Identifier{Name: "_value"},
										Value: &semantic.BinaryExpression{
											Operator: ast.AdditionOperator,
											Left: &semantic.MemberExpression{
												Object: &semantic.IdentifierExpression{
													Name: "r",
												},
												Property: "_value",
											},
											Right: &semantic.FloatLiteral{
												Value: 5,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"_measurement", "host"},
				ColMeta: []flux.ColMeta{
					{Label: "_measurement", Type: flux.TString},
					{Label: "host", Type: flux.TString},
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{"m", "A", execute.Time(1), 1.0},
					{"m", "A", execute.Time(2), 6.0},
				},
			}},
			want: []*executetest.Table{{
				KeyCols: []string{"_measurement", "host"},
				ColMeta: []flux.ColMeta{
					{Label: "_measurement", Type: flux.TString},
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "host", Type: flux.TString},
				},
				Data: [][]interface{}{
					{"m", execute.Time(1), 6.0, "A.local"},
					{"m", execute.Time(2), 11.0, "A.local"},
				},
			}},
		},
		{
			name: `with _value+5 regroup fan out`,
			spec: &universe.MapProcedureSpec{
				Fn: interpreter.ResolvedFunction{
					Scope: builtIns,
					Fn: &semantic.FunctionExpression{
						Block: &semantic.FunctionBlock{
							Parameters: &semantic.FunctionParameters{
								List: []*semantic.FunctionParameter{{Key: &semantic.Identifier{Name: "r"}}},
							},
							Body: &semantic.ObjectExpression{
								With: &semantic.IdentifierExpression{Name: "r"},
								Properties: []*semantic.Property{
									{
										Key: &semantic.Identifier{Name: "host"},
										Value: &semantic.BinaryExpression{
											Operator: ast.AdditionOperator,
											Left: &semantic.MemberExpression{
												Object: &semantic.IdentifierExpression{
													Name: "r",
												},
												Property: "host",
											},
											Right: &semantic.BinaryExpression{
												Operator: ast.AdditionOperator,
												Left:     &semantic.StringLiteral{Value: "."},
												Right: &semantic.MemberExpression{
													Object: &semantic.IdentifierExpression{
														Name: "r",
													},
													Property: "domain",
												},
											},
										},
									},
									{
										Key: &semantic.Identifier{Name: "_value"},
										Value: &semantic.BinaryExpression{
											Operator: ast.AdditionOperator,
											Left: &semantic.MemberExpression{
												Object: &semantic.IdentifierExpression{
													Name: "r",
												},
												Property: "_value",
											},
											Right: &semantic.FloatLiteral{
												Value: 5,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"_measurement", "host"},
				ColMeta: []flux.ColMeta{
					{Label: "_measurement", Type: flux.TString},
					{Label: "host", Type: flux.TString},
					{Label: "domain", Type: flux.TString},
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{"m", "A", "example.com", execute.Time(1), 1.0},
					{"m", "A", "www.example.com", execute.Time(2), 6.0},
				},
			}},
			want: []*executetest.Table{
				{
					KeyCols: []string{"_measurement", "host"},
					ColMeta: []flux.ColMeta{
						{Label: "_measurement", Type: flux.TString},
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "domain", Type: flux.TString},
						{Label: "host", Type: flux.TString},
					},
					Data: [][]interface{}{
						{"m", execute.Time(1), 6.0, "example.com", "A.example.com"},
					},
				},
				{
					KeyCols: []string{"_measurement", "host"},
					ColMeta: []flux.ColMeta{
						{Label: "_measurement", Type: flux.TString},
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "domain", Type: flux.TString},
						{Label: "host", Type: flux.TString},
					},
					Data: [][]interface{}{
						{"m", execute.Time(2), 11.0, "www.example.com", "A.www.example.com"},
					},
				},
			},
		},
		{
			name: `with _value+5 with nulls`,
			spec: &universe.MapProcedureSpec{
				Fn: interpreter.ResolvedFunction{
					Scope: builtIns,
					Fn: &semantic.FunctionExpression{
						Block: &semantic.FunctionBlock{
							Parameters: &semantic.FunctionParameters{
								List: []*semantic.FunctionParameter{{Key: &semantic.Identifier{Name: "r"}}},
							},
							Body: &semantic.ObjectExpression{
								With: &semantic.IdentifierExpression{Name: "r"},
								Properties: []*semantic.Property{
									{
										Key: &semantic.Identifier{Name: "_value"},
										Value: &semantic.BinaryExpression{
											Operator: ast.AdditionOperator,
											Left: &semantic.MemberExpression{
												Object: &semantic.IdentifierExpression{
													Name: "r",
												},
												Property: "_value",
											},
											Right: &semantic.FloatLiteral{
												Value: 5,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"_measurement", "host"},
				ColMeta: []flux.ColMeta{
					{Label: "_measurement", Type: flux.TString},
					{Label: "host", Type: flux.TString},
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{"m", nil, execute.Time(1), 1.0},
					{"m", nil, execute.Time(2), 6.0},
				},
			}},
			want: []*executetest.Table{{
				KeyCols: []string{"_measurement", "host"},
				ColMeta: []flux.ColMeta{
					{Label: "_measurement", Type: flux.TString},
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "host", Type: flux.TString},
				},
				Data: [][]interface{}{
					{"m", execute.Time(1), 6.0, nil},
					{"m", execute.Time(2), 11.0, nil},
				},
			}},
		},
		{
			name: `_value+5 mergeKey=true regroup`,
			spec: &universe.MapProcedureSpec{
				MergeKey: true,
				Fn: interpreter.ResolvedFunction{
					Scope: builtIns,
					Fn: &semantic.FunctionExpression{
						Block: &semantic.FunctionBlock{
							Parameters: &semantic.FunctionParameters{
								List: []*semantic.FunctionParameter{{Key: &semantic.Identifier{Name: "r"}}},
							},
							Body: &semantic.ObjectExpression{
								Properties: []*semantic.Property{
									{
										Key: &semantic.Identifier{Name: "_time"},
										Value: &semantic.MemberExpression{
											Object: &semantic.IdentifierExpression{
												Name: "r",
											},
											Property: "_time",
										},
									},
									{
										Key: &semantic.Identifier{Name: "host"},
										Value: &semantic.BinaryExpression{
											Operator: ast.AdditionOperator,
											Left: &semantic.MemberExpression{
												Object: &semantic.IdentifierExpression{
													Name: "r",
												},
												Property: "host",
											},
											Right: &semantic.StringLiteral{Value: ".local"},
										},
									},
									{
										Key: &semantic.Identifier{Name: "_value"},
										Value: &semantic.BinaryExpression{
											Operator: ast.AdditionOperator,
											Left: &semantic.MemberExpression{
												Object: &semantic.IdentifierExpression{
													Name: "r",
												},
												Property: "_value",
											},
											Right: &semantic.FloatLiteral{
												Value: 5,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"_measurement", "host"},
				ColMeta: []flux.ColMeta{
					{Label: "_measurement", Type: flux.TString},
					{Label: "host", Type: flux.TString},
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{"m", "A", execute.Time(1), 1.0},
					{"m", "A", execute.Time(2), 6.0},
				},
			}},
			want: []*executetest.Table{{
				KeyCols: []string{"_measurement", "host"},
				ColMeta: []flux.ColMeta{
					{Label: "_measurement", Type: flux.TString},
					{Label: "host", Type: flux.TString},
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{"m", "A.local", execute.Time(1), 6.0},
					{"m", "A.local", execute.Time(2), 11.0},
				},
			}},
		},
		{
			name: `_value+5 mergeKey=true regroup fan out`,
			spec: &universe.MapProcedureSpec{
				MergeKey: true,
				Fn: interpreter.ResolvedFunction{
					Scope: builtIns,
					Fn: &semantic.FunctionExpression{
						Block: &semantic.FunctionBlock{
							Parameters: &semantic.FunctionParameters{
								List: []*semantic.FunctionParameter{{Key: &semantic.Identifier{Name: "r"}}},
							},
							Body: &semantic.ObjectExpression{
								Properties: []*semantic.Property{
									{
										Key: &semantic.Identifier{Name: "_time"},
										Value: &semantic.MemberExpression{
											Object: &semantic.IdentifierExpression{
												Name: "r",
											},
											Property: "_time",
										},
									},
									{
										Key: &semantic.Identifier{Name: "host"},
										Value: &semantic.BinaryExpression{
											Operator: ast.AdditionOperator,
											Left: &semantic.MemberExpression{
												Object: &semantic.IdentifierExpression{
													Name: "r",
												},
												Property: "host",
											},
											Right: &semantic.BinaryExpression{
												Operator: ast.AdditionOperator,
												Left:     &semantic.StringLiteral{Value: "."},
												Right: &semantic.MemberExpression{
													Object: &semantic.IdentifierExpression{
														Name: "r",
													},
													Property: "domain",
												},
											},
										},
									},
									{
										Key: &semantic.Identifier{Name: "_value"},
										Value: &semantic.BinaryExpression{
											Operator: ast.AdditionOperator,
											Left: &semantic.MemberExpression{
												Object: &semantic.IdentifierExpression{
													Name: "r",
												},
												Property: "_value",
											},
											Right: &semantic.FloatLiteral{
												Value: 5,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"_measurement", "host"},
				ColMeta: []flux.ColMeta{
					{Label: "_measurement", Type: flux.TString},
					{Label: "host", Type: flux.TString},
					{Label: "domain", Type: flux.TString},
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{"m", "A", "example.com", execute.Time(1), 1.0},
					{"m", "A", "www.example.com", execute.Time(2), 6.0},
				},
			}},
			want: []*executetest.Table{
				{
					KeyCols: []string{"_measurement", "host"},
					ColMeta: []flux.ColMeta{
						{Label: "_measurement", Type: flux.TString},
						{Label: "host", Type: flux.TString},
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{"m", "A.example.com", execute.Time(1), 6.0},
					},
				},
				{
					KeyCols: []string{"_measurement", "host"},
					ColMeta: []flux.ColMeta{
						{Label: "_measurement", Type: flux.TString},
						{Label: "host", Type: flux.TString},
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{"m", "A.www.example.com", execute.Time(2), 11.0},
					},
				},
			},
		},
		{
			name: `_value+5 mergeKey=true with nulls`,
			spec: &universe.MapProcedureSpec{
				MergeKey: true,
				Fn: interpreter.ResolvedFunction{
					Scope: builtIns,
					Fn: &semantic.FunctionExpression{
						Block: &semantic.FunctionBlock{
							Parameters: &semantic.FunctionParameters{
								List: []*semantic.FunctionParameter{{Key: &semantic.Identifier{Name: "r"}}},
							},
							Body: &semantic.ObjectExpression{
								Properties: []*semantic.Property{
									{
										Key: &semantic.Identifier{Name: "_time"},
										Value: &semantic.MemberExpression{
											Object: &semantic.IdentifierExpression{
												Name: "r",
											},
											Property: "_time",
										},
									},
									{
										Key: &semantic.Identifier{Name: "_value"},
										Value: &semantic.BinaryExpression{
											Operator: ast.AdditionOperator,
											Left: &semantic.MemberExpression{
												Object: &semantic.IdentifierExpression{
													Name: "r",
												},
												Property: "_value",
											},
											Right: &semantic.FloatLiteral{
												Value: 5,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"_measurement", "host"},
				ColMeta: []flux.ColMeta{
					{Label: "_measurement", Type: flux.TString},
					{Label: "host", Type: flux.TString},
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{"m", nil, execute.Time(1), 1.0},
					{"m", nil, execute.Time(2), 6.0},
				},
			}},
			want: []*executetest.Table{{
				KeyCols: []string{"_measurement", "host"},
				ColMeta: []flux.ColMeta{
					{Label: "_measurement", Type: flux.TString},
					{Label: "host", Type: flux.TString},
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{"m", nil, execute.Time(1), 6.0},
					{"m", nil, execute.Time(2), 11.0},
				},
			}},
		},
		{
			name: `_value*_value`,
			spec: &universe.MapProcedureSpec{
				Fn: interpreter.ResolvedFunction{
					Scope: builtIns,
					Fn: &semantic.FunctionExpression{
						Block: &semantic.FunctionBlock{
							Parameters: &semantic.FunctionParameters{
								List: []*semantic.FunctionParameter{{Key: &semantic.Identifier{Name: "r"}}},
							},
							Body: &semantic.ObjectExpression{
								Properties: []*semantic.Property{
									{
										Key: &semantic.Identifier{Name: "_time"},
										Value: &semantic.MemberExpression{
											Object: &semantic.IdentifierExpression{
												Name: "r",
											},
											Property: "_time",
										},
									},
									{
										Key: &semantic.Identifier{Name: "_value"},
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
													Name: "r",
												},
												Property: "_value",
											},
										},
									},
								},
							},
						},
					},
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
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 1.0},
					{execute.Time(2), 36.0},
				},
			}},
		},
		{
			name: "float(r._value) int",
			spec: &universe.MapProcedureSpec{
				Fn: interpreter.ResolvedFunction{
					Scope: builtIns,
					Fn: &semantic.FunctionExpression{
						Block: &semantic.FunctionBlock{
							Parameters: &semantic.FunctionParameters{
								List: []*semantic.FunctionParameter{{Key: &semantic.Identifier{Name: "r"}}},
							},
							Body: &semantic.ObjectExpression{
								Properties: []*semantic.Property{
									{
										Key: &semantic.Identifier{Name: "_time"},
										Value: &semantic.MemberExpression{
											Object: &semantic.IdentifierExpression{
												Name: "r",
											},
											Property: "_time",
										},
									},
									{
										Key: &semantic.Identifier{Name: "_value"},
										Value: &semantic.CallExpression{
											Callee: &semantic.IdentifierExpression{Name: "float"},
											Arguments: &semantic.ObjectExpression{
												Properties: []*semantic.Property{{
													Key: &semantic.Identifier{Name: "v"},
													Value: &semantic.MemberExpression{
														Object: &semantic.IdentifierExpression{
															Name: "r",
														},
														Property: "_value",
													},
												}},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{execute.Time(1), int64(1)},
					{execute.Time(2), int64(6)},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 1.0},
					{execute.Time(2), 6.0},
				},
			}},
		},
		{
			name: "float(r._value) uint",
			spec: &universe.MapProcedureSpec{
				Fn: interpreter.ResolvedFunction{
					Scope: builtIns,
					Fn: &semantic.FunctionExpression{
						Block: &semantic.FunctionBlock{
							Parameters: &semantic.FunctionParameters{
								List: []*semantic.FunctionParameter{{Key: &semantic.Identifier{Name: "r"}}},
							},
							Body: &semantic.ObjectExpression{
								Properties: []*semantic.Property{
									{
										Key: &semantic.Identifier{Name: "_time"},
										Value: &semantic.MemberExpression{
											Object: &semantic.IdentifierExpression{
												Name: "r",
											},
											Property: "_time",
										},
									},
									{
										Key: &semantic.Identifier{Name: "_value"},
										Value: &semantic.CallExpression{
											Callee: &semantic.IdentifierExpression{Name: "float"},
											Arguments: &semantic.ObjectExpression{
												Properties: []*semantic.Property{{
													Key: &semantic.Identifier{Name: "v"},
													Value: &semantic.MemberExpression{
														Object: &semantic.IdentifierExpression{
															Name: "r",
														},
														Property: "_value",
													},
												}},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TUInt},
				},
				Data: [][]interface{}{
					{execute.Time(1), uint64(1)},
					{execute.Time(2), uint64(6)},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 1.0},
					{execute.Time(2), 6.0},
				},
			}},
		},
		{
			name: `float("foo") produces error`,
			spec: &universe.MapProcedureSpec{
				Fn: interpreter.ResolvedFunction{
					Scope: builtIns,
					Fn: &semantic.FunctionExpression{
						Block: &semantic.FunctionBlock{
							Parameters: &semantic.FunctionParameters{
								List: []*semantic.FunctionParameter{{Key: &semantic.Identifier{Name: "r"}}},
							},
							Body: &semantic.ObjectExpression{
								Properties: []*semantic.Property{
									{
										Key: &semantic.Identifier{Name: "_time"},
										Value: &semantic.MemberExpression{
											Object: &semantic.IdentifierExpression{
												Name: "r",
											},
											Property: "_time",
										},
									},
									{
										Key: &semantic.Identifier{Name: "_value"},
										Value: &semantic.CallExpression{
											Callee: &semantic.IdentifierExpression{Name: "float"},
											Arguments: &semantic.ObjectExpression{
												Properties: []*semantic.Property{{
													Key: &semantic.Identifier{Name: "v"},
													Value: &semantic.StringLiteral{
														Value: "foo",
													},
												}},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TUInt},
				},
				Data: [][]interface{}{
					{execute.Time(1), uint64(1)},
					{execute.Time(2), uint64(6)},
				},
			}},
			wantErr: errors.New(`failed to evaluate map function: strconv.ParseFloat: parsing "foo": invalid syntax`),
		},
		{
			name: `with null record`,
			spec: &universe.MapProcedureSpec{
				Fn: interpreter.ResolvedFunction{
					Scope: builtIns,
					Fn: &semantic.FunctionExpression{
						Block: &semantic.FunctionBlock{
							Parameters: &semantic.FunctionParameters{
								List: []*semantic.FunctionParameter{{Key: &semantic.Identifier{Name: "r"}}},
							},
							Body: &semantic.ObjectExpression{
								Properties: []*semantic.Property{
									{
										Key: &semantic.Identifier{Name: "value"},
										Value: &semantic.BinaryExpression{
											Operator: ast.AdditionOperator,
											Left: &semantic.MemberExpression{
												Object: &semantic.IdentifierExpression{
													Name: "r",
												},
												Property: "_value",
											},
											Right: &semantic.FloatLiteral{
												Value: 5,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), nil},
					{execute.Time(2), 6.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{nil},
					{11.0},
				},
			}},
		},
		{
			name: `with null column`,
			spec: &universe.MapProcedureSpec{
				Fn: interpreter.ResolvedFunction{
					Scope: builtIns,
					Fn: &semantic.FunctionExpression{
						Block: &semantic.FunctionBlock{
							Parameters: &semantic.FunctionParameters{
								List: []*semantic.FunctionParameter{{Key: &semantic.Identifier{Name: "r"}}},
							},
							Body: &semantic.ObjectExpression{
								Properties: []*semantic.Property{
									{
										Key: &semantic.Identifier{Name: "value"},
										Value: &semantic.MemberExpression{
											Object: &semantic.IdentifierExpression{
												Name: "r",
											},
											Property: "_value",
										},
									},
									{
										Key: &semantic.Identifier{Name: "missing"},
										Value: &semantic.MemberExpression{
											Object: &semantic.IdentifierExpression{
												Name: "r",
											},
											Property: "_missing",
										},
									},
								},
							},
						},
					},
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
					{Label: "value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0},
					{6.0},
				},
			}},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Skip("https://github.com/influxdata/flux/issues/2489")
			executetest.ProcessTestHelper(
				t,
				tc.data,
				tc.want,
				tc.wantErr,
				func(d execute.Dataset, c execute.TableBuilderCache) execute.Transformation {
					ctx := dependenciestest.Default().Inject(context.Background())
					f, err := universe.NewMapTransformation(ctx, tc.spec, d, c)
					if err != nil {
						t.Fatal(err)
					}
					return f
				},
			)
		})
	}
}
