package http_test

import (
	"testing"

	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/stdlib/http"
	"github.com/influxdata/flux/values"
)

func TestBasicAuth(t *testing.T) {
	u, p := "me", "mypassword"
	want := values.NewString("Basic bWU6bXlwYXNzd29yZA==")

	args := interpreter.NewArguments(values.NewObjectWithValues(
		map[string]values.Value{
			"u": values.NewString(u),
			"p": values.NewString(p),
		}),
	)
	got, err := http.BasicAuth(args)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if !want.Equal(got) {
		t.Fatalf("unexpected value -want/+got:\n\t- %#v\n\t+ %#v", want, got)
	}
}
