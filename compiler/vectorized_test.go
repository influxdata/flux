package compiler_test

import (
	"context"
	"fmt"
	"math"
	"testing"

	arrow "github.com/apache/arrow/go/v7/arrow/memory"
	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/compiler"
	"github.com/influxdata/flux/execute/executetest"
	fluxfeature "github.com/influxdata/flux/internal/feature"
	"github.com/influxdata/flux/internal/pkg/feature"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

func vectorizedObjectFromMap(mp map[string]interface{}, mem memory.Allocator) values.Object {
	obj := make(map[string]values.Value)
	for k, v := range mp {
		switch s := v.(type) {
		case []interface{}:
			obj[k] = values.NewVectorFromElements(mem, s...)
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
//     2. The number of bytes allocated is 0 once evaluation is complete
//        and values are released
//     3. Only certain function expressions are vectorized when invoking the
//        analyzer from go code. The criteria for supported expressions may
//        change in the future, but right now we only support trivial identity
//        functions (i.e., those in the form of `(r) => ({a: r.a})`, or something
//        similar)
func TestVectorizedFns(t *testing.T) {
	type TestCase struct {
		name         string
		fn           string
		vectorizable bool
		inType       semantic.MonoType
		input        map[string]interface{}
		want         map[string]interface{}
		skipComp     bool
		flagger      executetest.TestFlagger
	}

	testCases := []TestCase{
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
			name:         "addition expression nested",
			fn:           `(r) => ({c: r.a + r.b + r.a})`,
			vectorizable: true,
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
				"c": []interface{}{int64(4)},
			},
			flagger: executetest.TestFlagger{},
		},
		{
			name:         "addition expression multiple",
			fn:           `(r) => ({x: r.a + r.b, y: r.c + r.d})`,
			vectorizable: true,
			inType: semantic.NewObjectType([]semantic.PropertyType{
				{Key: []byte("r"), Value: semantic.NewObjectType([]semantic.PropertyType{
					{Key: []byte("a"), Value: semantic.NewVectorType(semantic.BasicInt)},
					{Key: []byte("b"), Value: semantic.NewVectorType(semantic.BasicInt)},
					{Key: []byte("c"), Value: semantic.NewVectorType(semantic.BasicFloat)},
					{Key: []byte("d"), Value: semantic.NewVectorType(semantic.BasicFloat)},
				})},
			}),
			input: map[string]interface{}{
				"r": map[string]interface{}{
					"a": []interface{}{int64(1)},
					"b": []interface{}{int64(2)},
					"c": []interface{}{49.0},
					"d": []interface{}{51.0},
				},
			},
			want: map[string]interface{}{
				"x": []interface{}{int64(3)},
				"y": []interface{}{100.0},
			},
			flagger: executetest.TestFlagger{},
		},
		{
			name:         "no binary expressions without feature flag",
			fn:           `(r) => ({c: r.a / r.b})`,
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

	additionTests := []struct {
		inType semantic.MonoType
		input  map[string]interface{}
		want   map[string]interface{}
	}{
		{
			inType: semantic.BasicInt,
			input: map[string]interface{}{
				"r": map[string]interface{}{
					"a": []interface{}{int64(1), int64(3)},
					"b": []interface{}{int64(2), int64(4)},
				},
			},
			want: map[string]interface{}{
				"c": []interface{}{int64(3), int64(7)},
			},
		},
		{
			inType: semantic.BasicUint,
			input: map[string]interface{}{
				"r": map[string]interface{}{
					"a": []interface{}{uint64(1), uint64(3)},
					"b": []interface{}{uint64(2), uint64(4)},
				},
			},
			want: map[string]interface{}{
				"c": []interface{}{uint64(3), uint64(7)},
			},
		},
		{
			inType: semantic.BasicFloat,
			input: map[string]interface{}{
				"r": map[string]interface{}{
					"a": []interface{}{1.0, 3.0},
					"b": []interface{}{2.0, 4.0},
				},
			},
			want: map[string]interface{}{
				"c": []interface{}{3.0, 7.0},
			},
		},
		{
			inType: semantic.BasicString,
			input: map[string]interface{}{
				"r": map[string]interface{}{
					"a": []interface{}{"a", "c"},
					"b": []interface{}{"b", "d"},
				},
			},
			want: map[string]interface{}{
				"c": []interface{}{"ab", "cd"},
			},
		},
	}

	for _, test := range additionTests {
		testCases = append(testCases, TestCase{
			name:         fmt.Sprintf("addition expression %s", test.inType.String()),
			fn:           `(r) => ({c: r.a + r.b})`,
			vectorizable: true,
			inType: semantic.NewObjectType([]semantic.PropertyType{
				{Key: []byte("r"), Value: semantic.NewObjectType([]semantic.PropertyType{
					{Key: []byte("a"), Value: semantic.NewVectorType(test.inType)},
					{Key: []byte("b"), Value: semantic.NewVectorType(test.inType)},
				})},
			}),
			input: test.input,
			want:  test.want,

			flagger: executetest.TestFlagger{},
		})
	}

	operatorTests := []struct {
		operator  string
		input     [][2]int64
		transform func(int64, int64) interface{}
	}{
		{
			operator: "-",
			input: [][2]int64{
				{1, 2},
				{10, 5},
				{112487, 66547},
			},
			transform: func(l, r int64) interface{} {
				return l - r
			},
		},
		{
			operator: "*",
			input: [][2]int64{
				{1, 2},
				{10, 5},
				{112487, 66547},
			},
			transform: func(l, r int64) interface{} {
				return l * r
			},
		},
		{
			operator: "/",
			input: [][2]int64{
				{1, 2},
				{10, 5},
				{112487, 66547},
			},
			transform: func(l, r int64) interface{} {
				return l / r
			},
		},
		{
			operator: "%",
			input: [][2]int64{
				{1, 2},
				{10, 5},
				{112487, 66547},
			},
			transform: func(l, r int64) interface{} {
				return l % r
			},
		},
		{
			operator: "^",
			input: [][2]int64{
				{1, 2},
				{10, 5},
				{112487, 66547},
			},
			transform: func(l, r int64) interface{} {
				return math.Pow(float64(l), float64(r))
			},
		},
	}
	for _, test := range operatorTests {
		a := []interface{}{}
		b := []interface{}{}
		output := []interface{}{}
		for _, item := range test.input {
			a = append(a, item[0])
			b = append(b, item[1])
			output = append(output, test.transform(item[0], item[1]))
		}

		testCases = append(testCases, TestCase{
			name:         fmt.Sprintf("%s expression %s", test.operator, semantic.BasicInt.String()),
			fn:           fmt.Sprintf("(r) => ({c: r.a %s r.b})", test.operator),
			vectorizable: true,
			inType: semantic.NewObjectType([]semantic.PropertyType{
				{Key: []byte("r"), Value: semantic.NewObjectType([]semantic.PropertyType{
					{Key: []byte("a"), Value: semantic.NewVectorType(semantic.BasicInt)},
					{Key: []byte("b"), Value: semantic.NewVectorType(semantic.BasicInt)},
				})},
			}),
			input: map[string]interface{}{
				"r": map[string]interface{}{
					"a": a,
					"b": b,
				},
			},
			want: map[string]interface{}{
				"c": output,
			},

			flagger: executetest.TestFlagger{
				fluxfeature.VectorizeOperators().Key(): true,
			},
		})
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			checked := arrow.NewCheckedAllocator(memory.DefaultAllocator)
			mem := memory.NewResourceAllocator(checked)

			flagger := tc.flagger
			if flagger == nil {
				flagger = executetest.TestFlagger{}
			}
			flagger[fluxfeature.VectorizedMap().Key()] = true
			ctx := context.Background()
			ctx = feature.Inject(
				ctx,
				flagger,
			)
			ctx = memory.WithAllocator(ctx, mem)

			pkg, err := runtime.AnalyzeSource(ctx, tc.fn)
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
			input := vectorizedObjectFromMap(tc.input, mem)
			got, err := f.Eval(ctx, input)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			want := vectorizedObjectFromMap(tc.want, memory.NewResourceAllocator(nil))
			if !cmp.Equal(want, got, CmpOptions...) {
				t.Errorf("unexpected value -want/+got\n%s", cmp.Diff(want, got, CmpOptions...))
			}

			got.Release()
			input.Release()

			checked.AssertSize(t, 0)
		})
	}
}
