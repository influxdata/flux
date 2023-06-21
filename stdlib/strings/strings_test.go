package strings_test

// n.b. strings_test.go belongs to the "strings" package so it can access
// private functions from the implementation.
// Importing "flux/fluxinit/static" in there causes a cyclical import.
// This is why we have a strings__test.go in addition to strings_test.go.

import (
	"context"
	"testing"

	"github.com/InfluxCommunity/flux/dependency"
	fluxstdlibstrings "github.com/InfluxCommunity/flux/stdlib/strings"

	"github.com/InfluxCommunity/flux/dependencies/dependenciestest"
	_ "github.com/InfluxCommunity/flux/fluxinit/static"
	"github.com/InfluxCommunity/flux/runtime"
	"github.com/InfluxCommunity/flux/semantic"
	"github.com/InfluxCommunity/flux/values"
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

func TestJoinStr_NullInArrParam(t *testing.T) {
	fluxFunc := fluxstdlibstrings.SpecialFns["joinStr"]
	arr := values.NewArrayWithBacking(semantic.NewArrayType(semantic.BasicString), []values.Value{
		values.NewString("a"), values.NewString("b"), values.NewNull(semantic.BasicString)})
	fluxArg := values.NewObjectWithValues(map[string]values.Value{"arr": arr, "v": values.NewString(", ")})
	wantErr := "expected elements of argument \"arr\" to be of type string, got type string value <nil>"
	ctx, deps := dependency.Inject(context.Background(), dependenciestest.Default())
	defer deps.Finish()
	gotErr, err := fluxFunc.Call(ctx, fluxArg)
	if err != nil {
		if gotErr, wantErr := err.Error(), wantErr; gotErr != wantErr {
			t.Errorf("unexpected error - wantErr: %s, gotErr: %s", wantErr, gotErr)
		}
		return
	}
	if wantErr != gotErr.Str() {
		t.Errorf("input %f: expected %v, gotErr %f", arr, wantErr, gotErr)
	}
}
