package compiler

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/libflux/go/libflux"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/semantic/semantictest"
	"github.com/influxdata/flux/values"
)

type mockImporter struct {
}

func (m mockImporter) ImportPackageObject(_ string) (*interpreter.Package, error) {
	panic("unimplemented")
}

// TestFunctionValue_Resolve just demonstrates that
// functionValue implements the values.Resolver interface
func TestFunctionValue_Resolve(t *testing.T) {
	src1 := `x = 42 y = 100`
	src2 := `() => x + y`

	// we want to show that a functionValue like the above
	// will be transformed to () => 42 + 100

	// First create a scope with definitions of x and y
	scope := values.NewScope()
	{
		semPkg, err := runtime.AnalyzeSource(src1)
		if err != nil {
			t.Fatal(err)
		}

		itrp := interpreter.NewInterpreter(nil, nil)
		_, err = itrp.Eval(context.Background(), semPkg, scope, mockImporter{})
		if err != nil {
			t.Fatal(err)
		}
	}

	// Create a functionValue from that uses the scope we just created
	var fnVal functionValue
	{
		pkg := libflux.ParseString(src1)
		fn := libflux.ParseString(src2)
		if err := libflux.MergePackages(pkg, fn); err != nil {
			t.Fatal(err)
		}

		semPkg, err := runtime.AnalyzePackage(pkg)
		if err != nil {
			t.Fatal(err)
		}
		stmt := semPkg.Files[1].Body[0]
		fnExpr := stmt.(*semantic.ExpressionStatement).Expression.(*semantic.FunctionExpression)
		fnVal = functionValue{
			t:     fnExpr.TypeOf(),
			fn:    fnExpr,
			scope: runtimeScope{Scope: scope},
		}
	}

	var want *semantic.FunctionExpression
	{
		pkg, err := runtime.AnalyzeSource(`() => 42 + 100`)
		if err != nil {
			t.Fatal(err)
		}
		want = pkg.Files[0].Body[0].(*semantic.ExpressionStatement).Expression.(*semantic.FunctionExpression)
	}

	got, err := fnVal.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(want, got, semantictest.CmpOptions...) {
		t.Errorf("unexpected resolved function: -want/+got\n%s", cmp.Diff(want, got, semantictest.CmpOptions...))
	}
}
