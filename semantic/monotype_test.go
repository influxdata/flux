package semantic_test

import (
	"strings"
	"testing"

	"github.com/InfluxCommunity/flux/semantic"
)

func TestBasicType(t *testing.T) {
	for _, tt := range []struct {
		typ  semantic.MonoType
		want string
	}{
		{typ: semantic.BasicBool, want: "bool"},
		{typ: semantic.BasicInt, want: "int"},
		{typ: semantic.BasicUint, want: "uint"},
		{typ: semantic.BasicFloat, want: "float"},
		{typ: semantic.BasicString, want: "string"},
		{typ: semantic.BasicDuration, want: "duration"},
		{typ: semantic.BasicTime, want: "time"},
		{typ: semantic.BasicRegexp, want: "regexp"},
		{typ: semantic.BasicBytes, want: "bytes"},
	} {
		//lint:ignore SA1019 Test code that's not important to update
		t.Run(strings.Title(tt.want), func(t *testing.T) {
			if want, got := tt.typ.String(), tt.want; want != got {
				t.Errorf("unexpected monotype -want/+got:\n\t- %s\n\t+ %s", want, got)
			}
		})
	}
}

func TestNewArrayType(t *testing.T) {
	arrayType := semantic.NewArrayType(semantic.BasicInt)
	if want, got := arrayType.String(), "[int]"; want != got {
		t.Errorf("unexpected monotype -want/+got:\n\t- %s\n\t+ %s", want, got)
	}
}

func TestNewFunctionType(t *testing.T) {
	functionType := semantic.NewFunctionType(
		semantic.BasicString,
		[]semantic.ArgumentType{
			{Name: []byte("v"), Type: semantic.BasicInt},
		},
	)
	if want, got := functionType.String(), "(v: int) => string"; want != got {
		t.Errorf("unexpected monotype -want/+got:\n\t- %s\n\t+ %s", want, got)
	}
}

func TestNewObjectType(t *testing.T) {
	objectType := semantic.NewObjectType(
		[]semantic.PropertyType{
			{Key: []byte("a"), Value: semantic.BasicInt},
			{Key: []byte("b"), Value: semantic.BasicString},
		},
	)
	if want, got := objectType.String(), "{a: int, b: string}"; want != got {
		t.Errorf("unexpected monotype -want/+got:\n\t- %s\n\t+ %s", want, got)
	}

	// Nest the object type in another object.
	objectType = semantic.NewObjectType(
		[]semantic.PropertyType{
			{Key: []byte("r"), Value: objectType},
		},
	)
	if want, got := objectType.String(), "{r: {a: int, b: string}}"; want != got {
		t.Errorf("unexpected monotype -want/+got:\n\t- %s\n\t+ %s", want, got)
	}
}
