package runtime

import (
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/parser"
)

// Parse parses a Flux script and produces an ast.Package.
func Parse(flux string) (*ast.Package, error) {
	astPkg := parser.ParseSource(flux)
	if ast.Check(astPkg) > 0 {
		return nil, ast.GetError(astPkg)
	}

	return astPkg, nil
}
