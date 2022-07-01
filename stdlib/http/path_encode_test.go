package http_test

import (
	"testing"

	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/stdlib/http"
	"github.com/influxdata/flux/values"
)

func TestPathEscape(t *testing.T) {
	inputString := "random:/#"
	want := values.NewString("random:%2F%23")

	args := interpreter.NewArguments(values.NewObjectWithValues(
		map[string]values.Value{
			"inputString": values.NewString(inputString),
		}),
	)

	got, err := http.PathEncode(args)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if !want.Equal(got) {
		t.Fatalf("unexpected value -want/+got:\n\t- %#v\n\t+ %#v", want, got)
	}
}
