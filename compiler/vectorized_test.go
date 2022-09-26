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
//  1. Vectorized inputs yield vectorized outputs when compiled and evaluated
//  2. The number of bytes allocated is 0 once evaluation is complete
//     and values are released
//  3. Only certain function expressions are vectorized when invoking the
//     analyzer from go code. The criteria for supported expressions may
//     change in the future, but right now we only support trivial identity
//     functions (i.e., those in the form of `(r) => ({a: r.a})`, or something
//     similar)
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
			name:         "string literals",
			fn:           `(r) => ({r with c: "count"})`,
			vectorizable: true,
			skipComp:     true,
		},
		{
			name:         "int literals",
			fn:           `(r) => ({r with c: 123})`,
			vectorizable: true,
			skipComp:     true,
		},
		{
			name:         "float literals",
			fn:           `(r) => ({r with c: 1.23})`,
			vectorizable: true,
			skipComp:     true,
		},
		{
			name:         "time literals",
			fn:           `(r) => ({r with c: 2022-07-07})`,
			vectorizable: true,
			skipComp:     true,
		},
		{
			name:         "bool literals",
			fn:           `(r) => ({r with c: true, d: false})`,
			vectorizable: true,
			skipComp:     true,
		},
		{
			name:         "no duration literals",
			fn:           `(r) => ({r with c: 1h})`,
			vectorizable: false,
			skipComp:     true,
		},
		{
			name:         "equality operators",
			fn:           `(r) => ({r with c: 2 > 1})`,
			vectorizable: true,
			skipComp:     true,
		},
		{
			name:         "conditional expressions",
			fn:           `(r) => ({ r with c: if r.cond then 1 else 0 })`,
			vectorizable: true,
			skipComp:     true,
		},
		{
			name:         "call expressions float",
			fn:           `(r) => ({ r with c: float(v: 1) })`,
			vectorizable: true,
			skipComp:     true,
		},
		{
			name:         "unary expressions equality",
			fn:           "(r) => ({ r with c: exists r._a, d: not r._b })",
			vectorizable: true,
			skipComp:     true,
		},
		{
			name:         "unary expressions arithmetic",
			fn:           "(r) => ({ r with c: -r._a, d: +r._b })",
			vectorizable: true,
			skipComp:     true,
		},
		{
			name:         "logical expressions",
			fn:           "(r) => ({ r with c: r._a and r._b, d: r._a or r._b })",
			vectorizable: true,
			skipComp:     true,
		},
	}

	logicalTests := []struct {
		name  string
		input map[string]interface{}
		want  map[string]interface{}
	}{
		{
			name: "null handling",
			input: map[string]interface{}{
				"r": map[string]interface{}{
					"a": []interface{}{true, true, false, nil, false, nil},
					"b": []interface{}{false, nil, true, true, nil, nil},
				},
			},
			want: map[string]interface{}{
				"c": []interface{}{false, nil, false, nil, false, nil},
				"d": []interface{}{true, true, true, true, nil, nil},
			},
		},
		{
			name: "all true LHS",
			input: map[string]interface{}{
				"r": map[string]interface{}{
					"a": []interface{}{true, true, true, true},
					"b": []interface{}{false, false, true, false},
				},
			},
			want: map[string]interface{}{
				"c": []interface{}{false, false, true, false},
				"d": []interface{}{true, true, true, true},
			},
		},
		{
			name: "all false both sides",
			input: map[string]interface{}{
				"r": map[string]interface{}{
					"a": []interface{}{false, false, false, false},
					"b": []interface{}{false, false, false, false},
				},
			},
			want: map[string]interface{}{
				"c": []interface{}{false, false, false, false},
				"d": []interface{}{false, false, false, false},
			},
		},
		{
			name: "all true RHS",
			input: map[string]interface{}{
				"r": map[string]interface{}{
					"a": []interface{}{false, true, true, false},
					"b": []interface{}{true, true, true, true},
				},
			},
			want: map[string]interface{}{
				"c": []interface{}{false, true, true, false},
				"d": []interface{}{true, true, true, true},
			},
		},
	}

	for _, test := range logicalTests {
		testCases = append(testCases, TestCase{
			name:         fmt.Sprintf("logical expression %s", test.name),
			fn:           `(r) => ({c: r.a and r.b, d: r.a or r.b})`,
			vectorizable: true,
			inType: semantic.NewObjectType([]semantic.PropertyType{
				{Key: []byte("r"), Value: semantic.NewObjectType([]semantic.PropertyType{
					{Key: []byte("a"), Value: semantic.NewVectorType(semantic.BasicBool)},
					{Key: []byte("b"), Value: semantic.NewVectorType(semantic.BasicBool)},
				})},
			}),
			input: test.input,
			want:  test.want,

			flagger: executetest.TestFlagger{},
		})
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

	unaryOperatorTests := []struct {
		operator  string
		inType    semantic.MonoType
		input     []interface{}
		transform func(interface{}) interface{}
	}{
		{
			operator: "-",
			inType:   semantic.BasicInt,
			input: []interface{}{
				int64(1),
				int64(10),
				int64(112487),
				nil,
			},
			transform: func(arg interface{}) interface{} {
				if arg == nil {
					return nil
				}
				return -arg.(int64)
			},
		},
		{
			operator: "+",
			inType:   semantic.BasicInt,
			input: []interface{}{
				int64(1),
				int64(10),
				int64(112487),
				nil,
			},
			transform: func(arg interface{}) interface{} {
				if arg == nil {
					return nil
				}
				// unary add "does nothing"
				return arg
			},
		},
		{

			operator: "exists",
			inType:   semantic.BasicInt,
			input: []interface{}{
				int64(1),
				nil,
			},
			transform: func(arg interface{}) interface{} {
				return arg != nil
			},
		},
		{
			operator: "not",
			inType:   semantic.BasicBool,
			input: []interface{}{
				true,
				false,
				nil,
			},
			transform: func(arg interface{}) interface{} {
				if arg == nil {
					return nil
				}
				return !arg.(bool)
			},
		},
	}

	for _, test := range unaryOperatorTests {
		a := []interface{}{}
		output := []interface{}{}
		for _, item := range test.input {
			a = append(a, item)
			output = append(output, test.transform(item))
		}

		testCases = append(testCases, TestCase{
			name:         fmt.Sprintf("unary %s expression %s", test.operator, test.inType.String()),
			fn:           fmt.Sprintf("(r) => ({b: %s r.a})", test.operator),
			vectorizable: true,
			skipComp:     false,
			inType: semantic.NewObjectType([]semantic.PropertyType{
				{Key: []byte("r"), Value: semantic.NewObjectType([]semantic.PropertyType{
					{Key: []byte("a"), Value: semantic.NewVectorType(test.inType)},
				})},
			}),
			input: map[string]interface{}{
				"r": map[string]interface{}{
					"a": a,
				},
			},
			want: map[string]interface{}{
				"b": output,
			},

			flagger: executetest.TestFlagger{},
		})
	}

	binaryOperatorTests := []struct {
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
		{
			operator: "==",
			input: [][2]int64{
				{1, 2},
				{5, 5},
				{112487, 66547},
			},
			transform: func(l, r int64) interface{} {
				return l == r
			},
		},
		{
			operator: "!=",
			input: [][2]int64{
				{1, 2},
				{5, 5},
				{112487, 66547},
			},
			transform: func(l, r int64) interface{} {
				return l != r
			},
		},
		{
			operator: "<",
			input: [][2]int64{
				{1, 2},
				{5, 5},
				{112487, 66547},
			},
			transform: func(l, r int64) interface{} {
				return l < r
			},
		},
		{
			operator: "<=",
			input: [][2]int64{
				{1, 2},
				{5, 5},
				{112487, 66547},
			},
			transform: func(l, r int64) interface{} {
				return l <= r
			},
		},
		{
			operator: ">",
			input: [][2]int64{
				{1, 2},
				{5, 5},
				{112487, 66547},
			},
			transform: func(l, r int64) interface{} {
				return l > r
			},
		}, {
			operator: ">=",
			input: [][2]int64{
				{1, 2},
				{5, 5},
				{112487, 66547},
			},
			transform: func(l, r int64) interface{} {
				return l >= r
			},
		},
	}

	for _, test := range binaryOperatorTests {
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

			flagger: executetest.TestFlagger{},
		})
	}

	conditionalTests := []struct {
		name   string // a default name will be generated based on the input type, but can optionally be overridden
		inType semantic.MonoType
		input  map[string]interface{}
		want   map[string]interface{}
	}{
		{
			inType: semantic.BasicInt,
			input: map[string]interface{}{
				"r": map[string]interface{}{
					"a":    []interface{}{int64(1), int64(3)},
					"b":    []interface{}{int64(2), int64(4)},
					"cond": []interface{}{true, false},
				},
			},
			want: map[string]interface{}{
				"c": []interface{}{int64(1), int64(4)},
			},
		},
		{
			inType: semantic.BasicUint,
			input: map[string]interface{}{
				"r": map[string]interface{}{
					"a":    []interface{}{uint64(1), uint64(3)},
					"b":    []interface{}{uint64(2), uint64(4)},
					"cond": []interface{}{true, false},
				},
			},
			want: map[string]interface{}{
				"c": []interface{}{uint64(1), uint64(4)},
			},
		},
		{
			inType: semantic.BasicFloat,
			input: map[string]interface{}{
				"r": map[string]interface{}{
					"a":    []interface{}{1.0, 3.0},
					"b":    []interface{}{2.0, 4.0},
					"cond": []interface{}{true, false},
				},
			},
			want: map[string]interface{}{
				"c": []interface{}{1.0, 4.0},
			},
		},
		{
			inType: semantic.BasicString,
			input: map[string]interface{}{
				"r": map[string]interface{}{
					"a":    []interface{}{"a", "c"},
					"b":    []interface{}{"b", "d"},
					"cond": []interface{}{true, false},
				},
			},
			want: map[string]interface{}{
				"c": []interface{}{"a", "d"},
			},
		},
		{
			name:   "conditional expression nil test",
			inType: semantic.BasicInt,
			input: map[string]interface{}{
				"r": map[string]interface{}{
					"a":    []interface{}{int64(1), int64(3)},
					"b":    []interface{}{int64(2), int64(4)},
					"cond": []interface{}{nil, false},
				},
			},
			want: map[string]interface{}{
				// nil is considered "false" so the 1st item comes from `b`, the alternate
				"c": []interface{}{int64(2), int64(4)},
			},
		},
		{
			name:   "conditional expression nil consequent",
			inType: semantic.BasicInt,
			input: map[string]interface{}{
				"r": map[string]interface{}{
					"a":    []interface{}{nil, int64(3)},
					"b":    []interface{}{int64(2), int64(4)},
					"cond": []interface{}{true, false},
				},
			},
			want: map[string]interface{}{
				// when a nil value is selected, it gets passed through
				"c": []interface{}{nil, int64(4)},
			},
		},
		{
			name:   "conditional expression nil alternate",
			inType: semantic.BasicInt,
			input: map[string]interface{}{
				"r": map[string]interface{}{
					"a":    []interface{}{int64(1), int64(3)},
					"b":    []interface{}{int64(2), nil},
					"cond": []interface{}{true, false},
				},
			},
			want: map[string]interface{}{
				// when a nil value is selected, it gets passed through
				"c": []interface{}{int64(1), nil},
			},
		},
	}

	for _, test := range conditionalTests {
		name := test.name
		if len(name) == 0 {
			name = fmt.Sprintf("conditional expression %s", test.inType.String())
		}

		testCases = append(testCases, TestCase{
			name:         name,
			fn:           `(r) => ({c: if r.cond then r.a else r.b})`,
			vectorizable: true,
			inType: semantic.NewObjectType([]semantic.PropertyType{
				{Key: []byte("r"), Value: semantic.NewObjectType([]semantic.PropertyType{
					{Key: []byte("a"), Value: semantic.NewVectorType(test.inType)},
					{Key: []byte("b"), Value: semantic.NewVectorType(test.inType)},
					{Key: []byte("cond"), Value: semantic.NewVectorType(semantic.BasicBool)},
				})},
			}),
			input:   test.input,
			want:    test.want,
			flagger: executetest.TestFlagger{},
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
			flagger[fluxfeature.VectorizedConst().Key()] = true
			flagger[fluxfeature.VectorizedConditionals().Key()] = true
			flagger[fluxfeature.VectorizedFloat().Key()] = true
			flagger[fluxfeature.VectorizedUnaryOps().Key()] = true
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
				// XXX: test the vectorized version instead of the row-based one.
				fn = fn.Vectorized
			} else {
				if fn.Vectorized != nil {
					t.Fatal("Vectorized node is populated when it should be nil")
				}
			}

			if tc.skipComp {
				return
			}

			f, err := compiler.Compile(ctx, nil, fn, tc.inType)
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
