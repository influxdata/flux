package ast_test

import (
	"bytes"
	"testing"

	"github.com/influxdata/flux/ast"
)

func TestPrintErrors(t *testing.T) {
	program := &ast.Program{
		Body: []ast.Statement{
			&ast.BadStatement{
				BaseNode: ast.BaseNode{
					Loc: &ast.SourceLocation{
						Start: ast.Position{
							Line:   1,
							Column: 1,
						},
					},
					Errors: []ast.Error{
						{Msg: "invalid statement: @"},
					},
				},
			},
			&ast.BadStatement{
				BaseNode: ast.BaseNode{
					Loc: &ast.SourceLocation{
						Start: ast.Position{
							Line:   2,
							Column: 7,
						},
					},
					Errors: []ast.Error{
						{Msg: "invalid statement: &"},
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	ast.PrintErrors(&buf, program)

	if got, want := buf.String(), `error:1:1: invalid statement: @
error:2:7: invalid statement: &
`; want != got {
		t.Errorf("unexpected output -want/+got\n\t- %q\n\t+ %q", want, got)
	}
}
