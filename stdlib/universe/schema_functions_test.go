package universe_test

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/dependencies/dependenciestest"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/querytest"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/stdlib/influxdata/influxdb"
	"github.com/influxdata/flux/stdlib/universe"
	"github.com/influxdata/flux/values/valuestest"
)

func TestSchemaMutions_NewQueries(t *testing.T) {
	tests := []querytest.NewQueryTestCase{
		{
			Name: "test rename query",
			Raw:  `from(bucket:"mybucket") |> rename(columns:{old:"new"}) |> sum()`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: "mybucket",
						},
					},
					{
						ID: "rename1",
						Spec: &universe.RenameOpSpec{
							Columns: map[string]string{
								"old": "new",
							},
						},
					},
					{
						ID: "sum2",
						Spec: &universe.SumOpSpec{
							AggregateConfig: execute.DefaultAggregateConfig,
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "rename1"},
					{Parent: "rename1", Child: "sum2"},
				},
			},
		},
		{
			Name: "test drop query",
			Raw:  `from(bucket:"mybucket") |> drop(columns:["col1", "col2", "col3"]) |> sum()`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: "mybucket",
						},
					},
					{
						ID: "drop1",
						Spec: &universe.DropOpSpec{
							Columns: []string{"col1", "col2", "col3"},
						},
					},
					{
						ID: "sum2",
						Spec: &universe.SumOpSpec{
							AggregateConfig: execute.DefaultAggregateConfig,
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "drop1"},
					{Parent: "drop1", Child: "sum2"},
				},
			},
		},
		{
			Name: "test keep query",
			Raw:  `from(bucket:"mybucket") |> keep(columns:["col1", "col2", "col3"]) |> sum()`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: "mybucket",
						},
					},
					{
						ID: "keep1",
						Spec: &universe.KeepOpSpec{
							Columns: []string{"col1", "col2", "col3"},
						},
					},
					{
						ID: "sum2",
						Spec: &universe.SumOpSpec{
							AggregateConfig: execute.DefaultAggregateConfig,
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "keep1"},
					{Parent: "keep1", Child: "sum2"},
				},
			},
		},
		{
			Name: "test duplicate query",
			Raw:  `from(bucket:"mybucket") |> duplicate(column: "col1", as: "col1_new") |> sum()`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: "mybucket",
						},
					},
					{
						ID: "duplicate1",
						Spec: &universe.DuplicateOpSpec{
							Column: "col1",
							As:     "col1_new",
						},
					},
					{
						ID: "sum2",
						Spec: &universe.SumOpSpec{
							AggregateConfig: execute.DefaultAggregateConfig,
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "duplicate1"},
					{Parent: "duplicate1", Child: "sum2"},
				},
			},
		},
		{
			Name: "test drop query fn param",
			Raw:  `from(bucket:"mybucket") |> drop(fn: (column) => column =~ /reg*/) |> sum()`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: "mybucket",
						},
					},
					{
						ID: "drop1",
						Spec: &universe.DropOpSpec{
							Predicate: interpreter.ResolvedFunction{
								Fn: &semantic.FunctionExpression{
									Block: &semantic.FunctionBlock{
										Parameters: &semantic.FunctionParameters{
											List: []*semantic.FunctionParameter{{Key: &semantic.Identifier{Name: "column"}}},
										},
										Body: &semantic.BinaryExpression{
											Operator: ast.RegexpMatchOperator,
											Left: &semantic.IdentifierExpression{
												Name: "column",
											},
											Right: &semantic.RegexpLiteral{
												Value: regexp.MustCompile(`reg*`),
											},
										},
									},
								},
								Scope: valuestest.Scope(),
							},
						},
					},
					{
						ID: "sum2",
						Spec: &universe.SumOpSpec{
							AggregateConfig: execute.DefaultAggregateConfig,
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "drop1"},
					{Parent: "drop1", Child: "sum2"},
				},
			},
		},
		{
			Name: "test keep query fn param",
			Raw:  `from(bucket:"mybucket") |> keep(fn: (column) => column =~ /reg*/) |> sum()`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: "mybucket",
						},
					},
					{
						ID: "keep1",
						Spec: &universe.KeepOpSpec{
							Predicate: interpreter.ResolvedFunction{
								Fn: &semantic.FunctionExpression{
									Block: &semantic.FunctionBlock{
										Parameters: &semantic.FunctionParameters{
											List: []*semantic.FunctionParameter{{Key: &semantic.Identifier{Name: "column"}}},
										},
										Body: &semantic.BinaryExpression{
											Operator: ast.RegexpMatchOperator,
											Left: &semantic.IdentifierExpression{
												Name: "column",
											},
											Right: &semantic.RegexpLiteral{
												Value: regexp.MustCompile(`reg*`),
											},
										},
									},
								},
								Scope: valuestest.Scope(),
							},
						},
					},
					{
						ID: "sum2",
						Spec: &universe.SumOpSpec{
							AggregateConfig: execute.DefaultAggregateConfig,
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "keep1"},
					{Parent: "keep1", Child: "sum2"},
				},
			},
		},
		{
			Name: "test rename query fn param",
			Raw:  `from(bucket:"mybucket") |> rename(fn: (column) => "new_name") |> sum()`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: "mybucket",
						},
					},
					{
						ID: "rename1",
						Spec: &universe.RenameOpSpec{
							Fn: interpreter.ResolvedFunction{
								Fn: &semantic.FunctionExpression{
									Block: &semantic.FunctionBlock{
										Parameters: &semantic.FunctionParameters{
											List: []*semantic.FunctionParameter{{Key: &semantic.Identifier{Name: "column"}}},
										},
										Body: &semantic.StringLiteral{
											Value: "new_name",
										},
									},
								},
								Scope: valuestest.Scope(),
							},
						},
					},
					{
						ID: "sum2",
						Spec: &universe.SumOpSpec{
							AggregateConfig: execute.DefaultAggregateConfig,
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "rename1"},
					{Parent: "rename1", Child: "sum2"},
				},
			},
		},
		{
			Name:    "test rename query invalid",
			Raw:     `from(bucket:"mybucket") |> rename(fn: (column) => "new_name", columns: {a:"b", c:"d"}) |> sum()`,
			Want:    nil,
			WantErr: true,
		},
		{
			Name:    "test drop query invalid",
			Raw:     `from(bucket:"mybucket") |> drop(fn: (column) => column == target, columns: ["a", "b"]) |> sum()`,
			Want:    nil,
			WantErr: true,
		},
		{
			Name:    "test keep query invalid",
			Raw:     `from(bucket:"mybucket") |> keep(fn: (column) => column == target, columns: ["a", "b"]) |> sum()`,
			Want:    nil,
			WantErr: true,
		},
		{
			Name:    "test duplicate query invalid",
			Raw:     `from(bucket:"mybucket") |> duplicate(columns: ["a", "b"], n: -1) |> sum()`,
			Want:    nil,
			WantErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Skip("https://github.com/influxdata/flux/issues/2490")
			t.Parallel()
			querytest.NewQueryTestHelper(t, tc)
		})
	}
}

func TestDropRenameKeep_Process(t *testing.T) {
	testCases := []struct {
		name    string
		spec    plan.ProcedureSpec
		data    []flux.Table
		want    []*executetest.Table
		wantErr error
	}{
		{
			name: "rename multiple cols",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.RenameOpSpec{
						Columns: map[string]string{
							"1a": "1b",
							"2a": "2b",
							"3a": "3b",
						},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "1a", Type: flux.TFloat},
					{Label: "2a", Type: flux.TFloat},
					{Label: "3a", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{11.0, 12.0, 13.0},
					{21.0, 22.0, 23.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "1b", Type: flux.TFloat},
					{Label: "2b", Type: flux.TFloat},
					{Label: "3b", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{11.0, 12.0, 13.0},
					{21.0, 22.0, 23.0},
				},
			}},
		},

		{
			name: "drop multiple cols",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.DropOpSpec{
						Columns: []string{"a", "b"},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "a", Type: flux.TFloat},
					{Label: "b", Type: flux.TFloat},
					{Label: "c", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{11.0, 12.0, 13.0},
					{21.0, 22.0, 23.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "c", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{3.0},
					{13.0},
					{23.0},
				},
			}},
		},
		{
			name: "drop key col merge tables",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.DropOpSpec{
						Columns: []string{"b"},
					},
				},
			},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "a", Type: flux.TString},
						{Label: "b", Type: flux.TString},
						{Label: "c", Type: flux.TFloat},
					},
					KeyCols: []string{"a", "b"},
					Data: [][]interface{}{
						{"one", "two", 3.0},
						{"one", "two", 13.0},
					},
				},
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "a", Type: flux.TString},
						{Label: "b", Type: flux.TString},
						{Label: "c", Type: flux.TFloat},
					},
					KeyCols: []string{"a", "b"},
					Data: [][]interface{}{
						{"one", "three", 5.0},
						{"one", "three", 15.0},
					},
				},
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "a", Type: flux.TString},
					{Label: "c", Type: flux.TFloat},
				},
				KeyCols: []string{"a"},
				Data: [][]interface{}{
					{"one", 3.0},
					{"one", 13.0},
					{"one", 5.0},
					{"one", 15.0},
				},
			}},
		},
		{
			name: "drop key col merge error column count",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.DropOpSpec{
						Columns: []string{"b"},
					},
				},
			},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "a", Type: flux.TString},
						{Label: "b", Type: flux.TString},
						{Label: "c", Type: flux.TFloat},
					},
					KeyCols: []string{"a", "b"},
					Data: [][]interface{}{
						{"one", "two", 3.0},
						{"one", "two", 13.0},
					},
				},
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "a", Type: flux.TString},
						{Label: "b", Type: flux.TString},
					},
					KeyCols: []string{"a", "b"},
					Data: [][]interface{}{
						{"one", "three"},
						{"one", "three"},
					},
				},
			},
			wantErr: errors.New("requested operation merges tables with different numbers of columns for group key {a=one}"),
		},
		{
			name: "drop key col merge error column type",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.DropOpSpec{
						Columns: []string{"b"},
					},
				},
			},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "a", Type: flux.TString},
						{Label: "b", Type: flux.TString},
						{Label: "c", Type: flux.TFloat},
					},
					KeyCols: []string{"a", "b"},
					Data: [][]interface{}{
						{"one", "two", 3.0},
						{"one", "two", 13.0},
					},
				},
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "a", Type: flux.TString},
						{Label: "b", Type: flux.TString},
						{Label: "c", Type: flux.TString},
					},
					KeyCols: []string{"a", "b"},
					Data: [][]interface{}{
						{"one", "three", "val"},
						{"one", "three", "val"},
					},
				},
			},
			wantErr: errors.New("requested operation merges tables with different schemas for group key {a=one}"),
		},
		{
			name: "drop no exist",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.DropOpSpec{
						Columns: []string{"boo"},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "a", Type: flux.TString},
					{Label: "b", Type: flux.TString},
					{Label: "c", Type: flux.TFloat},
				},
				KeyCols: []string{"a", "b"},
				Data: [][]interface{}{
					{"one", "two", 3.0},
					{"one", "two", 13.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "a", Type: flux.TString},
					{Label: "b", Type: flux.TString},
					{Label: "c", Type: flux.TFloat},
				},
				KeyCols: []string{"a", "b"},
				Data: [][]interface{}{
					{"one", "two", 3.0},
					{"one", "two", 13.0},
				},
			}},
		},
		{
			name: "keep multiple cols",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.KeepOpSpec{
						Columns: []string{"a"},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "a", Type: flux.TFloat},
					{Label: "b", Type: flux.TFloat},
					{Label: "c", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{11.0, 12.0, 13.0},
					{21.0, 22.0, 23.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "a", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0},
					{11.0},
					{21.0},
				},
			}},
		},
		{
			name: "keep one key col merge tables",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.KeepOpSpec{
						Columns: []string{"a", "c"},
					},
				},
			},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "a", Type: flux.TString},
						{Label: "b", Type: flux.TString},
						{Label: "c", Type: flux.TFloat},
					},
					KeyCols: []string{"a", "b"},
					Data: [][]interface{}{
						{"one", "two", 3.0},
						{"one", "two", 13.0},
					},
				},
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "a", Type: flux.TString},
						{Label: "b", Type: flux.TString},
						{Label: "c", Type: flux.TFloat},
					},
					KeyCols: []string{"a", "b"},
					Data: [][]interface{}{
						{"one", "three", 5.0},
						{"one", "three", 15.0},
					},
				},
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "a", Type: flux.TString},
					{Label: "c", Type: flux.TFloat},
				},
				KeyCols: []string{"a"},
				Data: [][]interface{}{
					{"one", 3.0},
					{"one", 13.0},
					{"one", 5.0},
					{"one", 15.0},
				},
			}},
		},
		{
			name: "keep one key col merge error column count",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.KeepOpSpec{
						Columns: []string{"a", "c"},
					},
				},
			},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "a", Type: flux.TString},
						{Label: "b", Type: flux.TString},
						{Label: "c", Type: flux.TFloat},
					},
					KeyCols: []string{"a", "b"},
					Data: [][]interface{}{
						{"one", "two", 3.0},
						{"one", "two", 13.0},
					},
				},
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "a", Type: flux.TString},
						{Label: "b", Type: flux.TString},
					},
					KeyCols: []string{"a", "b"},
					Data: [][]interface{}{
						{"one", "three"},
						{"one", "three"},
					},
				},
			},
			wantErr: errors.New("requested operation merges tables with different numbers of columns for group key {a=one}"),
		},
		{
			name: "keep one key col merge error column type",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.KeepOpSpec{
						Columns: []string{"a", "c"},
					},
				},
			},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "a", Type: flux.TString},
						{Label: "b", Type: flux.TString},
						{Label: "c", Type: flux.TFloat},
					},
					KeyCols: []string{"a", "b"},
					Data: [][]interface{}{
						{"one", "two", 3.0},
						{"one", "two", 13.0},
					},
				},
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "a", Type: flux.TString},
						{Label: "b", Type: flux.TString},
						{Label: "c", Type: flux.TString},
					},
					KeyCols: []string{"a", "b"},
					Data: [][]interface{}{
						{"one", "three", "foo"},
						{"one", "three", "bar"},
					},
				},
			},
			wantErr: errors.New("requested operation merges tables with different schemas for group key {a=one}"),
		},
		{
			name: "duplicate single col",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.DuplicateOpSpec{
						Column: "a",
						As:     "a_1",
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "a", Type: flux.TFloat},
					{Label: "b", Type: flux.TFloat},
					{Label: "c", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{11.0, 12.0, 13.0},
					{21.0, 22.0, 23.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "a", Type: flux.TFloat},
					{Label: "b", Type: flux.TFloat},
					{Label: "c", Type: flux.TFloat},
					{Label: "a_1", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0, 1.0},
					{11.0, 12.0, 13.0, 11.0},
					{21.0, 22.0, 23.0, 21.0},
				},
			}},
		},
		{
			name: "rename map fn (column) => name",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.RenameOpSpec{
						Fn: interpreter.ResolvedFunction{
							Fn: &semantic.FunctionExpression{
								Block: &semantic.FunctionBlock{
									Parameters: &semantic.FunctionParameters{
										List: []*semantic.FunctionParameter{{Key: &semantic.Identifier{Name: "column"}}},
									},
									Body: &semantic.StringLiteral{
										Value: "new_name",
									},
								},
							},
							Scope: valuestest.Scope(),
						},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "1a", Type: flux.TFloat},
					{Label: "2a", Type: flux.TFloat},
					{Label: "3a", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{11.0, 12.0, 13.0},
					{21.0, 22.0, 23.0},
				},
			}},
			wantErr: errors.New("table builder already has column with label new_name"),
		},
		{
			name: "drop predicate (column) => column ~= /reg/",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.DropOpSpec{
						Predicate: interpreter.ResolvedFunction{
							Fn: &semantic.FunctionExpression{
								Block: &semantic.FunctionBlock{
									Parameters: &semantic.FunctionParameters{
										List: []*semantic.FunctionParameter{{Key: &semantic.Identifier{Name: "column"}}},
									},
									Body: &semantic.BinaryExpression{
										Operator: ast.RegexpMatchOperator,
										Left: &semantic.IdentifierExpression{
											Name: "column",
										},
										Right: &semantic.RegexpLiteral{
											Value: regexp.MustCompile(`server*`),
										},
									},
								},
							},
							Scope: valuestest.Scope(),
						},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "server1", Type: flux.TFloat},
					{Label: "local", Type: flux.TFloat},
					{Label: "server2", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{11.0, 12.0, 13.0},
					{21.0, 22.0, 23.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "local", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{2.0},
					{12.0},
					{22.0},
				},
			}},
		},
		{
			name: "keep predicate (column) => column ~= /reg/",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.KeepOpSpec{
						Predicate: interpreter.ResolvedFunction{
							Fn: &semantic.FunctionExpression{
								Block: &semantic.FunctionBlock{
									Parameters: &semantic.FunctionParameters{
										List: []*semantic.FunctionParameter{{Key: &semantic.Identifier{Name: "column"}}},
									},
									Body: &semantic.BinaryExpression{
										Operator: ast.RegexpMatchOperator,
										Left: &semantic.IdentifierExpression{
											Name: "column",
										},
										Right: &semantic.RegexpLiteral{
											Value: regexp.MustCompile(`server*`),
										},
									},
								},
							},
							Scope: valuestest.Scope(),
						},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "server1", Type: flux.TFloat},
					{Label: "local", Type: flux.TFloat},
					{Label: "server2", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{11.0, 12.0, 13.0},
					{21.0, 22.0, 23.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "server1", Type: flux.TFloat},
					{Label: "server2", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 3.0},
					{11.0, 13.0},
					{21.0, 23.0},
				},
			}},
		},
		{
			name: "drop and rename",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.DropOpSpec{
						Columns: []string{"server1", "server2"},
					},
					&universe.RenameOpSpec{
						Columns: map[string]string{
							"local": "localhost",
						},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "server1", Type: flux.TFloat},
					{Label: "local", Type: flux.TFloat},
					{Label: "server2", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{11.0, 12.0, 13.0},
					{21.0, 22.0, 23.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "localhost", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{2.0},
					{12.0},
					{22.0},
				},
			}},
		},
		{
			name: "drop no exist",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.DropOpSpec{
						Columns: []string{"no_exist"},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "server1", Type: flux.TFloat},
					{Label: "local", Type: flux.TFloat},
					{Label: "server2", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{11.0, 12.0, 13.0},
					{21.0, 22.0, 23.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "server1", Type: flux.TFloat},
					{Label: "local", Type: flux.TFloat},
					{Label: "server2", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{11.0, 12.0, 13.0},
					{21.0, 22.0, 23.0},
				},
			}},
		},
		{
			name: "rename no exist",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.RenameOpSpec{
						Columns: map[string]string{
							"no_exist": "noexist",
						},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "server1", Type: flux.TFloat},
					{Label: "local", Type: flux.TFloat},
					{Label: "server2", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{11.0, 12.0, 13.0},
					{21.0, 22.0, 23.0},
				},
			}},
			want:    []*executetest.Table(nil),
			wantErr: errors.New(`rename error: column "no_exist" doesn't exist`),
		},
		{
			name: "keep no exist",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.KeepOpSpec{
						Columns: []string{"no_exist"},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "server1", Type: flux.TFloat},
					{Label: "local", Type: flux.TFloat},
					{Label: "server2", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{11.0, 12.0, 13.0},
					{21.0, 22.0, 23.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta(nil),
				Data:    [][]interface{}(nil),
			}},
		},
		{
			name: "keep no exist along with all other columns",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.KeepOpSpec{
						Columns: []string{"no_exist", "server1", "local", "server2"},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "server1", Type: flux.TFloat},
					{Label: "local", Type: flux.TFloat},
					{Label: "server2", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{11.0, 12.0, 13.0},
					{21.0, 22.0, 23.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "server1", Type: flux.TFloat},
					{Label: "local", Type: flux.TFloat},
					{Label: "server2", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{11.0, 12.0, 13.0},
					{21.0, 22.0, 23.0},
				},
			}},
		},
		{
			name: "duplicate no exist",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.DuplicateOpSpec{
						Column: "no_exist",
						As:     "no_exist_2",
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "server1", Type: flux.TFloat},
					{Label: "local", Type: flux.TFloat},
					{Label: "server2", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{11.0, 12.0, 13.0},
					{21.0, 22.0, 23.0},
				},
			}},
			want:    []*executetest.Table(nil),
			wantErr: errors.New(`duplicate error: column "no_exist" doesn't exist`),
		},
		{
			name: "rename group key",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.RenameOpSpec{
						Columns: map[string]string{
							"1a": "1b",
							"2a": "2b",
							"3a": "3b",
						},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"1a"},
				ColMeta: []flux.ColMeta{
					{Label: "1a", Type: flux.TFloat},
					{Label: "2a", Type: flux.TFloat},
					{Label: "3a", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{1.0, 12.0, 13.0},
					{1.0, 22.0, 23.0},
				},
			}},
			want: []*executetest.Table{{
				KeyCols: []string{"1b"},
				ColMeta: []flux.ColMeta{
					{Label: "1b", Type: flux.TFloat},
					{Label: "2b", Type: flux.TFloat},
					{Label: "3b", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{1.0, 12.0, 13.0},
					{1.0, 22.0, 23.0},
				},
			}},
		},
		{
			name: "drop group key",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.DropOpSpec{
						Columns: []string{"2a"},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"2a"},
				ColMeta: []flux.ColMeta{
					{Label: "1a", Type: flux.TFloat},
					{Label: "2a", Type: flux.TFloat},
					{Label: "3a", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{11.0, 2.0, 13.0},
					{21.0, 2.0, 23.0},
				},
			}},
			want: []*executetest.Table{{
				KeyCols: []string(nil),
				ColMeta: []flux.ColMeta{
					{Label: "1a", Type: flux.TFloat},
					{Label: "3a", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 3.0},
					{11.0, 13.0},
					{21.0, 23.0},
				},
			}},
		},
		{
			name: "keep group key",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.KeepOpSpec{
						Columns: []string{"1a"},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"1a", "3a"},
				ColMeta: []flux.ColMeta{
					{Label: "1a", Type: flux.TFloat},
					{Label: "2a", Type: flux.TFloat},
					{Label: "3a", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{1.0, 12.0, 3.0},
					{1.0, 22.0, 3.0},
				},
			}},
			want: []*executetest.Table{{
				KeyCols: []string{"1a"},
				ColMeta: []flux.ColMeta{
					{Label: "1a", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0},
					{1.0},
					{1.0},
				},
			}},
		},
		{
			name: "duplicate group key",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.DuplicateOpSpec{
						Column: "1a",
						As:     "1a_1",
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"1a", "3a"},
				ColMeta: []flux.ColMeta{
					{Label: "1a", Type: flux.TFloat},
					{Label: "2a", Type: flux.TFloat},
					{Label: "3a", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{1.0, 12.0, 3.0},
					{1.0, 22.0, 3.0},
				},
			}},
			want: []*executetest.Table{{
				KeyCols: []string{"1a", "3a"},
				ColMeta: []flux.ColMeta{
					{Label: "1a", Type: flux.TFloat},
					{Label: "2a", Type: flux.TFloat},
					{Label: "3a", Type: flux.TFloat},
					{Label: "1a_1", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0, 1.0},
					{1.0, 12.0, 3.0, 1.0},
					{1.0, 22.0, 3.0, 1.0},
				},
			}},
		},
		{
			name: "keep with changing schema",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.KeepOpSpec{
						Columns: []string{"a"},
					},
				},
			},
			data: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"a"},
					ColMeta: []flux.ColMeta{
						{Label: "a", Type: flux.TInt},
						{Label: "b", Type: flux.TFloat},
						{Label: "c", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{int64(1), 10.0, 3.0},
						{int64(1), 12.0, 4.0},
						{int64(1), 22.0, 5.0},
					},
				},
				&executetest.Table{
					KeyCols: []string{"a"},
					ColMeta: []flux.ColMeta{
						{Label: "a", Type: flux.TInt},
						{Label: "b", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{int64(2), 11.0},
						{int64(2), 13.0},
						{int64(2), 23.0},
					},
				},
			},
			want: []*executetest.Table{
				{
					KeyCols: []string{"a"},
					ColMeta: []flux.ColMeta{{Label: "a", Type: flux.TInt}},
					Data: [][]interface{}{
						{int64(1)},
						{int64(1)},
						{int64(1)},
					},
				},
				{
					KeyCols: []string{"a"},
					ColMeta: []flux.ColMeta{{Label: "a", Type: flux.TInt}},
					Data: [][]interface{}{
						{int64(2)},
						{int64(2)},
						{int64(2)},
					},
				},
			},
		},
		{
			name: "rename with nulls",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.RenameOpSpec{
						Columns: map[string]string{
							"1a": "1b",
							"2a": "2b",
							"3a": "3b",
						},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "1a", Type: flux.TFloat},
					{Label: "2a", Type: flux.TFloat},
					{Label: "3a", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{nil, 2.0, 3.0},
					{11.0, 12.0, nil},
					{21.0, nil, nil},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "1b", Type: flux.TFloat},
					{Label: "2b", Type: flux.TFloat},
					{Label: "3b", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{nil, 2.0, 3.0},
					{11.0, 12.0, nil},
					{21.0, nil, nil},
				},
			}},
		},

		{
			name: "drop with nulls",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.DropOpSpec{
						Columns: []string{"a", "b"},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "a", Type: flux.TFloat},
					{Label: "b", Type: flux.TFloat},
					{Label: "c", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{nil, 2.0, 3.0},
					{nil, nil, nil},
					{nil, 22.0, nil},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "c", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{3.0},
					{nil},
					{nil},
				},
			}},
		},
		{
			name: "keep with nulls",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.KeepOpSpec{
						Columns: []string{"a"},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "a", Type: flux.TFloat},
					{Label: "b", Type: flux.TFloat},
					{Label: "c", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, nil},
					{nil, 12.0, 13.0},
					{21.0, nil, 23.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "a", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0},
					{nil},
					{21.0},
				},
			}},
		},
		{
			name: "duplicate with nulls",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.DuplicateOpSpec{
						Column: "a",
						As:     "a_1",
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "a", Type: flux.TFloat},
					{Label: "b", Type: flux.TFloat},
					{Label: "c", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{nil, nil, 3.0},
					{nil, 12.0, nil},
					{21.0, nil, 23.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "a", Type: flux.TFloat},
					{Label: "b", Type: flux.TFloat},
					{Label: "c", Type: flux.TFloat},
					{Label: "a_1", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{nil, nil, 3.0, nil},
					{nil, 12.0, nil, nil},
					{21.0, nil, 23.0, 21.0},
				},
			}},
		},
		{
			name: "rename group key with nulls",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.RenameOpSpec{
						Columns: map[string]string{
							"1a": "1b",
							"2a": "2b",
							"3a": "3b",
						},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"1a"},
				ColMeta: []flux.ColMeta{
					{Label: "1a", Type: flux.TFloat},
					{Label: "2a", Type: flux.TFloat},
					{Label: "3a", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{nil, 2.0, 3.0},
					{nil, 12.0, nil},
					{nil, nil, 23.0},
				},
			}},
			want: []*executetest.Table{{
				KeyCols: []string{"1b"},
				ColMeta: []flux.ColMeta{
					{Label: "1b", Type: flux.TFloat},
					{Label: "2b", Type: flux.TFloat},
					{Label: "3b", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{nil, 2.0, 3.0},
					{nil, 12.0, nil},
					{nil, nil, 23.0},
				},
			}},
		},
		{
			name: "drop group key with nulls",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.DropOpSpec{
						Columns: []string{"2a"},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"2a"},
				ColMeta: []flux.ColMeta{
					{Label: "1a", Type: flux.TFloat},
					{Label: "2a", Type: flux.TFloat},
					{Label: "3a", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, nil, 3.0},
					{nil, nil, 13.0},
					{21.0, nil, nil},
				},
			}},
			want: []*executetest.Table{{
				KeyCols: []string(nil),
				ColMeta: []flux.ColMeta{
					{Label: "1a", Type: flux.TFloat},
					{Label: "3a", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 3.0},
					{nil, 13.0},
					{21.0, nil},
				},
			}},
		},
		{
			name: "keep group key with nulls",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.KeepOpSpec{
						Columns: []string{"1a"},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"1a", "3a"},
				ColMeta: []flux.ColMeta{
					{Label: "1a", Type: flux.TFloat},
					{Label: "2a", Type: flux.TFloat},
					{Label: "3a", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{nil, 2.0, nil},
					{nil, 12.0, nil},
					{nil, 22.0, nil},
				},
			}},
			want: []*executetest.Table{{
				KeyCols: []string{"1a"},
				ColMeta: []flux.ColMeta{
					{Label: "1a", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{nil},
					{nil},
					{nil},
				},
			}},
		},
		{
			name: "duplicate group key with nulls",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.DuplicateOpSpec{
						Column: "3a",
						As:     "3a_1",
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"1a", "3a"},
				ColMeta: []flux.ColMeta{
					{Label: "1a", Type: flux.TFloat},
					{Label: "2a", Type: flux.TFloat},
					{Label: "3a", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, nil},
					{1.0, 12.0, nil},
					{1.0, 22.0, nil},
				},
			}},
			want: []*executetest.Table{{
				KeyCols: []string{"1a", "3a"},
				ColMeta: []flux.ColMeta{
					{Label: "1a", Type: flux.TFloat},
					{Label: "2a", Type: flux.TFloat},
					{Label: "3a", Type: flux.TFloat},
					{Label: "3a_1", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, nil, nil},
					{1.0, 12.0, nil, nil},
					{1.0, 22.0, nil, nil},
				},
			}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Skip("https://github.com/influxdata/flux/issues/2490")
			executetest.ProcessTestHelper(
				t,
				tc.data,
				tc.want,
				tc.wantErr,
				func(d execute.Dataset, c execute.TableBuilderCache) execute.Transformation {
					ctx := dependenciestest.Default().Inject(context.Background())
					tr, err := universe.NewSchemaMutationTransformation(ctx, tc.spec, d, c)
					if err != nil {
						t.Fatal(err)
					}
					return tr
				},
			)
		})
	}
}

// TODO: determine SchemaMutationProcedureSpec pushdown/rewrite rules
/*
func TestRenameDrop_PushDown(t *testing.T) {
	m1, _ := functions.NewRenameMutator(&functions.RenameOpSpec{
		Cols: map[string]string{},
	})

	root := &plan.Procedure{
		Spec: &functions.SchemaMutationProcedureSpec{
			Mutations: []functions.SchemaMutator{m1},
		},
	}

	m2, _ := functions.NewDropKeepMutator(&functions.DropOpSpec{
		Cols: []string{},
	})

	m3, _ := functions.NewDropKeepMutator(&functions.KeepOpSpec{
		Cols: []string{},
	})

	spec := &functions.SchemaMutationProcedureSpec{
		Mutations: []functions.SchemaMutator{m2, m3},
	}

	want := &plan.Procedure{
		Spec: &functions.SchemaMutationProcedureSpec{
			Mutations: []functions.SchemaMutator{m1, m2, m3},
		},
	}
	plantest.PhysicalPlan_PushDown_TestHelper(t, spec, root, false, want)
}
*/
