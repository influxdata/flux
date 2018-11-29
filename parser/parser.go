// +build !parser_debug

package parser

//go:generate pigeon -optimize-parser -o flux.go flux.peg

import (
	"github.com/influxdata/flux/ast"
	fastparser "github.com/influxdata/flux/internal/parser"
	"github.com/influxdata/flux/internal/scanner"
	"github.com/influxdata/flux/internal/token"
)

// NewAST parses Flux query and produces an ast.Program
func NewAST(flux string, opts ...Option) (*ast.Program, error) {
	fset := token.NewFileSet()
	f := fset.AddFile("", -1, len(flux))
	s := scanner.New(f, []byte(flux))
	return fastparser.NewAST(s), nil
}
