package json

import (
	"testing"

	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

func TestUnmarshalToValue(t *testing.T) {
	var testCases = []struct {
		name    string
		bytes   string
		want    values.Value
		wantErr error
	}{
		{
			name:  "all basic types",
			bytes: `{"a":"x","b":1,"c":true,"d":[1,2,3]}`,
			want: values.NewObjectWithValues(map[string]values.Value{
				"a": values.NewString("x"),
				"b": values.NewFloat(1),
				"c": values.NewBool(true),
				"d": values.NewArrayWithBacking(semantic.NewArrayType(semantic.BasicFloat), []values.Value{
					values.NewFloat(1),
					values.NewFloat(2),
					values.NewFloat(3),
				}),
			}),
		},
		{
			name:    "mixed array types",
			bytes:   `{"a":[1,false,3]}`,
			wantErr: errors.New(codes.Invalid, "array values must all be the same type"),
		},
		{
			name:  "bare float",
			bytes: `1`,
			want:  values.NewFloat(1),
		},
		{
			name:  "bare string",
			bytes: `"a"`,
			want:  values.NewString("a"),
		},
		{
			name:  "bare bool",
			bytes: `false`,
			want:  values.NewBool(false),
		},
		{
			name:  "bare array",
			bytes: `[1,2,3]`,
			want: values.NewArrayWithBacking(semantic.NewArrayType(semantic.BasicFloat), []values.Value{
				values.NewFloat(1),
				values.NewFloat(2),
				values.NewFloat(3),
			}),
		},
		{
			name:  "nested objects",
			bytes: `{"a":{"x":1}}`,
			want: values.NewObjectWithValues(map[string]values.Value{
				"a": values.NewObjectWithValues(map[string]values.Value{
					"x": values.NewFloat(1),
				}),
			}),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got, err := unmarshalToValue([]byte(tc.bytes))
			if tc.wantErr != nil && err == nil {
				t.Fatalf("expected error: %s", tc.wantErr)
			}
			if tc.wantErr != nil && err != nil {
				if tc.wantErr.Error() != err.Error() {
					t.Fatalf("expected errors: want %s got: %s", tc.wantErr, err)
				} else {
					return
				}
			}
			if tc.wantErr == nil && err != nil {
				t.Fatal(err)
			}
			if !tc.want.Equal(got) {
				t.Errorf("unequal values \nwant: %v\ngot:  %v", tc.want, got)
			}
		})
	}
}

func TestUnmarshalToValueNull(t *testing.T) {
	bytes := `{"a":null}`
	got, err := unmarshalToValue([]byte(bytes))
	if err != nil {
		t.Fatal(err)
	}
	if got.Type().Nature() != semantic.Object {
		t.Fatal("data should be an object")
	}

	a, ok := got.Object().Get("a")
	if !ok {
		t.Fatal("object property 'a' should exist")
	}
	if !a.IsNull() {
		t.Error("property 'a' should be null")
	}
}
