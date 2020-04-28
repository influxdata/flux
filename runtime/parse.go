package runtime

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/libflux/go/libflux"
	"github.com/influxdata/flux/parser"
)

// Parse parses a Flux script and produces an ast.Package.
func Parse(flux string) (flux.ASTHandle, error) {
	astPkg, err := parser.ParseToHandle([]byte(flux))
	if err != nil {
		return nil, err
	}
	return astPkg, nil
}

func ParseToJSON(flux string) ([]byte, error) {
	h, err := Parse(flux)
	if err != nil {
		return nil, err
	}
	return parser.HandleToJSON(h)
}

func MergePackages(dst, src flux.ASTHandle) error {
	dstPkg, srcPkg := dst.(*libflux.ASTPkg), src.(*libflux.ASTPkg)
	return libflux.MergePackages(dstPkg, srcPkg)
}
