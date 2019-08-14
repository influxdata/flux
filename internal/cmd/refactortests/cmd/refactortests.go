package cmd

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/ast/edit"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/csv"
	"github.com/influxdata/flux/dependencies"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/lang"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/parser"
	_ "github.com/influxdata/flux/stdlib" // Import the Flux standard library
	"github.com/influxdata/flux/stdlib/testing"
	"github.com/spf13/cobra"
)

func init() {
	flux.FinalizeBuiltIns()
}

// refactorCmd represents the refactortests command
var refactorCmd = &cobra.Command{
	Use:   "refactortests /path/to/testfiles [testname]",
	Short: "Refactor end-to-end tests",
	Long: `This utility allows the user to edit end-to-end tests in place.
It allows to edit the query under test or to replace the expected result
with the result produced by the query in the test.
`,
	Args: cobra.RangeArgs(1, 2),
	RunE: refactor,
}

func Execute() {
	if err := refactorCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func refactor(cmd *cobra.Command, args []string) error {
	var fnames []string
	path := args[0]

	if len(args) == 1 {
		fluxFiles, err := filepath.Glob(filepath.Join(path, "*.flux"))
		if err != nil {
			return err
		}
		if len(fluxFiles) == 0 {
			return fmt.Errorf("could not find any .flux files in directory \"%s\"", path)
		}
		fnames = fluxFiles
	} else {
		fnames = make([]string, len(args)-1)
		for i, fname := range args[1:] {
			fnames[i] = filepath.Join(path, fname+".flux")
		}
	}

	l := len(fnames)
	for i, fname := range fnames {
		fmt.Printf("\n\n**** (%v/%v) Starting refactor for %s ****\n\n", i+1, l, fname)
		if err := doRefactor(fname); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating output for %v: %v\n", fname, err)
			fmt.Printf("Press <return> to continue to next file\n")
			bufio.NewReader(os.Stdin).ReadString('\n')
		}
	}

	return nil
}

func doRefactor(fname string) error {
	pkg, err := loadScript(fname)
	if err != nil {
		return err
	}

	fromFile := ast.Format(pkg)
	for {
		fmt.Printf("Content of the script:\n\n%s\n\n", ast.Format(pkg))
		fmt.Println("Running...")
		rawResult, rawDiff, err := executeScript(pkg)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			y, err := yesOrNo("Do you want to reload the script from file (y/n)?", false)
			if err != nil {
				return err
			}
			if y {
				if p, err := loadScript(fname); err != nil {
					return err
				} else {
					pkg = p
					fromFile = ast.Format(pkg)
				}
				continue
			}
			// there is an error and the user doesn't want to reload, out of the loop.
			break
		} else {
			if len(rawDiff) > 0 {
				fmt.Printf("want/got not equal, printing diff (or error message):\n\n")
				fmt.Printf("%s\n\n", rawDiff)
			} else {
				fmt.Println("want/got are equal; nothing to do.")
				break
			}
		}

		fmt.Printf("The result from the script execution is:\n\n%s\n\n", rawResult)
		choice, err := updateDataOrReload()
		if err != nil {
			return err
		}
		switch choice {
		case 1:
			if err := updateOutData(pkg, rawResult); err != nil {
				return err
			}
			// data has been updated, at the next iteration want should be equal to got.
		case 2:
			if p, err := loadScript(fname); err != nil {
				return err
			} else {
				pkg = p
				fromFile = ast.Format(pkg)
			}
		}
	}

	if fromFile != ast.Format(pkg) {
		return updateScript(fname, pkg)
	}
	return nil
}

func loadScript(fname string) (*ast.Package, error) {
	bscript, err := ioutil.ReadFile(fname)
	if err != nil {
		return nil, err
	}
	script := string(bscript)

	pkg := parser.ParseSource(script)
	if ast.Check(pkg) > 0 {
		if err := ast.GetError(pkg); err != nil {
			return nil, err
		}
	}

	return pkg, nil
}

// executeScript runs the script from AST and returns the result of the testing function, the diff want/got, or error.
// If the diff is empty, this means that test went well.
func executeScript(pkg *ast.Package) (string, string, error) {
	testPkg, err := inPlaceTestGen(pkg)
	if err != nil {
		return "", "", errors.Wrap(err, codes.Inherit, "error during test generation")
	}

	c := lang.FluxCompiler{
		Query: ast.Format(testPkg),
	}

	program, err := c.Compile(context.Background())
	if p, ok := program.(lang.DependenciesAwareProgram); ok {
		p.SetExecutorDependencies(execute.Dependencies{dependencies.InterpreterDepsKey: dependencies.NewDefaultDependencies()})
	}
	if err != nil {
		fmt.Println(ast.Format(testPkg))
		return "", "", errors.Wrap(err, codes.Inherit, "error during compilation, check your script and retry")
	}

	alloc := &memory.Allocator{}
	q, err := program.Start(context.Background(), alloc)
	if err != nil {
		return "", "", errors.Wrap(err, codes.Inherit, "error while executing program")
	}
	defer q.Done()
	results := make(map[string]flux.Result)
	for r := range q.Results() {
		results[r.Name()] = r
	}
	if err := q.Err(); err != nil {
		return "", "", errors.Wrap(err, codes.Inherit, "error retrieving query result")
	}

	var diffBuf, resultBuf bytes.Buffer
	// encode diff if present
	if diff, in := results["diff"]; in {
		if err := diff.Tables().Do(func(tbl flux.Table) error {
			_, _ = execute.NewFormatter(tbl, nil).WriteTo(&diffBuf)
			return nil
		}); err != nil {
			// do not return diff error, but show it
			fmt.Fprintln(os.Stderr, errors.Wrap(err, codes.Inherit, "error while running test script"))
		}
	}

	var aee *testing.AssertEqualsError
	enc := csv.NewResultEncoder(csv.DefaultEncoderConfig())
	// encode test result if present
	if tr, in := results["_test_result"]; in {
		if _, err := enc.Encode(&resultBuf, tr); err != nil {
			fmt.Fprintln(os.Stderr, errors.Wrap(err, codes.Inherit, "encoding error while running test script"))
		}
	} else {
		// cannot use MultiResultEncoder, because it encodes errors, but I need that
		// errors to assert their type.
		fmt.Fprintln(os.Stderr, "This test doesn't use the test framework, using every result produced as output data.")
		fmt.Fprintf(os.Stderr, "The tool will check for assertEquals errors.\n\n")
		for _, r := range results {
			if _, err := enc.Encode(&resultBuf, r); err != nil {
				fmt.Fprintln(os.Stderr, errors.Wrap(err, codes.Inherit, "encoding error while running test script"))
				if e := asAssertEqualsErrors(err); e != nil {
					aee = e
				}
			}
		}
	}

	diff := diffBuf.String()
	if aee != nil && len(diff) == 0 {
		// populate diff because there was an assertion error
		// the caller should know there was a difference, indeed.
		diff = aee.Error()
	}

	return resultBuf.String(), diff, nil
}

// inPlaceTestGen adds the statements to run the test cases in pkg.
// It doesn't have side effects on pkg, rather, it returns a new, edited one.
// It avoids to run the code generation step.
func inPlaceTestGen(pkg *ast.Package) (*ast.Package, error) {
	testPkg := pkg.Copy().(*ast.Package)
	// make it a main package for execution
	testPkg.Package = "main"
	// find and edit TestStatements
	pattern := &ast.File{
		Body: []ast.Statement{
			&ast.TestStatement{},
		},
	}
	files := edit.Match(testPkg, pattern, true)

	for _, file := range files {
		f := file.(*ast.File)
		for _, stmt := range f.Body {
			if ts, ok := stmt.(*ast.TestStatement); ok {
				fnCall := &ast.ExpressionStatement{
					Expression: &ast.CallExpression{
						Callee: &ast.MemberExpression{
							Object:   &ast.Identifier{Name: "testing"},
							Property: &ast.Identifier{Name: "run"},
						},
						Arguments: []ast.Expression{
							&ast.ObjectExpression{
								Properties: []*ast.Property{
									{
										Key:   &ast.Identifier{Name: "case"},
										Value: ts.Assignment.ID,
									},
								},
							},
						},
					},
				}
				f.Body = append(f.Body, fnCall)
			}
		}
	}

	return testPkg, nil
}

func updateDataOrReload() (int, error) {
	fmt.Println("Do you want to edit the expected result (1) or change the file and reload it from disk (2)? (any other to exit)")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	if scanner.Err() != nil {
		return -1, scanner.Err()
	}
	choice, _ := strconv.Atoi(scanner.Text())
	return choice, nil
}

func updateOutData(node *ast.Package, rawResult string) error {
	as, err := getOutDataVariable(node)
	if err != nil {
		return err
	} else if as == nil {
		// no assignment chosen, stop
		return nil
	}

	as.Init = &ast.StringLiteral{Value: "\n" + rawResult}
	fmt.Println("Data has been updated")
	return nil
}

func updateScript(fname string, script *ast.Package) error {
	fmt.Printf("Content of the new script:\n\n%s\n\n", ast.Format(script))
	y, err := yesOrNo("Do you want to update the script (y/n)?", true)
	if err != nil {
		return err
	}
	if !y {
		fmt.Println("Script NOT updated")
		return nil
	}
	err = ioutil.WriteFile(fname, []byte(ast.Format(script)), 0644)
	if err != nil {
		return err
	}
	fmt.Printf("%s updated\n", fname)
	return nil
}

func getOutDataVariable(node ast.Node) (*ast.VariableAssignment, error) {
	fmt.Println(`Provide the identifier for the variable holding the expected result.`)
	fmt.Println(`The default is "outData", press <enter> to confirm, or enter a new identifier:`)
	id := "outData"
	reader := bufio.NewReader(os.Stdin)
	if newId, err := reader.ReadString('\n'); err != nil {
		return nil, err
	} else {
		if newId == "\n" {
			if as := doGetOutDataVariable(node, id); as != nil {
				return as, nil
			}
		} else {
			id = strings.Trim(newId, "\n")
		}
	}

	for {
		as := doGetOutDataVariable(node, id)
		if as != nil {
			return as, nil
		}

		fmt.Println("Provide new identifier, or press <return> to exit:")
		if newId, err := reader.ReadString('\n'); err != nil {
			return nil, err
		} else {
			if newId == "\n" {
				break
			}
			id = strings.Trim(newId, "\n")
		}
	}

	return nil, nil
}

func doGetOutDataVariable(node ast.Node, id string) *ast.VariableAssignment {
	match := edit.Match(node, &ast.VariableAssignment{ID: &ast.Identifier{Name: id}}, false)
	if len(match) == 0 {
		fmt.Printf("No \"%s\" found, retry\n", id)
		return nil
	}
	if len(match) > 1 {
		fmt.Println("More than one match found, editing the first occurrence.")
	}
	return match[0].(*ast.VariableAssignment)
}

func yesOrNo(question string, forceY bool) (bool, error) {
	fmt.Println(question)
	reader := bufio.NewReader(os.Stdin)
	if ans, err := reader.ReadString('\n'); err != nil {
		return false, err
	} else {
		return ans == "y\n" || (!forceY && ans == "\n"), nil
	}
}

// asAssertEqualsError will return the error as a *testing.AssertEqualsError
// if it is in the error chain.
func asAssertEqualsErrors(err error) *testing.AssertEqualsError {
	for err != nil {
		if aee, ok := err.(*testing.AssertEqualsError); ok {
			return aee
		}

		wrapErr, ok := err.(interface{ Unwrap() error })
		if !ok {
			break
		}
		err = wrapErr.Unwrap()
	}
	return nil
}
