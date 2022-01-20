package compiler_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/compiler"
	"github.com/influxdata/flux/internal/arrowutil"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

func vectorizedObjectFromMap(mp map[string]interface{}, mem *memory.Allocator) values.Object {
	obj := make(map[string]values.Value)
	for k, v := range mp {
		switch s := v.(type) {
		case []interface{}:
			obj[k] = arrowutil.NewVectorFromElements(mem, s...)
		case map[string]interface{}:
			obj[k] = vectorizedObjectFromMap(v.(map[string]interface{}), mem)
		default:
			panic("bad input to vectorizedObjectFromMap")
		}
	}
	return values.NewObjectWithValues(obj)
}

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
		allocated    int64
		maxAllocated int64
		inType       semantic.MonoType
		input        map[string]interface{}
		want         map[string]interface{}
		skipComp     bool
	}{
		{
			name:         "field access",
			fn:           `(r) => ({c: r.a, d: r.b})`,
			vectorizable: true,
			allocated:    256, // This should be 128 (2 vectorized fields, each containing 1 int64)
			maxAllocated: 448,
			inType: semantic.NewObjectType([]semantic.PropertyType{
				{Key: []byte("r"), Value: semantic.NewObjectType([]semantic.PropertyType{
					{Key: []byte("a"), Value: semantic.NewVectorType(semantic.BasicInt)},
					{Key: []byte("b"), Value: semantic.NewVectorType(semantic.BasicInt)},
				})},
			}),
			input: map[string]interface{}{
				"r": map[string]interface{}{
					"a": []interface{}{int64(1)},
					"b": []interface{}{int64(2)},
				},
			},
			want: map[string]interface{}{
				"c": []interface{}{int64(1)},
				"d": []interface{}{int64(2)},
			},
		},
		{
			name:         "extend record",
			fn:           `(r) => ({r with b: r.a})`,
			vectorizable: true,
			allocated:    128, // This should be 64 (1 vectorized field that contains 1 float64)
			maxAllocated: 320,
			inType: semantic.NewObjectType([]semantic.PropertyType{
				{Key: []byte("r"), Value: semantic.NewObjectType([]semantic.PropertyType{
					{Key: []byte("a"), Value: semantic.NewVectorType(semantic.BasicFloat)},
				})},
			}),
			input: map[string]interface{}{
				"r": map[string]interface{}{
					"a": []interface{}{1.2},
				},
			},
			want: map[string]interface{}{
				"a": []interface{}{1.2},
				"b": []interface{}{1.2},
			},
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
			mem := memory.Allocator{}
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
			input := vectorizedObjectFromMap(tc.input, &mem)
			got, err := f.Eval(context.TODO(), input)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			want := vectorizedObjectFromMap(tc.want, &memory.Allocator{})
			if !cmp.Equal(want, got, CmpOptions...) {
				t.Errorf("unexpected value -want/+got\n%s", cmp.Diff(want, got, CmpOptions...))
			}

			if want, got := tc.allocated, mem.Allocated(); want != got {
				t.Fatalf("unexpected allocated count -want/+got\n\t- %d\n\t+ %d", want, got)
			}
			if want, got := tc.maxAllocated, mem.MaxAllocated(); want != got {
				t.Fatalf("unexpected max allocated count -want/+got\n\t- %d\n\t+ %d", want, got)
			}
		})
	}
}
