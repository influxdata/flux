package testing_test

import (
	"bufio"
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/lang"
	"github.com/influxdata/flux/querytest"
	_ "github.com/influxdata/flux/stdlib" // Import the built-ins
)

func init() {
	flux.FinalizeBuiltIns()
}

var skipTests = map[string]string{
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
}

var querier = querytest.NewQuerier()

func withEachFluxFile(t testing.TB, fn func(prefix, caseName string)) {
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "testdata")
	os.Chdir(path)

	fluxFiles, err := filepath.Glob("*.flux")
	if err != nil {
		t.Fatalf("error searching for Flux files: %s", err)
	}

	for _, fluxFile := range fluxFiles {
		ext := filepath.Ext(fluxFile)
		prefix := fluxFile[0 : len(fluxFile)-len(ext)]
		_, caseName := filepath.Split(prefix)
		fn(prefix, caseName)
	}
}

func Test_QueryEndToEnd(t *testing.T) {
	withEachFluxFile(t, func(prefix, caseName string) {
		reason, skip := skipTests[caseName]

		fluxName := caseName + ".flux"
		t.Run(fluxName, func(t *testing.T) {
			if skip {
				t.Skip(reason)
			}
			testFlux(t, querier, prefix, ".flux")
		})
	})
}

func Benchmark_QueryEndToEnd(b *testing.B) {
	withEachFluxFile(b, func(prefix, caseName string) {
		reason, skip := skipTests[caseName]

		fluxName := caseName + ".flux"
		b.Run(fluxName, func(b *testing.B) {
			if skip {
				b.Skip(reason)
			}
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				testFlux(b, querier, prefix, ".flux")
			}
		})
	})
}

func testFlux(t testing.TB, querier *querytest.Querier, prefix, queryExt string) {
	q, err := ioutil.ReadFile(prefix + queryExt)
	if err != nil {
		t.Fatal(err)
	}

	if err != nil {
		t.Fatal(err)
	}

	c := lang.FluxCompiler{
		Query: string(q),
	}

	r, err := querier.C.Query(context.Background(), c)
	if err != nil {
		t.Fatalf("test error %s", err)
	}
	defer r.Done()
	result, ok := <-r.Ready()
	if !ok {
		t.Fatalf("TEST error retrieving query result: %s", r.Err())
	}

	var out bytes.Buffer
	defer func() {
		if t.Failed() {
			scanner := bufio.NewScanner(&out)
			for scanner.Scan() {
				t.Log(scanner.Text())
			}
		}
	}()

	for _, res := range result {
		if err := res.Tables().Do(func(tbl flux.Table) error {
			_, _ = execute.NewFormatter(tbl, nil).WriteTo(&out)
			return nil
		}); err != nil {
			t.Error(err)
		}
	}
}
