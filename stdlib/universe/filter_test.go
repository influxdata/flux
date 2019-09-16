package universe_test

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/dependencies/dependenciestest"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/plan/plantest"
	"github.com/influxdata/flux/querytest"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/stdlib/influxdata/influxdb"
	"github.com/influxdata/flux/stdlib/universe"
	"github.com/influxdata/flux/values/valuestest"
)

func TestFilter_NewQuery(t *testing.T) {
	tests := []querytest.NewQueryTestCase{
		{
			Name: "from with database filter and range",
			Raw:  `from(bucket:"mybucket") |> filter(fn: (r) => r["t1"]=="val1" and r["t2"]=="val2") |> range(start:-4h, stop:-2h) |> count()`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: "mybucket",
						},
					},
					{
						ID: "filter1",
						Spec: &universe.FilterOpSpec{
							Fn: interpreter.ResolvedFunction{
								Fn: &semantic.FunctionExpression{
									Block: &semantic.FunctionBlock{
										Parameters: &semantic.FunctionParameters{
											List: []*semantic.FunctionParameter{{Key: &semantic.Identifier{Name: "r"}}},
										},
										Body: &semantic.LogicalExpression{
											Operator: ast.AndOperator,
											Left: &semantic.BinaryExpression{
												Operator: ast.EqualOperator,
												Left: &semantic.MemberExpression{
													Object:   &semantic.IdentifierExpression{Name: "r"},
													Property: "t1",
												},
												Right: &semantic.StringLiteral{Value: "val1"},
											},
											Right: &semantic.BinaryExpression{
												Operator: ast.EqualOperator,
												Left: &semantic.MemberExpression{
													Object:   &semantic.IdentifierExpression{Name: "r"},
													Property: "t2",
												},
												Right: &semantic.StringLiteral{Value: "val2"},
											},
										},
									},
								},
								Scope: valuestest.NowScope(),
							},
						},
					},
					{
						ID: "range2",
						Spec: &universe.RangeOpSpec{
							Start: flux.Time{
								Relative:   -4 * time.Hour,
								IsRelative: true,
							},
							Stop: flux.Time{
								Relative:   -2 * time.Hour,
								IsRelative: true,
							},
							TimeColumn:  "_time",
							StartColumn: "_start",
							StopColumn:  "_stop",
						},
					},
					{
						ID: "count3",
						Spec: &universe.CountOpSpec{
							AggregateConfig: execute.DefaultAggregateConfig,
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "filter1"},
					{Parent: "filter1", Child: "range2"},
					{Parent: "range2", Child: "count3"},
				},
			},
		},
		{
			Name: "from with database filter (and with or) and range",
			Raw: `from(bucket:"mybucket")
						|> filter(fn: (r) =>
								(
									(r["t1"]=="val1")
									and
									(r["t2"]=="val2")
								)
								or
								(r["t3"]=="val3")
							)
						|> range(start:-4h, stop:-2h)
						|> count()`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: "mybucket",
						},
					},
					{
						ID: "filter1",
						Spec: &universe.FilterOpSpec{
							Fn: interpreter.ResolvedFunction{
								Fn: &semantic.FunctionExpression{
									Block: &semantic.FunctionBlock{
										Parameters: &semantic.FunctionParameters{
											List: []*semantic.FunctionParameter{{Key: &semantic.Identifier{Name: "r"}}},
										},
										Body: &semantic.LogicalExpression{
											Operator: ast.OrOperator,
											Left: &semantic.LogicalExpression{
												Operator: ast.AndOperator,
												Left: &semantic.BinaryExpression{
													Operator: ast.EqualOperator,
													Left: &semantic.MemberExpression{
														Object:   &semantic.IdentifierExpression{Name: "r"},
														Property: "t1",
													},
													Right: &semantic.StringLiteral{Value: "val1"},
												},
												Right: &semantic.BinaryExpression{
													Operator: ast.EqualOperator,
													Left: &semantic.MemberExpression{
														Object:   &semantic.IdentifierExpression{Name: "r"},
														Property: "t2",
													},
													Right: &semantic.StringLiteral{Value: "val2"},
												},
											},
											Right: &semantic.BinaryExpression{
												Operator: ast.EqualOperator,
												Left: &semantic.MemberExpression{
													Object:   &semantic.IdentifierExpression{Name: "r"},
													Property: "t3",
												},
												Right: &semantic.StringLiteral{Value: "val3"},
											},
										},
									},
								},
								Scope: valuestest.NowScope(),
							},
						},
					},
					{
						ID: "range2",
						Spec: &universe.RangeOpSpec{
							Start: flux.Time{
								Relative:   -4 * time.Hour,
								IsRelative: true,
							},
							Stop: flux.Time{
								Relative:   -2 * time.Hour,
								IsRelative: true,
							},
							TimeColumn:  "_time",
							StartColumn: "_start",
							StopColumn:  "_stop",
						},
					},
					{
						ID: "count3",
						Spec: &universe.CountOpSpec{
							AggregateConfig: execute.DefaultAggregateConfig,
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "filter1"},
					{Parent: "filter1", Child: "range2"},
					{Parent: "range2", Child: "count3"},
				},
			},
		},
		{
			Name: "from with database filter including fields",
			Raw: `from(bucket:"mybucket")
						|> filter(fn: (r) =>
							(r["t1"] =="val1")
							and
							(r["_field"] == 10)
						)
						|> range(start:-4h, stop:-2h)
						|> count()`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: "mybucket",
						},
					},
					{
						ID: "filter1",
						Spec: &universe.FilterOpSpec{
							Fn: interpreter.ResolvedFunction{
								Scope: valuestest.NowScope(),
								Fn: &semantic.FunctionExpression{
									Block: &semantic.FunctionBlock{
										Parameters: &semantic.FunctionParameters{
											List: []*semantic.FunctionParameter{{Key: &semantic.Identifier{Name: "r"}}},
										},
										Body: &semantic.LogicalExpression{
											Operator: ast.AndOperator,
											Left: &semantic.BinaryExpression{
												Operator: ast.EqualOperator,
												Left: &semantic.MemberExpression{
													Object:   &semantic.IdentifierExpression{Name: "r"},
													Property: "t1",
												},
												Right: &semantic.StringLiteral{Value: "val1"},
											},
											Right: &semantic.BinaryExpression{
												Operator: ast.EqualOperator,
												Left: &semantic.MemberExpression{
													Object:   &semantic.IdentifierExpression{Name: "r"},
													Property: "_field",
												},
												Right: &semantic.IntegerLiteral{Value: 10},
											},
										},
									},
								},
							},
						},
					},
					{
						ID: "range2",
						Spec: &universe.RangeOpSpec{
							Start: flux.Time{
								Relative:   -4 * time.Hour,
								IsRelative: true,
							},
							Stop: flux.Time{
								Relative:   -2 * time.Hour,
								IsRelative: true,
							},
							TimeColumn:  "_time",
							StartColumn: "_start",
							StopColumn:  "_stop",
						},
					},
					{
						ID: "count3",
						Spec: &universe.CountOpSpec{
							AggregateConfig: execute.DefaultAggregateConfig,
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "filter1"},
					{Parent: "filter1", Child: "range2"},
					{Parent: "range2", Child: "count3"},
				},
			},
		},
		{
			Name: "from with database filter with no parens including fields",
			Raw: `from(bucket:"mybucket")
						|> filter(fn: (r) =>
							r["t1"]=="val1"
							and
							r["_field"] == 10
						)
						|> range(start:-4h, stop:-2h)
						|> count()`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: "mybucket",
						},
					},
					{
						ID: "filter1",
						Spec: &universe.FilterOpSpec{
							Fn: interpreter.ResolvedFunction{
								Fn: &semantic.FunctionExpression{
									Block: &semantic.FunctionBlock{
										Parameters: &semantic.FunctionParameters{
											List: []*semantic.FunctionParameter{{Key: &semantic.Identifier{Name: "r"}}},
										},
										Body: &semantic.LogicalExpression{
											Operator: ast.AndOperator,
											Left: &semantic.BinaryExpression{
												Operator: ast.EqualOperator,
												Left: &semantic.MemberExpression{
													Object:   &semantic.IdentifierExpression{Name: "r"},
													Property: "t1",
												},
												Right: &semantic.StringLiteral{Value: "val1"},
											},
											Right: &semantic.BinaryExpression{
												Operator: ast.EqualOperator,
												Left: &semantic.MemberExpression{
													Object:   &semantic.IdentifierExpression{Name: "r"},
													Property: "_field",
												},
												Right: &semantic.IntegerLiteral{Value: 10},
											},
										},
									},
								},
								Scope: valuestest.NowScope(),
							},
						},
					},
					{
						ID: "range2",
						Spec: &universe.RangeOpSpec{
							Start: flux.Time{
								Relative:   -4 * time.Hour,
								IsRelative: true,
							},
							Stop: flux.Time{
								Relative:   -2 * time.Hour,
								IsRelative: true,
							},
							TimeColumn:  "_time",
							StartColumn: "_start",
							StopColumn:  "_stop",
						},
					},
					{
						ID: "count3",
						Spec: &universe.CountOpSpec{
							AggregateConfig: execute.DefaultAggregateConfig,
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "filter1"},
					{Parent: "filter1", Child: "range2"},
					{Parent: "range2", Child: "count3"},
				},
			},
		},
		{
			Name: "from with database filter with no parens including regex and field",
			Raw: `from(bucket:"mybucket")
						|> filter(fn: (r) =>
							r["t1"]==/^val1/
							and
							r["_field"] == 10.5
						)
						|> range(start:-4h, stop:-2h)
						|> count()`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: "mybucket",
						},
					},
					{
						ID: "filter1",
						Spec: &universe.FilterOpSpec{
							Fn: interpreter.ResolvedFunction{
								Fn: &semantic.FunctionExpression{
									Block: &semantic.FunctionBlock{
										Parameters: &semantic.FunctionParameters{
											List: []*semantic.FunctionParameter{{Key: &semantic.Identifier{Name: "r"}}},
										},
										Body: &semantic.LogicalExpression{
											Operator: ast.AndOperator,
											Left: &semantic.BinaryExpression{
												Operator: ast.EqualOperator,
												Left: &semantic.MemberExpression{
													Object:   &semantic.IdentifierExpression{Name: "r"},
													Property: "t1",
												},
												Right: &semantic.RegexpLiteral{Value: regexp.MustCompile("^val1")},
											},
											Right: &semantic.BinaryExpression{
												Operator: ast.EqualOperator,
												Left: &semantic.MemberExpression{
													Object:   &semantic.IdentifierExpression{Name: "r"},
													Property: "_field",
												},
												Right: &semantic.FloatLiteral{Value: 10.5},
											},
										},
									},
								},
								Scope: valuestest.NowScope(),
							},
						},
					},
					{
						ID: "range2",
						Spec: &universe.RangeOpSpec{
							Start: flux.Time{
								Relative:   -4 * time.Hour,
								IsRelative: true,
							},
							Stop: flux.Time{
								Relative:   -2 * time.Hour,
								IsRelative: true,
							},
							TimeColumn:  "_time",
							StartColumn: "_start",
							StopColumn:  "_stop",
						},
					},
					{
						ID: "count3",
						Spec: &universe.CountOpSpec{
							AggregateConfig: execute.DefaultAggregateConfig,
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "filter1"},
					{Parent: "filter1", Child: "range2"},
					{Parent: "range2", Child: "count3"},
				},
			},
		},
		{
			Name: "from with database regex with escape",
			Raw: `from(bucket:"mybucket")
						|> filter(fn: (r) =>
							r["t1"]==/^va\/l1/
						)`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: "mybucket",
						},
					},
					{
						ID: "filter1",
						Spec: &universe.FilterOpSpec{
							Fn: interpreter.ResolvedFunction{
								Fn: &semantic.FunctionExpression{
									Block: &semantic.FunctionBlock{
										Parameters: &semantic.FunctionParameters{
											List: []*semantic.FunctionParameter{{Key: &semantic.Identifier{Name: "r"}}},
										},
										Body: &semantic.BinaryExpression{
											Operator: ast.EqualOperator,
											Left: &semantic.MemberExpression{
												Object:   &semantic.IdentifierExpression{Name: "r"},
												Property: "t1",
											},
											Right: &semantic.RegexpLiteral{Value: regexp.MustCompile(`^va/l1`)},
										},
									},
								},
								Scope: valuestest.NowScope(),
							},
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "filter1"},
				},
			},
		},
		{
			Name: "from with database with two regex",
			Raw: `from(bucket:"mybucket")
						|> filter(fn: (r) =>
							r["t1"]==/^va\/l1/
							and
							r["t2"] != /^val2/
						)`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: "mybucket",
						},
					},
					{
						ID: "filter1",
						Spec: &universe.FilterOpSpec{
							Fn: interpreter.ResolvedFunction{
								Fn: &semantic.FunctionExpression{
									Block: &semantic.FunctionBlock{
										Parameters: &semantic.FunctionParameters{
											List: []*semantic.FunctionParameter{{Key: &semantic.Identifier{Name: "r"}}},
										},
										Body: &semantic.LogicalExpression{
											Operator: ast.AndOperator,
											Left: &semantic.BinaryExpression{
												Operator: ast.EqualOperator,
												Left: &semantic.MemberExpression{
													Object:   &semantic.IdentifierExpression{Name: "r"},
													Property: "t1",
												},
												Right: &semantic.RegexpLiteral{Value: regexp.MustCompile(`^va/l1`)},
											},
											Right: &semantic.BinaryExpression{
												Operator: ast.NotEqualOperator,
												Left: &semantic.MemberExpression{
													Object:   &semantic.IdentifierExpression{Name: "r"},
													Property: "t2",
												},
												Right: &semantic.RegexpLiteral{Value: regexp.MustCompile(`^val2`)},
											},
										},
									},
								},
								Scope: valuestest.NowScope(),
							},
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "filter1"},
				},
			},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			querytest.NewQueryTestHelper(t, tc)
		})
	}
}
func TestFilterOperation_Marshaling(t *testing.T) {
	data := []byte(`{
		"id":"filter",
		"kind":"filter",
		"spec":{
			"fn":{
				"fn":{
					"type": "FunctionExpression",
					"block":{
						"type":"FunctionBlock",
						"parameters": {
							"type":"FunctionParameters",
							"list": [
								{"type":"FunctionParameter","key":{"type":"Identifier","name":"r"}}
							]
						},
						"body":{
							"type":"BinaryExpression",
							"operator": "!=",
							"left":{
								"type":"MemberExpression",
								"object": {
									"type": "IdentifierExpression",
									"name":"r"
								},
								"property": "_measurement"
							},
							"right":{
								"type":"StringLiteral",
								"value":"mem"
							}
						}
					}
				}
			}
		}
	}`)
	op := &flux.Operation{
		ID: "filter",
		Spec: &universe.FilterOpSpec{
			Fn: interpreter.ResolvedFunction{
				Fn: &semantic.FunctionExpression{
					Block: &semantic.FunctionBlock{
						Parameters: &semantic.FunctionParameters{
							List: []*semantic.FunctionParameter{{Key: &semantic.Identifier{Name: "r"}}},
						},
						Body: &semantic.BinaryExpression{
							Operator: ast.NotEqualOperator,
							Left: &semantic.MemberExpression{
								Object:   &semantic.IdentifierExpression{Name: "r"},
								Property: "_measurement",
							},
							Right: &semantic.StringLiteral{Value: "mem"},
						},
					},
				},
			},
		},
	}
	querytest.OperationMarshalingTestHelper(t, data, op)
}

func TestMergeFilterAnyRule(t *testing.T) {
	var (
		from        = &influxdb.FromProcedureSpec{}
		count       = &universe.CountProcedureSpec{}
		filterOther = &universe.FilterProcedureSpec{
			Fn: interpreter.ResolvedFunction{
				Fn: &semantic.FunctionExpression{
					Block: &semantic.FunctionBlock{
						Body: &semantic.IdentifierExpression{
							Name: "foo",
						},
					},
				},
				Scope: valuestest.NowScope(),
			},
		}
		filterTrue = &universe.FilterProcedureSpec{
			Fn: interpreter.ResolvedFunction{
				Fn: &semantic.FunctionExpression{
					Block: &semantic.FunctionBlock{
						Body: &semantic.BooleanLiteral{
							Value: true,
						},
					},
				},
				Scope: valuestest.NowScope(),
			},
		}
		filterFalse = &universe.FilterProcedureSpec{
			Fn: interpreter.ResolvedFunction{
				Fn: &semantic.FunctionExpression{
					Block: &semantic.FunctionBlock{
						Body: &semantic.BooleanLiteral{
							Value: false,
						},
					},
				},
				Scope: valuestest.NowScope(),
			},
		}
	)

	tests := []plantest.RuleTestCase{
		{
			Name: "filterOther",
			// from -> filter => from -> filter
			Rules: []plan.Rule{universe.RemoveTrivialFilterRule{}},
			Before: &plantest.PlanSpec{
				Nodes: []plan.Node{
					plan.CreatePhysicalNode("from", from),
					plan.CreatePhysicalNode("filter", filterOther),
				},
				Edges: [][2]int{{0, 1}},
			},
			NoChange: true,
		},
		{
			Name: "filterFalse",
			// from -> filter => from -> filter
			Rules: []plan.Rule{universe.RemoveTrivialFilterRule{}},
			Before: &plantest.PlanSpec{
				Nodes: []plan.Node{
					plan.CreatePhysicalNode("from", from),
					plan.CreatePhysicalNode("filter", filterFalse),
				},
				Edges: [][2]int{{0, 1}},
			},
			NoChange: true,
		},
		{
			Name: "filterTrue",
			// from -> filter => from
			Rules: []plan.Rule{universe.RemoveTrivialFilterRule{}},
			Before: &plantest.PlanSpec{
				Nodes: []plan.Node{
					plan.CreatePhysicalNode("from", from),
					plan.CreatePhysicalNode("filter", filterTrue),
				},
				Edges: [][2]int{{0, 1}},
			},
			After: &plantest.PlanSpec{
				Nodes: []plan.Node{
					plan.CreatePhysicalNode("from", from),
				},
			},
		},
		{
			Name: "count filterTrue",
			// count -> filter => count
			Rules: []plan.Rule{universe.RemoveTrivialFilterRule{}},
			Before: &plantest.PlanSpec{
				Nodes: []plan.Node{
					plan.CreatePhysicalNode("count", count),
					plan.CreatePhysicalNode("filter", filterTrue),
				},
				Edges: [][2]int{{0, 1}},
			},
			After: &plantest.PlanSpec{
				Nodes: []plan.Node{
					plan.CreatePhysicalNode("count", count),
				},
			},
		},
		{
			Name: "from filterTrue count",
			// from -> filter -> count => from -> count
			Rules: []plan.Rule{universe.RemoveTrivialFilterRule{}},
			Before: &plantest.PlanSpec{
				Nodes: []plan.Node{
					plan.CreatePhysicalNode("from", from),
					plan.CreatePhysicalNode("filter", filterTrue),
					plan.CreatePhysicalNode("count", count),
				},
				Edges: [][2]int{{0, 1}, {1, 2}},
			},
			After: &plantest.PlanSpec{
				Nodes: []plan.Node{
					plan.CreatePhysicalNode("from", from),
					plan.CreatePhysicalNode("count", count),
				},
				Edges: [][2]int{{0, 1}},
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			plantest.PhysicalRuleTestHelper(t, &tc)
		})
	}
}

func TestFilter_Process(t *testing.T) {
	testCases := []struct {
		name string
		spec *universe.FilterProcedureSpec
		data []flux.Table
		want []*executetest.Table
	}{
		{
			name: `_value>5`,
			spec: &universe.FilterProcedureSpec{
				Fn: interpreter.ResolvedFunction{
					Fn: &semantic.FunctionExpression{
						Block: &semantic.FunctionBlock{
							Parameters: &semantic.FunctionParameters{
								List: []*semantic.FunctionParameter{{Key: &semantic.Identifier{Name: "r"}}},
							},
							Body: &semantic.BinaryExpression{
								Operator: ast.GreaterThanOperator,
								Left: &semantic.MemberExpression{
									Object:   &semantic.IdentifierExpression{Name: "r"},
									Property: "_value",
								},
								Right: &semantic.FloatLiteral{Value: 5},
							},
						},
					},
					Scope: valuestest.NowScope(),
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
					{execute.Time(2), 6.0},
				},
			}},
		},
		{
			name: "_value>5 multiple blocks",
			spec: &universe.FilterProcedureSpec{
				Fn: interpreter.ResolvedFunction{
					Fn: &semantic.FunctionExpression{
						Block: &semantic.FunctionBlock{
							Parameters: &semantic.FunctionParameters{
								List: []*semantic.FunctionParameter{{Key: &semantic.Identifier{Name: "r"}}},
							},
							Body: &semantic.BinaryExpression{
								Operator: ast.GreaterThanOperator,
								Left: &semantic.MemberExpression{
									Object:   &semantic.IdentifierExpression{Name: "r"},
									Property: "_value",
								},
								Right: &semantic.FloatLiteral{
									Value: 5,
								},
							},
						},
					},
					Scope: valuestest.NowScope(),
				},
			},
			data: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"t1"},
					ColMeta: []flux.ColMeta{
						{Label: "t1", Type: flux.TString},
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{"a", execute.Time(1), 3.0},
						{"a", execute.Time(2), 6.0},
						{"a", execute.Time(2), 1.0},
					},
				},
				&executetest.Table{
					KeyCols: []string{"t1"},
					ColMeta: []flux.ColMeta{
						{Label: "t1", Type: flux.TString},
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{"b", execute.Time(3), 3.0},
						{"b", execute.Time(3), 2.0},
						{"b", execute.Time(4), 8.0},
					},
				},
			},
			want: []*executetest.Table{
				{
					KeyCols: []string{"t1"},
					ColMeta: []flux.ColMeta{
						{Label: "t1", Type: flux.TString},
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{"a", execute.Time(2), 6.0},
					},
				},
				{
					KeyCols: []string{"t1"},
					ColMeta: []flux.ColMeta{
						{Label: "t1", Type: flux.TString},
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{"b", execute.Time(4), 8.0},
					},
				},
			},
		},
		{
			name: "_value>5 and t1 = a and t2 = y",
			spec: &universe.FilterProcedureSpec{
				Fn: interpreter.ResolvedFunction{
					Fn: &semantic.FunctionExpression{
						Block: &semantic.FunctionBlock{
							Parameters: &semantic.FunctionParameters{
								List: []*semantic.FunctionParameter{{Key: &semantic.Identifier{Name: "r"}}},
							},
							Body: &semantic.LogicalExpression{
								Operator: ast.AndOperator,
								Left: &semantic.BinaryExpression{
									Operator: ast.GreaterThanOperator,
									Left: &semantic.MemberExpression{
										Object:   &semantic.IdentifierExpression{Name: "r"},
										Property: "_value",
									},
									Right: &semantic.FloatLiteral{
										Value: 5,
									},
								},
								Right: &semantic.LogicalExpression{
									Operator: ast.AndOperator,
									Left: &semantic.BinaryExpression{
										Operator: ast.EqualOperator,
										Left: &semantic.MemberExpression{
											Object:   &semantic.IdentifierExpression{Name: "r"},
											Property: "t1",
										},
										Right: &semantic.StringLiteral{
											Value: "a",
										},
									},
									Right: &semantic.BinaryExpression{
										Operator: ast.EqualOperator,
										Left: &semantic.MemberExpression{
											Object:   &semantic.IdentifierExpression{Name: "r"},
											Property: "t2",
										},
										Right: &semantic.StringLiteral{
											Value: "y",
										},
									},
								},
							},
						},
					},
					Scope: valuestest.NowScope(),
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "t1", Type: flux.TString},
					{Label: "t2", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(1), 1.0, "a", "x"},
					{execute.Time(2), 6.0, "a", "x"},
					{execute.Time(3), 8.0, "a", "y"},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "t1", Type: flux.TString},
					{Label: "t2", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(3), 8.0, "a", "y"},
				},
			}},
		},
		{
			name: `_value>5 with unused nulls`,
			spec: &universe.FilterProcedureSpec{
				Fn: interpreter.ResolvedFunction{
					Fn: &semantic.FunctionExpression{
						Block: &semantic.FunctionBlock{
							Parameters: &semantic.FunctionParameters{
								List: []*semantic.FunctionParameter{{Key: &semantic.Identifier{Name: "r"}}},
							},
							Body: &semantic.BinaryExpression{
								Operator: ast.GreaterThanOperator,
								Left: &semantic.MemberExpression{
									Object:   &semantic.IdentifierExpression{Name: "r"},
									Property: "_value",
								},
								Right: &semantic.FloatLiteral{Value: 5},
							},
						},
					},
					Scope: valuestest.NowScope(),
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "host", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(1), 1.0, "server01"},
					{execute.Time(2), 1.0, nil},
					{execute.Time(3), 6.0, "server02"},
					{execute.Time(4), 6.0, nil},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "host", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(3), 6.0, "server02"},
					{execute.Time(4), 6.0, nil},
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
				nil,
				func(d execute.Dataset, c execute.TableBuilderCache) execute.Transformation {
					ctx := dependenciestest.Default().Inject(context.Background())
					f, err := universe.NewFilterTransformation(ctx, tc.spec, d, c)
					if err != nil {
						t.Fatal(err)
					}
					return f
				},
			)
		})
	}
}
