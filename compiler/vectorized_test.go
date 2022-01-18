package compiler_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/compiler"
	"github.com/influxdata/flux/internal/arrowutil"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

// Check that:
//     1. Vectorized inputs yield vectorized outputs when compiled and evaluated
//     2. Only certain function expressions are vectorized when invoking the
//        analyzer from go code. The criteria for supported expressions may
//        change in the future, but right now we only support trivial identity
//        functions (i.e., those in the form of `(r) => ({a: r.a})`, or something
//        similar).
func TestVectorizedFns(t *testing.T) {
	testCases := []struct {
		name         string
		fn           string
		vectorizable bool
		inType       semantic.MonoType
		input        values.Object
		want         values.Value
		skipComp     bool
	}{
		{
			name:         "field access",
			fn:           `(r) => ({c: r.a, d: r.b})`,
			vectorizable: true,
			inType: semantic.NewObjectType([]semantic.PropertyType{
				{Key: []byte("r"), Value: semantic.NewObjectType([]semantic.PropertyType{
					{Key: []byte("a"), Value: semantic.NewVectorType(semantic.BasicInt)},
					{Key: []byte("b"), Value: semantic.NewVectorType(semantic.BasicInt)},
				})},
			}),
			input: values.NewObjectWithValues(map[string]values.Value{
				"r": values.NewObjectWithValues(map[string]values.Value{
					"a": arrowutil.NewVectorFromElements(int64(1)),
					"b": arrowutil.NewVectorFromElements(int64(2)),
				}),
			}),
			want: values.NewObjectWithValues(map[string]values.Value{
				"c": arrowutil.NewVectorFromElements(int64(1)),
				"d": arrowutil.NewVectorFromElements(int64(2)),
			}),
		},
		{
			name:         "extend record",
			fn:           `(r) => ({r with b: r.a})`,
			vectorizable: true,
			inType: semantic.NewObjectType([]semantic.PropertyType{
				{Key: []byte("r"), Value: semantic.NewObjectType([]semantic.PropertyType{
					{Key: []byte("a"), Value: semantic.NewVectorType(semantic.BasicFloat)},
				})},
			}),
			input: values.NewObjectWithValues(map[string]values.Value{
				"r": values.NewObjectWithValues(map[string]values.Value{
					"a": arrowutil.NewVectorFromElements(1.2),
				}),
			}),
			want: values.NewObjectWithValues(map[string]values.Value{
				"a": arrowutil.NewVectorFromElements(1.2),
				"b": arrowutil.NewVectorFromElements(1.2),
			}),
		},
		{
			name:         "no binary expressions",
			fn:           `(r) => ({c: r.a + r.b})`,
			vectorizable: false,
			skipComp:     true,
		},
		{
			name:         "no literals",
			fn:           `(r) => ({r with c: "count"})`,
			vectorizable: false,
			skipComp:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pkg, err := runtime.AnalyzeSource(tc.fn)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			stmt := pkg.Files[0].Body[0].(*semantic.ExpressionStatement)
			fn := stmt.Expression.(*semantic.FunctionExpression)

			if tc.vectorizable {
				if fn.Vectorized == nil {
					t.Fatal("Expected to find vectorized node, but found none")
				}
			} else {
				if fn.Vectorized != nil {
					t.Fatal("Vectorized node is populated when it should be nil")
				}
			}

			if tc.skipComp {
				return
			}

			f, err := compiler.Compile(nil, fn, tc.inType)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			got, err := f.Eval(context.TODO(), tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !cmp.Equal(tc.want, got, CmpOptions...) {
				t.Errorf("unexpected value -want/+got\n%s", cmp.Diff(tc.want, got, CmpOptions...))
			}
		})
	}
}
