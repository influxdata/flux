package runtime_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	_ "github.com/influxdata/flux/builtin"
	"github.com/influxdata/flux/dependencies/dependenciestest"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/parser"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/values"
)

func TestEval(t *testing.T) {
	src := `
		f = ((x) => x + 1)
		y = f(x: 41)`
	astSrc := parser.ParseSource(src)

	verify := func(sideEffects []interpreter.SideEffect, scope values.Scope, err error) {
		if err != nil {
			t.Fatal(err)
		}
		want := map[string]string{
			"f": "(x: int) -> int",
			"y": "42",
		}
		scope.LocalRange(func(k string, v values.Value) {
			wantV, ok := want[k]
			if !ok {
				t.Errorf("did not find %q in scope", k)
			}
			if gotV := fmt.Sprintf("%v", v); gotV != wantV {
				t.Errorf("wanted %q, got %q", wantV, gotV)
			}
		})
		if len(sideEffects) > 0 {
			t.Errorf("expected empty side effects, got %v", sideEffects)
		}

	}

	verify(runtime.Eval(context.Background(), src))
	verify(runtime.EvalAST(context.Background(), astSrc))
}

func TestEval_error(t *testing.T) {
	// parse error
	src := `x = ()`
	_, _, err := runtime.Eval(context.Background(), src)
	if err == nil {
		t.Fatal("expected error, got none")
	}
	if want, got := "error at @1:5-1:7: expected ARROW, got EOF", err.Error(); want != got {
		t.Errorf("wanted error %q, got %q", want, got)
	}

	// analysis error
	src = `x = 1.0 + "foo"`
	_, _, err = runtime.Eval(context.Background(), src)
	if err == nil {
		t.Fatal("expected error, got none")
	}
	if want, got := "type error @1:11-1:16: expected float but found string", err.Error(); want != got {
		t.Errorf("wanted error %q, got %q", want, got)
	}

}

// Example_option demonstrates retrieving an option value from a scope object
func Example_option() {

	// Import the universe package.
	importer := runtime.StdLib()
	universe, _ := importer.ImportPackageObject("universe")

	// Retrieve the default value for the now option
	nowFunc, _ := universe.Get("now")

	// The now option is a function value whose default behavior is to return
	// the current system time when called. The function now() doesn't take
	// any arguments so can be called with nil.
	nowTime, _ := nowFunc.Function().Call(dependenciestest.Default().Inject(context.TODO()), nil)
	fmt.Fprintf(os.Stderr, "The current system time (UTC) is: %v\n", nowTime)
	// Output:
}

// Example_overrideDefaultOptionExternally demonstrates how declaring an option
// in a Flux script will change that option's binding globally.
func Example_overrideDefaultOptionExternally() {
	queryString := `
		option now = () => 2018-07-13T00:00:00Z
		what_time_is_it = now()`

	ctx := dependenciestest.Default().Inject(context.Background())
	_, scope, err := runtime.Eval(ctx, queryString)
	if err != nil {
		fmt.Println(err)
	}

	// After evaluating the package, lookup the value of what_time_is_it
	now, _ := scope.Lookup("what_time_is_it")

	// what_time_is_it? Why it's ....
	fmt.Printf("The new current time (UTC) is: %v", now)
	// Output: The new current time (UTC) is: 2018-07-13T00:00:00.000000000Z
}
