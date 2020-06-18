package compiler_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/compiler"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/semantic/semantictest"
	"github.com/influxdata/flux/values"
)

var CmpOptions = semantictest.CmpOptions

func TestCompileAndEval(t *testing.T) {
	testCases := []struct {
		name    string
		fn      string
		inType  semantic.MonoType
		input   values.Object
		want    values.Value
		wantErr bool
	}{
		{
			name: "interpolated string expression",
			fn:   `(r) => "n = ${r.n}"`,
			inType: semantic.NewObjectType([]semantic.PropertyType{
				{Key: []byte("r"), Value: semantic.NewObjectType([]semantic.PropertyType{
					{Key: []byte("n"), Value: semantic.BasicString},
				})},
			}),
			input: values.NewObjectWithValues(map[string]values.Value{
				"r": values.NewObjectWithValues(map[string]values.Value{
					"n": values.NewString("2"),
				}),
			}),
			want: values.NewString("n = 2"),
		},
		{
			name: "interpolated string expression error",
			fn:   `(r) => "n = ${r.n}"`,
			inType: semantic.NewObjectType([]semantic.PropertyType{
				{Key: []byte("r"), Value: semantic.NewObjectType([]semantic.PropertyType{
					{Key: []byte("n"), Value: semantic.BasicInt},
				})},
			}),
			input: values.NewObjectWithValues(map[string]values.Value{
				"r": values.NewObjectWithValues(map[string]values.Value{
					"n": values.NewInt(10),
				}),
			}),
			wantErr: true,
		},
		{
			name: "simple ident return",
			fn:   `(r) => r`,
			inType: semantic.NewObjectType([]semantic.PropertyType{
				{Key: []byte("r"), Value: semantic.BasicInt},
			}),
			input: values.NewObjectWithValues(map[string]values.Value{
				"r": values.NewInt(4),
			}),
			want:    values.NewInt(4),
			wantErr: false,
		},
		{
			name: "call function",
			fn:   `(r) => ((a,b) => a + b)(a:1, b:r)`,
			inType: semantic.NewObjectType([]semantic.PropertyType{
				{Key: []byte("r"), Value: semantic.BasicInt},
			}),
			input: values.NewObjectWithValues(map[string]values.Value{
				"r": values.NewInt(4),
			}),
			want:    values.NewInt(5),
			wantErr: false,
		},
		{
			name: "call function with defaults",
			fn:   `(r) => ((a=0,b) => a + b)(b:r)`,
			inType: semantic.NewObjectType([]semantic.PropertyType{
				{Key: []byte("r"), Value: semantic.BasicInt},
			}),
			input: values.NewObjectWithValues(map[string]values.Value{
				"r": values.NewInt(4),
			}),
			want:    values.NewInt(4),
			wantErr: false,
		},
		{
			name: "call function via identifier",
			fn: `(r) => {
				f = (a,b) => a + b
				return f(a:1, b:r)
			}`,
			inType: semantic.NewObjectType([]semantic.PropertyType{
				{Key: []byte("r"), Value: semantic.BasicInt},
			}),
			input: values.NewObjectWithValues(map[string]values.Value{
				"r": values.NewInt(4),
			}),
			want:    values.NewInt(5),
			wantErr: false,
		},
		{
			name: "call function via identifier with different types",
			fn: `(r) => {
				i = (x) => x
				return i(x:i)(x:r+1)
			}`,
			inType: semantic.NewObjectType([]semantic.PropertyType{
				{Key: []byte("r"), Value: semantic.BasicInt},
			}),
			input: values.NewObjectWithValues(map[string]values.Value{
				"r": values.NewInt(4),
			}),
			want:    values.NewInt(5),
			wantErr: false,
		},
		{
			name: "call filter function with index expression",
			fn:   `(r) => r[2] == 3`,
			inType: semantic.NewObjectType([]semantic.PropertyType{
				{Key: []byte("r"), Value: semantic.NewArrayType(semantic.BasicInt)},
			}),
			input: values.NewObjectWithValues(map[string]values.Value{
				"r": values.NewArrayWithBacking(semantic.NewArrayType(semantic.BasicInt), []values.Value{
					values.NewInt(5),
					values.NewInt(6),
					values.NewInt(3),
				}),
			}),
			want:    values.NewBool(true),
			wantErr: false,
		},
		{
			name: "call filter function with complex index expression",
			fn:   `(r) => r[((x) => 2)(x: "anything")] == 3`,
			inType: semantic.NewObjectType([]semantic.PropertyType{
				{Key: []byte("r"), Value: semantic.NewArrayType(semantic.BasicInt)},
			}),
			input: values.NewObjectWithValues(map[string]values.Value{
				"r": values.NewArrayWithBacking(semantic.NewArrayType(semantic.BasicInt), []values.Value{
					values.NewInt(5),
					values.NewInt(6),
					values.NewInt(3),
				}),
			}),
			want:    values.NewBool(true),
			wantErr: false,
		},
		{
			name: "conditional",
			fn:   `(t, c, a) => if t then c else a`,
			inType: semantic.NewObjectType([]semantic.PropertyType{
				{Key: []byte("t"), Value: semantic.BasicBool},
				{Key: []byte("c"), Value: semantic.BasicString},
				{Key: []byte("a"), Value: semantic.BasicString},
			}),
			input: values.NewObjectWithValues(map[string]values.Value{
				"t": values.NewBool(true),
				"c": values.NewString("cats"),
				"a": values.NewString("dogs"),
			}),
			want: values.NewString("cats"),
		},
		{
			name: "unary logical operator - not",
			fn:   `(a, b) => not a or b`,
			inType: semantic.NewObjectType([]semantic.PropertyType{
				{Key: []byte("a"), Value: semantic.BasicBool},
				{Key: []byte("b"), Value: semantic.BasicBool},
			}),
			input: values.NewObjectWithValues(map[string]values.Value{
				"a": values.NewBool(true),
				"b": values.NewBool(true),
			}),
			want: values.NewBool(true),
		},
		{
			name: "unary logical operator - exists with null",
			fn:   `(a) => exists a`,
			inType: semantic.NewObjectType([]semantic.PropertyType{
				{Key: []byte("a"), Value: semantic.BasicString},
			}),
			input: values.NewObjectWithValues(map[string]values.Value{
				"a": values.NewNull(semantic.BasicString),
			}),
			want: values.NewBool(false),
		},
		{
			name: "unary logical operator - exists without null",
			fn:   `(a, b) => not a and exists b`,
			inType: semantic.NewObjectType([]semantic.PropertyType{
				{Key: []byte("a"), Value: semantic.BasicBool},
				{Key: []byte("b"), Value: semantic.BasicString},
			}),
			input: values.NewObjectWithValues(map[string]values.Value{
				"a": values.NewBool(true),
				"b": values.NewString("I exist"),
			}),
			want: values.NewBool(false),
		},
		{
			name: "unary operator",
			fn:   `(a) => if a < 0 then -a else +a`,
			inType: semantic.NewObjectType([]semantic.PropertyType{
				{Key: []byte("a"), Value: semantic.BasicInt},
			}),
			input: values.NewObjectWithValues(map[string]values.Value{
				"a": values.NewInt(5),
			}),
			want: values.NewInt(5),
		},
		{
			name: "filter with member expression",
			fn:   `(r) => r.m == "cpu"`,
			inType: semantic.NewObjectType([]semantic.PropertyType{
				{Key: []byte("r"), Value: semantic.NewObjectType([]semantic.PropertyType{
					{Key: []byte("m"), Value: semantic.BasicString},
				})},
			}),
			input: values.NewObjectWithValues(map[string]values.Value{
				"r": values.NewObjectWithValues(map[string]values.Value{
					"m": values.NewString("cpu"),
				}),
			}),
			want:    values.NewBool(true),
			wantErr: false,
		},
		{
			name: "regex literal filter",
			fn:   `(r) => r =~ /^(c|g)pu$/`,
			inType: semantic.NewObjectType([]semantic.PropertyType{
				{Key: []byte("r"), Value: semantic.BasicString},
			}),
			input: values.NewObjectWithValues(map[string]values.Value{
				"r": values.NewString("cpu"),
			}),
			want:    values.NewBool(true),
			wantErr: false,
		},
		{
			name: "block statement with conditional",
			fn: `(r) => {
				v = if r < 0 then -r else r
				return v * v
			}`,
			inType: semantic.NewObjectType([]semantic.PropertyType{
				{Key: []byte("r"), Value: semantic.BasicInt},
			}),
			input: values.NewObjectWithValues(map[string]values.Value{
				"r": values.NewInt(-3),
			}),
			want:    values.NewInt(9),
			wantErr: false,
		},
		{
			name:   "array literal",
			fn:     `() => [1.0, 2.0, 3.0]`,
			inType: semantic.NewObjectType(nil),
			input:  values.NewObjectWithValues(nil),
			want: values.NewArrayWithBacking(
				semantic.NewArrayType(semantic.BasicFloat),
				[]values.Value{
					values.NewFloat(1),
					values.NewFloat(2),
					values.NewFloat(3),
				},
			),
			wantErr: false,
		},
		{
			name: "logical expression",
			fn:   `(a, b) => a or b`,
			inType: semantic.NewObjectType([]semantic.PropertyType{
				{Key: []byte("a"), Value: semantic.BasicBool},
				{Key: []byte("b"), Value: semantic.BasicBool},
			}),
			input: values.NewObjectWithValues(map[string]values.Value{
				"a": values.NewBool(true),
				"b": values.NewBool(false),
			}),
			want:    values.NewBool(true),
			wantErr: false,
		},
		{
			name: "call with nonexistant value",
			fn:   `(r) => r.a + r.b`,
			inType: semantic.NewObjectType([]semantic.PropertyType{
				{Key: []byte("r"), Value: semantic.NewObjectType([]semantic.PropertyType{
					{Key: []byte("a"), Value: semantic.BasicInt},
				})},
			}),
			input: values.NewObjectWithValues(map[string]values.Value{
				"r": values.NewObjectWithValues(map[string]values.Value{
					"a": values.NewInt(4),
				}),
			}),
			want: values.Null,
		},
		{
			name: "call with null value",
			fn:   `(r) => r.a + r.b`,
			inType: semantic.NewObjectType([]semantic.PropertyType{
				{Key: []byte("r"), Value: semantic.NewObjectType([]semantic.PropertyType{
					{Key: []byte("a"), Value: semantic.BasicInt},
					{Key: []byte("b"), Value: semantic.BasicInt},
				})},
			}),
			input: values.NewObjectWithValues(map[string]values.Value{
				"r": values.NewObjectWithValues(map[string]values.Value{
					"a": values.NewInt(4),
					// The object is typed as an integer,
					// but it doesn't have an actual value
					// because it is null.
					"b": values.Null,
				}),
			}),
			want: values.Null,
		},
		{
			name: "call with null parameter",
			fn: `(r) => {
				eval = (a, b) => a + b
				return eval(a: r.a, b: r.b)
			}`,
			inType: semantic.NewObjectType([]semantic.PropertyType{
				{Key: []byte("r"), Value: semantic.NewObjectType([]semantic.PropertyType{
					{Key: []byte("a"), Value: semantic.BasicInt},
					// "b": semantic.Nil,
				})},
			}),
			input: values.NewObjectWithValues(map[string]values.Value{
				"r": values.NewObjectWithValues(map[string]values.Value{
					"a": values.NewInt(4),
					// "b": values.Null,
				}),
			}),
			want: values.Null,
		},
		{
			name: "return nonexistant value",
			fn:   `(r) => r.b`,
			inType: semantic.NewObjectType([]semantic.PropertyType{
				{Key: []byte("r"), Value: semantic.NewObjectType([]semantic.PropertyType{
					{Key: []byte("a"), Value: semantic.BasicInt},
				})},
			}),
			input: values.NewObjectWithValues(map[string]values.Value{
				"r": values.NewObjectWithValues(map[string]values.Value{
					"a": values.NewInt(4),
				}),
			}),
			want: values.Null,
		},
		{
			name: "return nonexistant and used parameter",
			fn: `(r) => {
				b = (r) => r.b
				return r.a + b(r: r)
			}`,
			inType: semantic.NewObjectType([]semantic.PropertyType{
				{Key: []byte("r"), Value: semantic.NewObjectType([]semantic.PropertyType{
					{Key: []byte("a"), Value: semantic.BasicInt},
					// "b": semantic.Nil,
				})},
			}),
			input: values.NewObjectWithValues(map[string]values.Value{
				"r": values.NewObjectWithValues(map[string]values.Value{
					"a": values.NewInt(4),
					// "b": values.Null,
				}),
			}),
			want: values.Null,
		},
		{
			name: "two null values are not equal",
			fn:   `(a, b) => a == b`,
			inType: semantic.NewObjectType([]semantic.PropertyType{
				{Key: []byte("a"), Value: semantic.BasicInt},
				{Key: []byte("b"), Value: semantic.BasicInt},
			}),
			input: values.NewObjectWithValues(map[string]values.Value{
				"a": values.Null,
				"b": values.Null,
			}),
			want: values.Null,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			pkg, err := runtime.AnalyzeSource(tc.fn)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			stmt := pkg.Files[0].Body[0].(*semantic.ExpressionStatement)
			fn := stmt.Expression.(*semantic.FunctionExpression)
			f, err := compiler.Compile(nil, fn, tc.inType)
			if err != nil {
				if !tc.wantErr {
					t.Fatalf("unexpected error: %s", err)
				}
				return
			} else if tc.wantErr {
				t.Fatal("wanted error but got nothing")
			}

			// ctx := dependenciestest.Default().Inject(context.Background())
			got, err := f.Eval(context.TODO(), tc.input)
			if tc.wantErr != (err != nil) {
				t.Errorf("unexpected error: %s", err)
			}

			if !cmp.Equal(tc.want, got, CmpOptions...) {
				t.Errorf("unexpected value -want/+got\n%s", cmp.Diff(tc.want, got, CmpOptions...))
			}
		})
	}
}

func TestCompiler_ReturnType(t *testing.T) {
	testCases := []struct {
		name   string
		fn     string
		inType semantic.MonoType
		want   string
	}{
		{
			name: "with",
			fn:   `(r) => ({r with _value: r._value * 2.0})`,
			inType: semantic.NewObjectType([]semantic.PropertyType{
				{Key: []byte("r"), Value: semantic.NewObjectType([]semantic.PropertyType{
					{Key: []byte("_value"), Value: semantic.BasicFloat},
					{Key: []byte("_time"), Value: semantic.BasicTime},
				})},
			}),
			want: `{_time: time | _value: float | _value: float}`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			pkg, err := runtime.AnalyzeSource(tc.fn)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			stmt := pkg.Files[0].Body[0].(*semantic.ExpressionStatement)
			fn := stmt.Expression.(*semantic.FunctionExpression)
			f, err := compiler.Compile(nil, fn, tc.inType)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if got, want := f.Type().String(), tc.want; got != want {
				t.Fatalf("unexpected return type -want/+got:\n\t- %s\n\t+ %s", want, got)
			}
		})
	}
}

func TestToScopeNil(t *testing.T) {
	if compiler.ToScope(nil) != nil {
		t.Fatal("ToScope made non-nil scope from a nil base")
	}
}
