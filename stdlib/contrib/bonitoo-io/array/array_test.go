package array_test

import (
	"context"
	"testing"

	"github.com/influxdata/flux/dependencies/dependenciestest"
	_ "github.com/influxdata/flux/fluxinit/static"
	"github.com/influxdata/flux/querytest"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/stdlib/contrib/bonitoo-io/array"
	"github.com/influxdata/flux/values"
)

func TestConcat_NewQuery(t *testing.T) {
	tests := []querytest.NewQueryTestCase{
		{
			Name:    "no args",
			Raw:     `import "contrib/bonitoo-io/array" array.concat()`,
			WantErr: true, // missing required args
		},
		{
			Name:    "invalid arr arg",
			Raw:     `import "contrib/bonitoo-io/array" array.concat(arr: 1, v: [2])`,
			WantErr: true, // expected [A] but found int (argument arr)
		},
		{
			Name:    "invalid v arg",
			Raw:     `import "contrib/bonitoo-io/array" array.concat(arr: [1], v: 2)`,
			WantErr: true, // expected [int] but found int (argument v)
		},
		{
			Name:    "type mismatch",
			Raw:     `import "contrib/bonitoo-io/array" array.concat(arr: [1], v: [2.0])`,
			WantErr: true, // expected int but found float (argument v)
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

func TestConcat_Process(t *testing.T) {
	type record struct {
		s string
		i int64
	}
	tovalarr := func(typ semantic.MonoType, arr []interface{}) []values.Value {
		vals := make([]values.Value, len(arr))
		for i, e := range arr {
			switch v := e.(type) {
			case int64, float64, string:
				vals[i] = values.New(v)
			case record:
				obj := values.NewObject(typ)
				obj.Set("s", values.New(v.s))
				obj.Set("i", values.New(v.i))
				vals[i] = obj
			default:
				t.Errorf("unspported type %T", e)
			}
		}
		return vals
	}
	testCases := []struct {
		name string
		typ  semantic.MonoType
		arr  []interface{}
		v    []interface{}
		want []interface{}
	}{
		{
			name: "int",
			typ:  semantic.BasicInt,
			arr:  []interface{}{int64(1), int64(2), int64(3)},
			v:    []interface{}{int64(4), int64(5)},
			want: []interface{}{int64(1), int64(2), int64(3), int64(4), int64(5)},
		},
		{
			name: "float",
			typ:  semantic.BasicFloat,
			arr:  []interface{}{1.1, 2.2, 3.3},
			v:    []interface{}{4.4, 5.5},
			want: []interface{}{1.1, 2.2, 3.3, 4.4, 5.5},
		},
		{
			name: "string",
			typ:  semantic.BasicString,
			arr:  []interface{}{"a", "b", "c"},
			v:    []interface{}{"d", "e"},
			want: []interface{}{"a", "b", "c", "d", "e"},
		},
		{
			name: "record",
			typ: semantic.NewObjectType([]semantic.PropertyType{
				{Key: []byte("s"), Value: semantic.BasicString},
				{Key: []byte("i"), Value: semantic.BasicInt},
			}),
			arr:  []interface{}{record{s: "a", i: 1}, record{s: "b", i: 2}, record{s: "c", i: 3}},
			v:    []interface{}{record{s: "d", i: 4}, record{s: "e", i: 5}},
			want: []interface{}{record{s: "a", i: 1}, record{s: "b", i: 2}, record{s: "c", i: 3}, record{s: "d", i: 4}, record{s: "e", i: 5}},
		},
	}

	concatFn := array.SpecialFns["concat"]

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			fluxArg := values.NewObjectWithValues(map[string]values.Value{
				"arr": values.NewArrayWithBacking(semantic.NewArrayType(tc.typ), tovalarr(tc.typ, tc.arr)),
				"v":   values.NewArrayWithBacking(semantic.NewArrayType(tc.typ), tovalarr(tc.typ, tc.v)),
			})
			want := values.NewArrayWithBacking(semantic.NewArrayType(tc.typ), tovalarr(tc.typ, tc.want))
			result, err := concatFn.Call(dependenciestest.Default().Inject(context.Background()), fluxArg)
			if err != nil {
				t.Error(err.Error())
			}
			got := result.Array()
			if !got.Equal(want) {
				t.Errorf("[%s] expected %v (%T), got %v (%T)", tc.name, want, want, got, got)
			}
		})
	}
}

func TestMap_NewQuery(t *testing.T) {
	tests := []querytest.NewQueryTestCase{
		{
			Name:    "no args",
			Raw:     `import "contrib/bonitoo-io/array" array.map()`,
			WantErr: true, // missing required args
		},
		{
			Name:    "invalid arr arg",
			Raw:     `import "contrib/bonitoo-io/array" array.map(arr: 1, fn: (x) => x)`,
			WantErr: true, // expected [A] but found int (argument arr)
		},
		{
			Name:    "invalid fn arg",
			Raw:     `import "contrib/bonitoo-io/array" array.map(arr: [1], fn: (x) => {x})`,
			WantErr: true, // missing return statement in block
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

func TestMap_Process(t *testing.T) {
	type record struct {
		f float64
	}
	tovalarr := func(typ semantic.MonoType, arr []interface{}) []values.Value {
		vals := make([]values.Value, len(arr))
		for i, e := range arr {
			switch v := e.(type) {
			case int64, float64, string:
				vals[i] = values.New(v)
			case record:
				obj := values.NewObject(typ)
				obj.Set("f", values.New(v.f))
				vals[i] = obj
			default:
				t.Errorf("unspported type %T", e)
			}
		}
		return vals
	}
	testCases := []struct {
		name string
		atyp semantic.MonoType
		arr  []interface{}
		fn   string
		wtyp semantic.MonoType
		want []interface{}
	}{
		{
			name: "int to string",
			atyp: semantic.BasicInt,
			arr:  []interface{}{int64(1), int64(2), int64(3)},
			fn:   `fx = (x) => string(v: x)`,
			wtyp: semantic.BasicString,
			want: []interface{}{"1", "2", "3"},
		},
		{
			name: "float to record",
			atyp: semantic.BasicFloat,
			arr:  []interface{}{1.1, 2.2, 3.3},
			fn:   `fx = (x) => ({f: x})`,
			wtyp: semantic.NewObjectType([]semantic.PropertyType{
				{Key: []byte("f"), Value: semantic.BasicFloat},
			}),
			want: []interface{}{record{f: 1.1}, record{f: 2.2}, record{f: 3.3}},
		},
	}

	mapFn := array.SpecialFns["map"]

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctx := dependenciestest.Default().Inject(context.Background())
			_, scope, err := runtime.Eval(ctx, tc.fn)
			if err != nil {
				t.Error(err.Error())
			}
			fx, ok := scope.Lookup("fx")
			if !ok {
				t.Error("must define a function to the fx variable")
			}
			fluxArg := values.NewObjectWithValues(map[string]values.Value{
				"arr": values.NewArrayWithBacking(semantic.NewArrayType(tc.atyp), tovalarr(tc.atyp, tc.arr)),
				"fn":  fx,
			})
			want := values.NewArrayWithBacking(semantic.NewArrayType(tc.wtyp), tovalarr(tc.wtyp, tc.want))
			result, err := mapFn.Call(ctx, fluxArg)
			if err != nil {
				t.Error(err.Error())
			}
			got := result.Array()
			if !got.Equal(want) {
				t.Errorf("[%s] expected %v (%T), got %v (%T)", tc.name, want, want, got, got)
			}
		})
	}
}
