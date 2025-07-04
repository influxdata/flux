package stdlib_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	arrowmem "github.com/apache/arrow-go/v18/arrow/memory"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/dependency"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/lang"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/stdlib"
)

func init() {
	runtime.FinalizeBuiltIns()
}

// list of end-to-end tests that are meant to be skipped and not run for various reasons
var skip = map[string]map[string]string{
	"universe": {
		"string_max":                  "error: invalid use of function: *functions.MaxSelector has no implementation for type string (https://github.com/influxdata/platform/issues/224)",
		"null_as_value":               "null not supported as value in influxql (https://github.com/influxdata/platform/issues/353)",
		"string_interp":               "string interpolation not working as expected in flux (https://github.com/influxdata/platform/issues/404)",
		"to":                          "to functions are not supported in the testing framework (https://github.com/influxdata/flux/issues/77)",
		"covariance_missing_column_1": "need to support known errors in new test framework (https://github.com/influxdata/flux/issues/536)",
		"covariance_missing_column_2": "need to support known errors in new test framework (https://github.com/influxdata/flux/issues/536)",
		"drop_before_rename":          "need to support known errors in new test framework (https://github.com/influxdata/flux/issues/536)",
		"drop_referenced":             "need to support known errors in new test framework (https://github.com/influxdata/flux/issues/536)",
		"yield":                       "yield requires special test case (https://github.com/influxdata/flux/issues/535)",
		"task_per_line":               "join produces inconsistent/racy results when table schemas do not match (https://github.com/influxdata/flux/issues/855)",
		"integral_columns":            "aggregates changed to operate on just a single column",
	},
	"http": {
		"http_endpoint": "need ability to test side effects in e2e tests: https://github.com/influxdata/flux/issues/1723)",
	},
	"interval": {
		"interval": "switch these tests cases to produce a non-table stream once that is supported (https://github.com/influxdata/flux/issues/535)",
	},
	"testing/chronograf": {
		"measurement_tag_keys":   "unskip chronograf flux tests once filter is refactored (https://github.com/influxdata/flux/issues/1289)",
		"aggregate_window_mean":  "unskip chronograf flux tests once filter is refactored (https://github.com/influxdata/flux/issues/1289)",
		"aggregate_window_count": "unskip chronograf flux tests once filter is refactored (https://github.com/influxdata/flux/issues/1289)",
	},
	"testing/pandas": {
		"extract_regexp_findStringIndex": "pandas. map does not correctly handled returned arrays (https://github.com/influxdata/flux/issues/1387)",
		"partition_strings_splitN":       "pandas. map does not correctly handled returned arrays (https://github.com/influxdata/flux/issues/1387)",
	},
}

var skipMemoryChecks = []string{
	"contrib/RohanSreerama5/naiveBayesClassifier/bayes",
	"contrib/anaisdg/statsmodels/linearreg",
	"experimental/geo",
	"experimental/geo/filterRowsPivoted",
	"experimental/geo/filterRowsStrict",
	"experimental/geo/filterRowsNotStrict",
	"experimental/geo/gridFilterLevel",
	"experimental/geo/gridFilter",
	"experimental/geo/groupByArea",
	"experimental/geo/shapeDataWithFilter",
	"influxdata/influxdb/sample/alignToNow",
	"universe/dynamic_query",
	"universe/table_fns",
	"universe/table_fns_findcolumn_map",
	"universe/table_fns_findrecord_map",
}

func TestFluxEndToEnd(t *testing.T) {
	pkgs, err := stdlib.FindTestPackages(".")
	if err != nil {
		t.Error(err)
		return
	}

	if len(pkgs) < 49 {
		t.Errorf("fewer test packages than expected. Check that there isn't an issue with findTestPackages or updates the expected value\nFound: %d", len(pkgs))
		return
	}

	runEndToEnd(t, pkgs)
}

func runEndToEnd(t *testing.T, pkgs []*ast.Package) {
	for _, pkg := range pkgs {
		test := func(t *testing.T, f func(t *testing.T)) {
			t.Run(pkg.Path, f)
		}
		if pkg.Path == "universe" {
			test = func(t *testing.T, f func(t *testing.T)) {
				f(t)
			}
		}

		test(t, func(t *testing.T) {
			for _, file := range pkg.Files {
				name := strings.TrimSuffix(file.Name, "_test.flux")
				t.Run(name, func(t *testing.T) {
					if reason, ok := skip[pkg.Path][name]; ok {
						t.Skip(reason)
					}
					fullname := fmt.Sprintf("%s/%s", pkg.Path, name)
					testFlux(t, fullname, file)
				})
			}
		})
	}
}

func makeTestPackage(file *ast.File) *ast.Package {
	file = file.Copy().(*ast.File)
	file.Package.Name.Name = "main"
	pkg := &ast.Package{
		Package: "main",
		Files:   []*ast.File{file},
	}
	return pkg
}

func testFlux(t testing.TB, name string, file *ast.File) flux.Statistics {
	pkg := makeTestPackage(file)
	pkg.Files = append(pkg.Files, stdlib.TestingRunCalls(pkg))
	bs, err := json.Marshal(pkg)
	if err != nil {
		t.Fatal(err)
	}
	c := lang.ASTCompiler{AST: bs}

	// testing.run
	stats := doTestRun(t, name, c)

	// testing.inspect
	if t.Failed() {
		// Rerun the test case using testing.inspect
		pkg.Files[len(pkg.Files)-1] = stdlib.TestingInspectCalls(pkg)
		bs, err := json.Marshal(pkg)
		if err != nil {
			t.Fatal(err)
		}
		c := lang.ASTCompiler{AST: bs}
		stats = doTestInspect(t, c)
	}
	return stats
}

func doTestRun(t testing.TB, name string, c flux.Compiler) flux.Statistics {
	program, err := c.Compile(context.Background(), runtime.Default)
	if err != nil {
		t.Fatalf("unexpected error while compiling query: %v", err)
	}

	ctx, deps := dependency.Inject(context.Background(), executetest.NewTestExecuteDependencies())
	defer deps.Finish()

	var alloc memory.Allocator
	if execute.ContainsStr(skipMemoryChecks, name) {
		alloc = memory.DefaultAllocator
	} else {
		mem := arrowmem.NewCheckedAllocator(memory.DefaultAllocator)
		defer mem.AssertSize(t, 0)
		alloc = memory.NewResourceAllocator(mem)
	}

	r, err := program.Start(ctx, alloc)
	if err != nil {
		t.Fatalf("unexpected error while executing testing.run: %v", err)
	}
	defer r.Done()

	// Read all results checking for errors
	for res := range r.Results() {
		err := res.Tables().Do(func(tbl flux.Table) error {
			tbl.Done()
			return nil
		})
		if err != nil {
			t.Error(err)
		}
	}
	if err := r.Err(); err != nil {
		t.Fatalf("unexpected error retrieving testing.run result: %s", err)
	}
	r.Done()
	return r.Statistics()
}

func doTestInspect(t testing.TB, c flux.Compiler) flux.Statistics {
	program, err := c.Compile(context.Background(), runtime.Default)
	if err != nil {
		t.Fatalf("unexpected error while compiling query: %v", err)
	}
	ctx, deps := dependency.Inject(context.Background(), executetest.NewTestExecuteDependencies())
	defer deps.Finish()

	r, err := program.Start(ctx, memory.DefaultAllocator)
	if err != nil {
		t.Fatalf("unexpected error while executing testing.inspect: %v", err)
	}
	defer r.Done()

	// Read all results and format them
	var out bytes.Buffer
	for res := range r.Results() {
		if err := execute.FormatResult(&out, res); err != nil {
			t.Error(err)
		}
	}
	if err := r.Err(); err != nil {
		t.Fatalf("unexpected error retrieving testing.inspect result: %s", err)
	}
	t.Log(out.String())
	r.Done()
	return r.Statistics()
}
