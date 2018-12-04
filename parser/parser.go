package parser

import (
	"github.com/influxdata/flux/ast"
	fastparser "github.com/influxdata/flux/internal/parser"
	"github.com/influxdata/flux/internal/scanner"
	"github.com/influxdata/flux/internal/token"
)

// NewAST parses Flux query and produces an ast.Program
func NewAST(flux string) (*ast.Program, error) {
	fset := token.NewFileSet()
	f := fset.AddFile("", -1, len(flux))
	s := scanner.New(f, []byte(flux))
	return fastparser.NewAST(s), nil
}
