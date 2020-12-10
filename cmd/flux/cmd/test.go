package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/ast/edit"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/execute/table"
	"github.com/influxdata/flux/fluxinit"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/lang"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/parser"
	"github.com/influxdata/flux/runtime"
	"github.com/spf13/cobra"
)

var testCommand = &cobra.Command{
	Use:   "test",
	Short: "Run flux tests",
	Long:  "Run flux tests",
	Run: func(cmd *cobra.Command, args []string) {
		fluxinit.FluxInit()
		runFluxTests()
	},
}

var testNames []string
var rootDir string
var verbosity int

func init() {
	rootCmd.AddCommand(testCommand)
	testCommand.Flags().StringVarP(&rootDir, "path", "p", "./stdlib", "The root level directory for all packages.")
	testCommand.Flags().StringSliceVar(&testNames, "test", []string{}, "The name of a specific test to run.")
	testCommand.Flags().CountVarP(&verbosity, "verbose", "v", "verbose (-v, or -vv)")
}

// runFluxTests invokes the test runner.
func runFluxTests() {
	runner := NewTestRunner(NewTestReporter(verbosity))
	err := runner.Gather(rootDir, testNames)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	runner.Run(verbosity)

	runner.Finish()
}

// Test wraps the functionality of a single testcase statement,
// to handle its execution and its pass/fail state.
type Test struct {
	ast *ast.Package
	err error
}

// NewTest creates a new Test instance from an ast.Package.
func NewTest(ast *ast.Package) Test {
	return Test{
		ast: ast,
	}
}

// Get the name of the Test.
func (t *Test) Name() string {
	return t.ast.Files[0].Name
}

// Get the error from the test, if one exists.
func (t *Test) Error() error {
	return t.err
}

// Run the test, saving the error to the err property of the struct.
func (t *Test) Run() {
	jsonAST, err := json.Marshal(t.ast)
	if err != nil {
		t.err = err
		return
	}
	c := lang.ASTCompiler{AST: jsonAST}

	ctx := executetest.NewTestExecuteDependencies().Inject(context.Background())
	program, err := c.Compile(ctx, runtime.Default)
	if err != nil {
		t.err = errors.Wrap(err, codes.Invalid, "failed to compile")
		return
	}

	alloc := &memory.Allocator{}
	query, err := program.Start(ctx, alloc)
	if err != nil {
		t.err = errors.Wrap(err, codes.Inherit, "error while executing program")
		return
	}
	defer query.Done()

	results := flux.NewResultIteratorFromQuery(query)
	for results.More() {
		result := results.Next()
		err := result.Tables().Do(func(tbl flux.Table) error {
			// The data returned here is the result of `testing.diff`, so any result means that
			// a comparison of two tables showed inequality. Capture that inequality as part of the error.
			// XXX: rockstar (08 Dec 2020) - This could use some ergonomic work, as the diff output
			// is not exactly "human readable."
			return fmt.Errorf("%s", table.Stringify(tbl))
		})
		if err != nil {
			t.err = err
		}
	}
}

// contains checks a slice of strings for a given string.
func contains(names []string, name string) bool {
	for _, n := range names {
		if n == name {
			return true
		}
	}
	return false
}

// TestRunner gathers and runs all tests.
type TestRunner struct {
	tests    []*Test
	reporter TestReporter
}

// NewTestRunner returns a new TestRunner.
func NewTestRunner(reporter TestReporter) TestRunner {
	return TestRunner{tests: []*Test{}, reporter: reporter}
}

// Gather gathers all tests from the filesystem and creates Test instances
// from that info.
func (t *TestRunner) Gather(rootDir string, names []string) error {
	root, err := filepath.Abs(rootDir)
	if err != nil {
		return err
	}

	return filepath.Walk(
		root,
		func(path string, info os.FileInfo, err error) error {
			if strings.HasSuffix(path, "_test.flux") {
				source, err := ioutil.ReadFile(path)
				if err != nil {
					return err
				}

				baseAST := parser.ParseSource(string(source))
				asts, err := edit.TestcaseTransform(baseAST)
				if err != nil {
					return err
				}
				for _, ast := range asts {
					test := NewTest(ast)
					if len(testNames) == 0 || contains(testNames, test.Name()) {
						t.tests = append(t.tests, &test)
					}
				}
			}
			return nil
		})
}

// Run runs all tests, reporting their results.
func (t *TestRunner) Run(verbosity int) {
	for _, test := range t.tests {
		test.Run()
		t.reporter.ReportTestRun(test)
	}
}

// Finish summarizes the test run, and returns the
// exit code based on success for failure.
func (t *TestRunner) Finish() {
	t.reporter.Summarize(t.tests)
	for _, test := range t.tests {
		if test.Error() != nil {
			os.Exit(1)
		}
	}
	os.Exit(0)
}

// TestReporter handles reporting of test results.
type TestReporter struct {
	verbosity int
}

// NewTestReporter creates a new TestReporter with a provided verbosity.
func NewTestReporter(verbosity int) TestReporter {
	return TestReporter{verbosity: verbosity}
}

// ReportTestRun reports the result a single test run, intended to be run as
// each test is run.
func (t *TestReporter) ReportTestRun(test *Test) {
	if t.verbosity == 0 {
		if test.Error() != nil {
			fmt.Print("x")
		} else {
			fmt.Print(".")
		}
	} else {
		if test.Error() != nil {
			fmt.Printf("%s...fail\n", test.Name())
		} else {
			fmt.Printf("%s...success\n", test.Name())
		}
	}
}

// Summarize summarizes the test run.
func (t *TestReporter) Summarize(tests []*Test) {
	failures := 0
	for _, test := range tests {
		if test.Error() != nil {
			failures = failures + 1
		}
	}
	fmt.Printf("\n---\nRan %d tests with %d failure(s)\n", len(tests), failures)
}
