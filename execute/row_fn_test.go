package execute_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/compiler"
	"github.com/influxdata/flux/dependencies/dependenciestest"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/semantic/semantictest"
	"github.com/influxdata/flux/values"
)

var CmpOptions = semantictest.CmpOptions

func prelude() compiler.Scope {
	return compiler.ToScope(flux.Prelude())
}

func createRecord(row []interface{}) (*execute.Record, error) {
	if len(row) == 0 {
		return execute.NewRecord(semantic.Invalid), nil
	}

	if len(row)%2 != 0 {
		return nil, errors.New("row must contain couples")
	}

	r := execute.NewRecord(semantic.Object)
	for i := 0; i < len(row); i += 2 {
		if key, ok := row[i].(string); !ok {
			return nil, fmt.Errorf("keys must be strings: %v", row[i])
		} else {
			val := values.New(row[i+1])
			r.Set(key, val)
		}
	}

	return r, nil
}

func TestRowMapFn_Eval(t *testing.T) {
	testCases := []struct {
		name       string
		f          func() (*execute.RowMapFn, error)
		data       *executetest.Table
		want       [][]interface{}
		prepareErr error
	}{
		{
			name: "_value + 1.0, tag + 'b'",
			f: func() (*execute.RowMapFn, error) {
				return execute.NewRowMapFn(&semantic.FunctionExpression{
					Block: &semantic.FunctionBlock{
						Parameters: &semantic.FunctionParameters{
							List: []*semantic.FunctionParameter{{Key: &semantic.Identifier{Name: "r"}}},
						},
						Body: &semantic.ObjectExpression{
							Properties: []*semantic.Property{
								{
									Key: &semantic.StringLiteral{Value: "_value"},
									Value: &semantic.BinaryExpression{
										Operator: ast.AdditionOperator,
										Left: &semantic.MemberExpression{
											Object: &semantic.IdentifierExpression{
												Name: "r",
											},
											Property: "_value",
										},
										Right: &semantic.FloatLiteral{Value: 1.0},
									},
								},
								{
									Key: &semantic.StringLiteral{Value: "tag"},
									Value: &semantic.BinaryExpression{
										Operator: ast.AdditionOperator,
										Left: &semantic.MemberExpression{
											Object: &semantic.IdentifierExpression{
												Name: "r",
											},
											Property: "tag",
										},
										Right: &semantic.StringLiteral{Value: "b"},
									},
								},
							},
						},
					},
				}, prelude())
			},
			data: &executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "tag", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(1), 1.0, "a"},
					{execute.Time(2), 2.0, "a"},
					{execute.Time(3), 3.0, "a"},
					{execute.Time(4), 4.0, "a"},
					{execute.Time(5), 5.0, "a"},
					{execute.Time(6), 6.0, "a"},
				},
			},
			want: [][]interface{}{
				{"_value", 2.0, "tag", "ab"},
				{"_value", 3.0, "tag", "ab"},
				{"_value", 4.0, "tag", "ab"},
				{"_value", 5.0, "tag", "ab"},
				{"_value", 6.0, "tag", "ab"},
				{"_value", 7.0, "tag", "ab"},
			},
		},
		{
			name: "_value - 3.0 with nulls",
			f: func() (*execute.RowMapFn, error) {
				return execute.NewRowMapFn(&semantic.FunctionExpression{
					Block: &semantic.FunctionBlock{
						Parameters: &semantic.FunctionParameters{
							List: []*semantic.FunctionParameter{{Key: &semantic.Identifier{Name: "r"}}},
						},
						Body: &semantic.ObjectExpression{
							Properties: []*semantic.Property{
								{
									Key: &semantic.StringLiteral{Value: "_value"},
									Value: &semantic.BinaryExpression{
										Operator: ast.SubtractionOperator,
										Left: &semantic.MemberExpression{
											Object: &semantic.IdentifierExpression{
												Name: "r",
											},
											Property: "_value",
										},
										Right: &semantic.FloatLiteral{Value: 3.0},
									},
								},
							},
						},
					},
				}, prelude())
			},
			data: &executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "tag", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(1), 1.0, "a"},
					{execute.Time(2), 2.0, "a"},
					{execute.Time(3), nil, "a"},
					{execute.Time(4), nil, "a"},
					{execute.Time(5), 5.0, "a"},
					{execute.Time(6), nil, "a"},
				},
			},
			want: [][]interface{}{
				{"_value", -2.0},
				{"_value", -1.0},
				{"_value", nil},
				{"_value", nil},
				{"_value", 2.0},
				{"_value", nil},
			},
		},
		{
			name: "error not returning object",
			f: func() (*execute.RowMapFn, error) {
				return execute.NewRowMapFn(&semantic.FunctionExpression{
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
							Right: &semantic.FloatLiteral{Value: 3.0},
						},
					},
				}, prelude())
			},
			data: &executetest.Table{
				ColMeta: []flux.ColMeta{
					// This is needed because the function accesses `_value` on `r`.
					// Otherwise, it would give a different error than expected.
					{Label: "_value", Type: flux.TFloat},
				},
			},
			prepareErr: fmt.Errorf("map function must return an object, got float"),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			f, err := tc.f()
			if err != nil {
				t.Fatal(err)
			}
			err = f.Prepare(tc.data.ColMeta)
			if err != nil {
				if tc.prepareErr != nil {
					if !cmp.Equal(tc.prepareErr.Error(), err.Error()) {
						t.Fatalf("unexpected prepare error -want/+got\n%s", cmp.Diff(tc.prepareErr.Error(), err.Error()))
					}
					return
				}
				t.Fatal(err)
			} else if tc.prepareErr != nil {
				t.Fatal("expected prepare error, got none")
			}

			// convert tc.want
			want := make([]*execute.Record, len(tc.want))
			for i := 0; i < len(tc.want); i++ {
				r, err := createRecord(tc.want[i])
				if err != nil {
					t.Fatal(err)
				}
				want[i] = r
			}

			ctx := dependenciestest.Default().Inject(context.Background())
			got := make([]*execute.Record, 0, len(tc.data.Data))
			if err := tc.data.Do(func(cr flux.ColReader) error {
				for i := 0; i < cr.Len(); i++ {
					obj, err := f.Eval(ctx, i, cr)
					if err != nil {
						got = append(got, execute.NewRecord(semantic.Invalid))
					} else {
						r := execute.NewRecord(semantic.Object)
						obj.Range(func(k string, v values.Value) {
							r.Set(k, v)
						})
						got = append(got, r)
					}
				}

				return nil
			}); err != nil {
				t.Fatal(err)
			}

			if !cmp.Equal(want, got, CmpOptions...) {
				t.Errorf("unexpected result -want/+got\n%s", cmp.Diff(want, got, CmpOptions...))
			}
		})
	}
}

func TestRowPredicateFn_EvalRow(t *testing.T) {
	gt2F := func() (*execute.RowPredicateFn, error) {
		return execute.NewRowPredicateFn(&semantic.FunctionExpression{
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
					Right: &semantic.FloatLiteral{Value: 2.0},
				},
			},
		}, prelude())
	}

	testCases := []struct {
		name string
		data *executetest.Table
		want []bool
	}{
		{
			name: "gt 2.0",
			data: &executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 1.0},
					{execute.Time(2), 2.0},
					{execute.Time(3), 3.0},
					{execute.Time(4), 4.0},
					{execute.Time(5), 5.0},
					{execute.Time(6), 6.0},
				},
			},
			want: []bool{
				false,
				false,
				true,
				true,
				true,
				true,
			},
		},
		{
			name: "gt 2.0 with nulls",
			data: &executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 1.0},
					{execute.Time(2), 2.0},
					{execute.Time(3), nil},
					{execute.Time(4), nil},
					{execute.Time(5), 5.0},
					{execute.Time(6), nil},
				},
			},
			want: []bool{
				false,
				false,
				false,
				false,
				true,
				false,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			f, err := gt2F()
			if err != nil {
				t.Fatal(err)
			}
			err = f.Prepare(tc.data.ColMeta)
			if err != nil {
				t.Fatal(err)
			}
			ctx := dependenciestest.Default().Inject(context.Background())
			got := make([]bool, 0, len(tc.data.Data))
			tc.data.Do(func(cr flux.ColReader) error {
				for i := 0; i < cr.Len(); i++ {
					b, err := f.EvalRow(ctx, i, cr)
					if err == nil {
						got = append(got, b)
					}
				}

				return nil
			})

			if !cmp.Equal(tc.want, got, CmpOptions...) {
				t.Errorf("unexpected result -want/+got\n%s", cmp.Diff(tc.want, got, CmpOptions...))
			}
		})
	}
}
