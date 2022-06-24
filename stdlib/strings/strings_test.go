package strings_test

// n.b. strings_test.go belongs to the "strings" package so it can access
// private functions from the implementation.
// Importing "flux/fluxinit/static" in there causes a cyclical import.
// This is why we have a strings__test.go in addition to strings_test.go.

import (
	"context"

	_ "github.com/mvn-trinhnguyen2-dn/flux/fluxinit/static"
	"github.com/mvn-trinhnguyen2-dn/flux/runtime"
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

	if want, got := "error @4:59-4:83: expected [string] (array) but found stream[string] (argument arr)", err.Error(); want != got {
		t.Errorf("wanted error %q, got %q", want, got)
	}
}
