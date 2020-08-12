package values_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	_ "github.com/influxdata/flux/builtin"
	"github.com/influxdata/flux/dependencies/dependenciestest"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/libflux/go/libflux"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/semantic/semantictest"
	"github.com/influxdata/flux/values"
)

var (
	prelude values.Scope
	stdlib  interpreter.Importer
)

// functionExpression takes a set of Flux source files and
// returns a type function expression. Having multiple sources allows for
// multiple definitions, eg.
//
// file1.flux
//    x = 53
// file2.flux
//    (y) => y + x
//
// This is pretty similar to executetest.FunctionExpression, except for it takes
// multiple files, and also will not automatically resolve the semantic graph,
// which we don't want, because we are expliticly testing that here.
func functionExpression(t testing.TB, srcFiles ...string) *semantic.FunctionExpression {
	if stdlib == nil {
		stdlib = runtime.StdLib()
	}
	if prelude == nil {
		prelude = values.NewScope()
		for _, path := range []string{"universe", "influxdata/influxdb"} {
			p, err := stdlib.ImportPackageObject(path)
			if err != nil {
				t.Fatalf("error importing prelude package %q: %s", path, err)
			}
			p.Range(prelude.Set)
		}
	}

	pkg := libflux.ParseString(srcFiles[0])
	for _, src := range srcFiles[1:] {
		func() {
			f := libflux.ParseString(src)
			defer f.Free()
			if err := libflux.MergePackages(pkg, f); err != nil {
				t.Fatal(err)
			}
		}()
	}

	semPkg, err := runtime.AnalyzePackage(pkg)
	if err != nil {
		t.Fatal(err)
	}

	// The last statement of the last package is the function expression we want.
	stmts := semPkg.Files[len(srcFiles)-1].Body
	fnExpr := stmts[len(stmts)-1].(*semantic.ExpressionStatement).Expression.(*semantic.FunctionExpression)
	return fnExpr
}

func TestResolveFunction(t *testing.T) {
	testcases := []struct {
		name string
		env  string
		fn   string
		want string
	}{
		{
			name: "simple assignment",
			env:  "x = 42",
			fn:   "(r) => r + x",
			want: "(r) => r + 42",
		},
		{
			name: "object assignment",
			env:  "v = {env: 42}",
			fn:   "(r) => r + v.env",
			want: "(r) => r + 42",
		},
		{
			name: "option assignment",
			env:  `option v = {env: "acc"}`,
			fn:   "(r) => r.env == v.env",
			want: `(r) => r.env == "acc"`,
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// Evaluate script with a function definition.
			ctx := dependenciestest.Default().Inject(context.Background())
			_, scope, err := runtime.Eval(ctx, tc.env)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			fe := functionExpression(t, tc.env, tc.fn)
			got, err := values.ResolveFunction(scope, fe)
			if err != nil {
				t.Fatalf("could not resolve function: %s", err)
			}

			want := executetest.FunctionExpression(t, tc.want)
			if !cmp.Equal(want, got, semantictest.CmpOptions...) {
				t.Errorf("unexpected resoved function: -want/+got\n%s", cmp.Diff(want, got, semantictest.CmpOptions...))
			}
		})
	}
}
