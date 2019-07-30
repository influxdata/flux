package execute_test

import (
	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/semantic"
	"testing"

	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
)

func TestNewRowListReduceFn(t *testing.T) {
	testCases := []struct {
		name       string
		f          func() (*execute.RowListReduceFn, error)
		data       *executetest.Table
		want       [][]interface{}
		prepareErr error
	}{
		{
			name: "[1].diff / [0]._value",
			f: func() (*execute.RowListReduceFn, error) {
				return execute.NewRowListReduceFn(&semantic.FunctionExpression{
					Block: &semantic.FunctionBlock{
						Parameters: &semantic.FunctionParameters{
							List: []*semantic.FunctionParameter{{Key: &semantic.Identifier{Name: "rows"}}},
						},
						Body: &semantic.ObjectExpression{
							Properties: []*semantic.Property{
								{
									Key: &semantic.StringLiteral{Value: "_value"},
									Value: &semantic.BinaryExpression{
										Operator: ast.DivisionOperator,
										Left: &semantic.MemberExpression{
											Object: &semantic.IndexExpression{
												Array: &semantic.IdentifierExpression{Name: "rows"},
												Index: &semantic.IntegerLiteral{Value: 1},
											},
											Property: "diff",
										},
										Right: &semantic.MemberExpression{
											Object: &semantic.IndexExpression{
												Array: &semantic.IdentifierExpression{Name: "rows"},
												Index: &semantic.IntegerLiteral{Value: 0},
											},
											Property: "_value",
										},
									},
								},
							},
						},
					},
				})
			},
			data: &executetest.Table{
				ColMeta:[]flux.ColMeta{
					{Label: "_value", Type: flux.TInt},
					{Label: "diff", Type: flux.TInt},
				},
				Data: [][]interface{}{
					{int64(2), int64(5)},
					{int64(10), int64(6)},
				},
			},
			want: [][]interface{}{
				{"_value", int64(3)},
			},
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

			got := make([]*execute.Record, 0, len(tc.data.Data))
			d := make([][]interface{}, 0)
			if err := tc.data.Do(func(cr flux.ColReader) error {
				for i := 0; i < cr.Len(); i++ {

					d = append(d, make([]interface{}, 0))

					for j := 0; j < len(tc.data.Cols()); j++ {
						d[i] = append(d[i], execute.ValueForRow(cr, i, j))
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