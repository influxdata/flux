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

/* Test wraps the functionality of a single `testcase` statement,
   to handle its execution and its pass/fail state.
*/
type Test struct {
	ast *ast.Package
	err error
}

/* Create a new `Test` instance from an ast.Package. */
func NewTest(ast *ast.Package) Test {
	return Test{
		ast: ast,
	}
}

/* Get the name of the `Test` */
func (t *Test) Name() string {
	return t.ast.Files[0].Name
}

/* Get the error from the test, if one exists. */
func (t *Test) Error() error {
	return t.err
}

/* Run the test, saving the error to the `err` property of the struct. */
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

var testCommand = &cobra.Command{
	Use:   "test",
	Short: "Run flux tests",
	Long:  "Run flux tests",
	Run: func(cmd *cobra.Command, args []string) {
		fluxinit.FluxInit()
		runFluxTests()
	},
}

func init() {
	rootCmd.AddCommand(testCommand)
}

func runFluxTests() {
	root, err := filepath.Abs("./stdlib")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	tests := []Test{}

	filepath.Walk(
		root,
		func(path string, info os.FileInfo, err error) error {
			if strings.HasSuffix(path, "_test.flux") {
				source, err := ioutil.ReadFile(path)
				if err != nil {
					fmt.Println(err)
					return err
				}

				baseAST := parser.ParseSource(string(source))
				asts, err := edit.TestcaseTransform(baseAST)
				if err != nil {
					return err
				}
				for _, ast := range asts {
					test := NewTest(ast)
					tests = append(tests, test)
				}
			}
			return nil
		})

	failures := []Test{}
	for _, test := range tests {
		test.Run()
		if test.Error() != nil {
			failures = append(failures, test)
			fmt.Print("x")
		} else {
			fmt.Print(".")
		}
	}
	fmt.Print("\n")

	// XXX: rockstar (09 Dec 2020) - This logic should be abstracted out
	// into a test reporter interface.
	if len(failures) > 0 {
		for _, test := range failures {
			fmt.Printf(`%s
----
%s

`, test.Name(), test.Error())
		}
	}
	fmt.Printf("\n---\nRan %d tests with %d failures.\n", len(tests), len(failures))
}
