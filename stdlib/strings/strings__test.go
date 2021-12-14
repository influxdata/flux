package strings_test

// n.b. strings_test.go belongs to the "strings" package so it can access
// private functions from the implementation.
// Importing "flux/fluxinit/static" in there causes a cyclical import.
// This is why we have a strings__test.go in addition to strings_test.go.

import (
	"context"

	_ "github.com/influxdata/flux/fluxinit/static"
	"github.com/influxdata/flux/runtime"
	"testing"
)

func TestJoinStr_ReceiveTableObjectIsError(t *testing.T) {
	src := `
	import "array"
	import "strings"
	strings.joinStr(arr: array.from(rows: [{_value: ""}]) |> map(fn: (r) => r._value), v: ",")`
	_, _, err := runtime.Eval(context.Background(), src)
	if err == nil {
		t.Fatal("expected error, got none")
	}

	if want, got := "error calling function \"joinStr\" @4:2-4:92: \"arr\" cannot be a table stream; expected an array", err.Error(); want != got {
		t.Errorf("wanted error %q, got %q", want, got)
	}
}
