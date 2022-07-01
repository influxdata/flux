package ast_test

import (
	"bytes"
	"testing"

	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
)

func TestPrintErrors(t *testing.T) {
	file := &ast.File{
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
	ast.PrintErrors(&buf, file)

	if got, want := buf.String(), `error:1:1: invalid statement: @
error:2:7: invalid statement: &
`; want != got {
		t.Errorf("unexpected output -want/+got\n\t- %q\n\t+ %q", want, got)
	}

	theErr := ast.GetError(file)
	if got, want := errors.Code(theErr), codes.Invalid; got != want {
		t.Errorf("wanted error code: %q, got %q", want.String(), got.String())
	}
}
