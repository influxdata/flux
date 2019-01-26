package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/csv"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/parser"
	"github.com/influxdata/flux/semantic"

	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run a Flux test file",
	Long:  "Run a Flux test file",
	Args:  cobra.ExactArgs(1),
	RunE:  test,
}

func init() {
	rootCmd.AddCommand(testCmd)
}

func test(cmd *cobra.Command, args []string) error {
	testFile := args[0]

	sourceBytes, err := ioutil.ReadFile(testFile)
	if err != nil {
		return err
	}

	source := string(sourceBytes)

	astPkg := parser.ParseSource(source)
	if ast.Check(astPkg) > 0 {
		return ast.GetError(astPkg)
	}

	semPkg, err := semantic.New(astPkg)
	if err != nil {
		return err
	}

	tests := testNames(semPkg.Files[0].Body)

	itrp := interpreter.NewInterpreter()
	universe := flux.Prelude()

	if _, err := itrp.Eval(semPkg, universe, flux.StdLib()); err != nil {
		return err
	}

	testCode := &semantic.Package{
		Files: []*semantic.File{
			{
				Imports: []*semantic.ImportDeclaration{
					{
						Path: &semantic.StringLiteral{Value: "testing"},
					},
				},
			},
		},
	}

	for _, name := range tests {
		testCode.Files[0].Body = append(testCode.Files[0].Body, &semantic.ExpressionStatement{
			Expression: &semantic.CallExpression{
				Callee: &semantic.MemberExpression{
					Object: &semantic.IdentifierExpression{
						Name: "testing",
					},
					Property: "test",
				},
				Arguments: &semantic.ObjectExpression{
					Properties: []*semantic.Property{
						{
							Key: &semantic.Identifier{Name: "case"},
							Value: &semantic.IdentifierExpression{
								Name: name,
							},
						},
					},
				},
			},
		})
	}

	c := fluxTestComiler{
		pkg:   testCode,
		scope: universe,
	}

	querier := NewQuerier()
	result, err := querier.Query(context.Background(), c)
	if err != nil {
		return err
	}
	defer result.Release()

	encoder := csv.NewMultiResultEncoder(csv.DefaultEncoderConfig())
	_, err = encoder.Encode(os.Stdout, result)
	return err
}

func testNames(stmts []semantic.Statement) []string {
	var names []string
	for _, s := range stmts {
		if n, ok := s.(*semantic.TestStatement); ok {
			names = append(names, n.Assignment.Identifier.Name)
		}
	}
	return names
}

type fluxTestComiler struct {
	pkg   *semantic.Package
	scope interpreter.Scope
}

func (c fluxTestComiler) Compile(ctx context.Context) (*flux.Spec, error) {
	itrp := interpreter.NewInterpreter()
	universe := flux.Prelude()

	sideEffects, err := itrp.Eval(c.pkg, universe, flux.StdLib())
	if err != nil {
		return nil, err
	}

	nowOpt, ok := universe.Lookup("now")
	if !ok {
		return nil, fmt.Errorf("now option not set")
	}

	nowTime, err := nowOpt.Function().Call(nil)
	if err != nil {
		return nil, err
	}

	return flux.ToSpec(sideEffects, nowTime.Time().Time())
}

func (c fluxTestComiler) CompilerType() flux.CompilerType {
	return "test"
}
