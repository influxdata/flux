// +build !libflux

package parser

import (
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/internal/parser"
	"github.com/influxdata/flux/internal/token"
)

func parseFile(f *token.File, src []byte) (*ast.File, error) {
	return parser.ParseFile(f, src), nil
}
