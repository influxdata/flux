package executetest

import (
	"context"
	"testing"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

var (
	prelude values.Scope
	stdlib  interpreter.Importer
)

// FunctionExpression will take a function expression as a string
// and return the *semantic.FunctionExpression.
//
// This will cause a fatal error in the test on failure.
func FunctionExpression(t testing.TB, source string) *semantic.FunctionExpression {
	t.Helper()

	if prelude == nil {
		prelude = flux.Prelude()
	}
	if stdlib == nil {
		stdlib = flux.StdLib()
	}

	pkg, err := semantic.AnalyzeSource(source)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	// Interpret and resolve the function which will replace
	// variables with their values (notably identifiers "true"
	// and "false" will be replaced with boolean literals)
	itrp := interpreter.NewInterpreter(nil)
	se, err := itrp.Eval(context.Background(), pkg, prelude, stdlib)
	if err != nil {
		t.Fatal(err)
	}

	if len(se) != 1 {
		t.Fatal("expected just one side effect")
	}

	f := se[0].Value.(values.Function)
	rf, err := interpreter.ResolveFunction(f)
	if err != nil {
		t.Fatal(err)
	}

	return rf.Fn
}
