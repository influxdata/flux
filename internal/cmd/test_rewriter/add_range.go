package main

import (
	"fmt"
	"time"

	"github.com/influxdata/flux/ast"
	"github.com/spf13/cobra"
)

var addRangeCmd = &cobra.Command{
	Use:   "add-range [test files...]",
	Short: "Update tests that lack a call to range()",
	RunE:  addRangeE,
	Args:  cobra.MinimumNArgs(1),
}

func init() {
	rootCmd.AddCommand(addRangeCmd)
}

func addRangeE(cmd *cobra.Command, args []string) error {
	return doSubCommand(addRangeCall, args)
}

func addRangeCall(fileName string) error {
	astPkg, err := getFileAST(fileName)
	if err != nil {
		return err
	}

	var rf rangeFinder
	ast.Walk(&rf, astPkg)

	if rf.found {
		fmt.Printf("  range found, nothing to do.\n")
		return nil
	}

	// add call to range to the leafiest |> we found
	pe := rf.pipeExpr
	if pe == nil {
		fmt.Println("  no pipe expression found")
		return nil
	}

	if err := addRangeToPipeline(pe); err != nil {
		return err
	}
	if err := addDropToPipeline(pe); err != nil {
		return err
	}

	if err := rewriteFile(fileName, astPkg); err != nil {
		return nil
	}
	fmt.Printf("  Rewrote %s with a call to range() added.\n", fileName)
	return nil
}

func addRangeToPipeline(pe *ast.PipeExpression) error {
	startTime, err := time.Parse(time.RFC3339, "1980-01-01T00:00:00.000Z")
	if err != nil {
		return err
	}

	oldArg := pe.Argument
	pe.Argument = &ast.PipeExpression{
		Argument: oldArg,
		Call: &ast.CallExpression{
			Callee: &ast.Identifier{Name: "range"},
			Arguments: []ast.Expression{
				&ast.ObjectExpression{
					Properties: []*ast.Property{
						{
							Key:   &ast.Identifier{Name: "start"},
							Value: &ast.DateTimeLiteral{Value: startTime},
						},
					},
				},
			},
		},
	}
	return nil
}

func addDropToPipeline(pe *ast.PipeExpression) error {
	oldArg := pe.Argument
	pe.Argument = &ast.PipeExpression{
		Argument: oldArg,
		Call: &ast.CallExpression{
			Callee: &ast.Identifier{Name: "drop"},
			Arguments: []ast.Expression{
				&ast.ObjectExpression{
					Properties: []*ast.Property{
						{
							Key: &ast.Identifier{Name: "columns"},
							Value: &ast.ArrayExpression{
								Elements: []ast.Expression{
									&ast.StringLiteral{Value: "_start"},
									&ast.StringLiteral{Value: "_stop"},
								},
							},
						},
					},
				},
			},
		},
	}
	return nil
}

type rangeFinder struct {
	found    bool
	pipeExpr *ast.PipeExpression
}

func (r *rangeFinder) Visit(node ast.Node) ast.Visitor {
	if r.found {
		return nil
	}
	switch n := node.(type) {
	case *ast.CallExpression:
		if id, ok := n.Callee.(*ast.Identifier); ok {
			if id.Name == "range" {
				r.found = true
				return nil
			}
		}
	}
	return r
}

func (r *rangeFinder) Done(node ast.Node) {
	switch n := node.(type) {
	case *ast.PipeExpression:
		if r.pipeExpr == nil {
			r.pipeExpr = n
		}
	}

}
