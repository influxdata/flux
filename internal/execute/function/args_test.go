package function_test

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/internal/execute/function"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

func TestReadArgs(t *testing.T) {
	readArgs := func(spec interface{}, args map[string]values.Value) error {
		fargs := interpreter.NewArguments(values.NewObjectWithValues(args))
		return function.ReadArgs(spec, fargs, nil)
	}
	for _, tt := range []struct {
		name    string
		args    map[string]values.Value
		want    interface{}
		wantErr string
	}{
		{
			name: "Any",
			args: map[string]values.Value{
				"a": values.NewString("t0"),
				"b": values.NewInt(4),
			},
			want: &struct {
				A values.Value
				B values.Value
			}{
				A: values.NewString("t0"),
				B: values.NewInt(4),
			},
		},
		{
			name: "Float64",
			args: map[string]values.Value{
				"a": values.NewFloat(2.0),
			},
			want: &struct{ A float64 }{A: 2.0},
		},
		{
			name: "Int64",
			args: map[string]values.Value{
				"a": values.NewInt(5),
			},
			want: &struct{ A int64 }{A: 5},
		},
		{
			name: "Uint64",
			args: map[string]values.Value{
				"a": values.NewUInt(8),
			},
			want: &struct{ A uint64 }{A: 8},
		},
		{
			name: "Pointer",
			args: map[string]values.Value{
				"a": values.NewFloat(2.0),
			},
			want: &struct {
				A *float64
				B *int64
			}{
				A: func(v float64) *float64 { return &v }(2.0),
				B: nil,
			},
		},
		{
			name: "Strings",
			args: map[string]values.Value{
				"values": values.NewArrayWithBacking(
					semantic.NewArrayType(semantic.BasicString),
					[]values.Value{
						values.NewString("a"),
						values.NewString("b"),
						values.NewString("c"),
					},
				),
			},
			want: &struct {
				Values []string
			}{
				Values: []string{"a", "b", "c"},
			},
		},
		{
			name: "Object",
			args: map[string]values.Value{
				"columns": values.NewObjectWithValues(map[string]values.Value{
					"name":  values.NewString("foo"),
					"value": values.NewInt(4),
				}),
			},
			want: &struct {
				Columns map[string]values.Value
			}{
				Columns: map[string]values.Value{
					"name":  values.NewString("foo"),
					"value": values.NewInt(4),
				},
			},
		},
		{
			name: "ObjectNativeType",
			args: map[string]values.Value{
				"columns": values.NewObjectWithValues(map[string]values.Value{
					"name":  values.NewString("foo"),
					"value": values.NewString("bar"),
				}),
			},
			want: &struct {
				Columns map[string]string
			}{
				Columns: map[string]string{
					"name":  "foo",
					"value": "bar",
				},
			},
		},
		{
			name: "Struct",
			args: map[string]values.Value{
				"o": values.NewObjectWithValues(map[string]values.Value{
					"a": values.NewInt(4),
					"b": values.NewFloat(7.0),
				}),
			},
			want: &struct {
				O struct {
					A int64
					B float64
				}
			}{
				O: struct {
					A int64
					B float64
				}{A: 4, B: 7.0},
			},
		},
		{
			name: "TableObject",
			args: map[string]values.Value{
				"tables": &flux.TableObject{Kind: "from"},
				"column": values.NewString("_value"),
			},
			want: &struct {
				Tables *function.TableObject
				Column string
			}{
				Tables: &function.TableObject{
					TableObject: &flux.TableObject{
						Kind: "from",
					},
				},
				Column: "_value",
			},
		},
		{
			name: "TagName",
			args: map[string]values.Value{
				"a": values.NewInt(5),
			},
			want: &struct {
				B int `flux:"a"`
			}{B: 5},
		},
		{
			name: "Ignored",
			args: map[string]values.Value{
				"a": values.NewInt(5),
				"b": values.NewString("foo"),
			},
			want: &struct {
				A int
				B string `flux:"-"`
			}{
				A: 5,
			},
		},
		{
			name: "MissingRequired",
			args: map[string]values.Value{
				"a": values.NewInt(5),
			},
			want: &struct {
				A int
				B string `flux:",required"`
			}{
				A: 5,
			},
			wantErr: "missing required keyword argument \"b\"",
		},
		{
			name: "BadString",
			args: map[string]values.Value{
				"a": values.NewInt(4),
			},
			want:    &struct{ A string }{},
			wantErr: "keyword argument \"a\" should be of kind string, but got int",
		},
		{
			name: "BadInt",
			args: map[string]values.Value{
				"a": values.NewFloat(4),
			},
			want:    &struct{ A int }{},
			wantErr: "keyword argument \"a\" should be of kind int, but got float",
		},
		{
			name: "BadUInt",
			args: map[string]values.Value{
				"a": values.NewFloat(4),
			},
			want:    &struct{ A uint }{},
			wantErr: "keyword argument \"a\" should be of kind uint, but got float",
		},
		{
			name: "BadFloat",
			args: map[string]values.Value{
				"a": values.NewInt(4),
			},
			want:    &struct{ A float64 }{},
			wantErr: "keyword argument \"a\" should be of kind float, but got int",
		},
		{
			name: "BadArray",
			args: map[string]values.Value{
				"a": values.NewInt(4),
			},
			want:    &struct{ A []int }{},
			wantErr: "keyword argument \"a\" should be of kind array, but got int",
		},
		{
			name: "BadObject",
			args: map[string]values.Value{
				"a": values.NewInt(4),
			},
			want:    &struct{ A map[string]int }{},
			wantErr: "keyword argument \"a\" should be of kind object, but got int",
		},
		{
			name: "BadObjectKey",
			args: map[string]values.Value{
				"a": values.NewObjectWithValues(map[string]values.Value{
					"b": values.NewInt(4),
				}),
			},
			want:    &struct{ A map[int]int }{},
			wantErr: "only string keys are supported for map types",
		},
		{
			name: "BadStruct",
			args: map[string]values.Value{
				"o": values.NewInt(4),
			},
			want: &struct {
				O struct {
					A int64
					B float64
				}
			}{},
			wantErr: "keyword argument \"o\" should be of kind object, but got int",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			typ := reflect.TypeOf(tt.want).Elem()
			got := reflect.New(typ).Interface()
			if err := readArgs(got, tt.args); err != nil {
				if want, got := tt.wantErr, err.Error(); want != got {
					t.Fatalf("unexpected error -want/+got:\n\t- %q\n\t+ %q", want, got)
				}
				return
			}

			if tt.wantErr != "" {
				t.Fatal("expected error")
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
