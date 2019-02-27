package stdlib_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/lang"
	"github.com/influxdata/flux/querytest"
	"github.com/influxdata/flux/stdlib"
)

func init() {
	flux.FinalizeBuiltIns()
}

// list of end-to-end tests that are meant to be skipped and not run for various reasons
var skip = map[string]string{
	"string_max":                  "error: invalid use of function: *functions.MaxSelector has no implementation for type string (https://github.com/influxdata/platform/issues/224)",
	"null_as_value":               "null not supported as value in influxql (https://github.com/influxdata/platform/issues/353)",
	"string_interp":               "string interpolation not working as expected in flux (https://github.com/influxdata/platform/issues/404)",
	"to":                          "to functions are not supported in the testing framework (https://github.com/influxdata/flux/issues/77)",
	"covariance_missing_column_1": "need to support known errors in new test framework (https://github.com/influxdata/flux/issues/536)",
	"covariance_missing_column_2": "need to support known errors in new test framework (https://github.com/influxdata/flux/issues/536)",
	"drop_before_rename":          "need to support known errors in new test framework (https://github.com/influxdata/flux/issues/536)",
	"drop_referenced":             "need to support known errors in new test framework (https://github.com/influxdata/flux/issues/536)",
	"drop_non_existent":           "need to support known errors in new test framework (https://github.com/influxdata/flux/issues/536)",
	"keep_non_existent":           "need to support known errors in new test framework (https://github.com/influxdata/flux/issues/536)",
	"yield":                       "yield requires special test case (https://github.com/influxdata/flux/issues/535)",
	"task_per_line":               "join produces inconsistent/racy results when table schemas do not match (https://github.com/influxdata/flux/issues/855)",
	"rowfn_with_import":           "imported libraries are not visible in user-defined functions (https://github.com/influxdata/flux/issues/1000)",
	"string_trim":                 "cannot reference a package function from within a row function",
}

var querier = querytest.NewQuerier()

func TestFluxEndToEnd(t *testing.T) {
	runEndToEnd(t, querier, stdlib.FluxTestPackages)
}
func BenchmarkFluxEndToEnd(b *testing.B) {
	benchEndToEnd(b, querier, stdlib.FluxTestPackages)
}

func runEndToEnd(t *testing.T, querier *querytest.Querier, pkgs []*ast.Package) {
	for _, pkg := range pkgs {
		pkg := pkg.Copy().(*ast.Package)
		name := pkg.Files[0].Name
		t.Run(name, func(t *testing.T) {
			n := strings.TrimSuffix(name, ".flux")
			if reason, ok := skip[n]; ok {
				t.Skip(reason)
			}
			testFlux(t, querier, pkg)
		})
	}
}

func benchEndToEnd(b *testing.B, querier *querytest.Querier, pkgs []*ast.Package) {
	for _, pkg := range pkgs {
		pkg := pkg.Copy().(*ast.Package)
		name := pkg.Files[0].Name
		b.Run(name, func(b *testing.B) {
			n := strings.TrimSuffix(name, ".flux")
			if reason, ok := skip[n]; ok {
				b.Skip(reason)
			}
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				testFlux(b, querier, pkg)
			}
		})
	}
}

func testFlux(t testing.TB, querier *querytest.Querier, pkg *ast.Package) {
	pkg.Files = append(pkg.Files, stdlib.TestingRunCalls(pkg))
	c := lang.ASTCompiler{AST: pkg}

	// testing.run
	doTestRun(t, querier, c)

	// testing.inspect
	if t.Failed() {
		// Rerun the test case using testing.inspect
		pkg.Files[len(pkg.Files)-1] = stdlib.TestingInspectCalls(pkg)
		c := lang.ASTCompiler{AST: pkg}
		doTestInspect(t, querier, c)
	}
}

func doTestRun(t testing.TB, querier *querytest.Querier, c flux.Compiler) {
	r, err := querier.C.Query(context.Background(), c)
	if err != nil {
		t.Fatalf("unexpected error while executing testing.run: %v", err)
	}
	defer r.Done()
	result, ok := <-r.Ready()
	if !ok {
		t.Fatalf("unexpected error retrieving testing.run result: %s", r.Err())
	}

	// Read all results checking for errors
	for _, res := range result {
		err := res.Tables().Do(func(flux.Table) error {
			return nil
		})
		if err != nil {
			t.Error(err)
		}
	}
}
func doTestInspect(t testing.TB, querier *querytest.Querier, c flux.Compiler) {
	r, err := querier.C.Query(context.Background(), c)
	if err != nil {
		t.Fatalf("unexpected error while executing testing.inspect: %v", err)
	}
	defer r.Done()
	result, ok := <-r.Ready()
	if !ok {
		t.Fatalf("unexpected error retrieving testing.inspect result: %s", r.Err())
	}
	// Read all results and format them
	var out bytes.Buffer
	for _, res := range result {
		if err := execute.FormatResult(&out, res); err != nil {
			t.Error(err)
		}
	}
	t.Log(out.String())
}
