package flux_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/parser"
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

	verify(flux.Eval(context.Background(), src))
	verify(flux.EvalAST(context.Background(), astSrc))
}

func TestEval_error(t *testing.T) {
	// parse error
	src := `x = ()`
	_, _, err := flux.Eval(context.Background(), src)
	if err == nil {
		t.Fatal("expected error, got none")
	}
	if want, got := "error at @1:5-1:7: expected ARROW, got EOF", err.Error(); want != got {
		t.Errorf("wanted error %q, got %q", want, got)
	}

	// analysis error
	src = `x = 1.0 + "foo"`
	_, _, err = flux.Eval(context.Background(), src)
	if err == nil {
		t.Fatal("expected error, got none")
	}
	if want, got := "cannot unify float with string", err.Error(); want != got {
		t.Errorf("wanted error %q, got %q", want, got)
	}

}
