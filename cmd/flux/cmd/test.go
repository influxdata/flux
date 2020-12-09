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

type Test struct {
	ast *ast.Package
	Err error
}

func NewTest(ast *ast.Package) Test {
	return Test{
		ast: ast,
	}
}

func (t *Test) Name() string {
	return t.ast.Files[0].Name
}

func (t *Test) Run() {
	jsonAST, Err := json.Marshal(t.ast)
	if Err != nil {
		t.Err = Err
		return
	}
	c := lang.ASTCompiler{AST: jsonAST}

	ctx := executetest.NewTestExecuteDependencies().Inject(context.Background())
	program, Err := c.Compile(ctx, runtime.Default)
	if Err != nil {
		t.Err = errors.Wrap(Err, codes.Invalid, "failed to compile")
	}

	alloc := &memory.Allocator{}
	result, Err := program.Start(ctx, alloc)
	if Err != nil {
		// XXX: rockstar (8 Dec 2020) - Not all tests should return streaming data.
		if !strings.Contains(Err.Error(), "this Flux script returns no streaming data") {
			t.Err = errors.Wrap(Err, codes.Inherit, "error while executing program")
			return
		}
	}
	defer result.Done()

	for res := range result.Results() {
		Err := res.Tables().Do(func(tbl flux.Table) error {
			// If there *is* streaming data from the flux test, it is assumed to come from
			// `testing.diff`, and if that returns tables, that they are showing the failed diff.
			// XXX: rockstar (08 Dec 2020) - This could use some ergonomic work, as the diff output
			// is not exactly "human readable."
			return fmt.Errorf("%s", table.Stringify(tbl))
		})
		if Err != nil {
			t.Err = Err
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
	root, Err := filepath.Abs("./stdlib")
	if Err != nil {
		fmt.Println(Err)
		os.Exit(1)
	}

	tests := []Test{}

	filepath.Walk(
		root,
		func(path string, info os.FileInfo, Err error) error {
			if strings.HasSuffix(path, "_test.flux") {
				source, Err := ioutil.ReadFile(path)
				if Err != nil {
					fmt.Println(Err)
					return Err
				}

				baseAST := parser.ParseSource(string(source))
				asts, Err := edit.TestcaseTransform(baseAST)
				if Err != nil {
					return Err
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
		if test.Err != nil {
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

`, test.Name(), test.Err.Error())
		}
	}
	fmt.Printf("\n---\nRan %d tests with %d failures.\n", len(tests), len(failures))
}
