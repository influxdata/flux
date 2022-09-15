package runtime

import (
	"context"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/libflux/go/libflux"
	"github.com/influxdata/flux/parser"
)

// Parse parses a Flux script and produces an ast.Package.
func Parse(ctx context.Context, flux string) (flux.ASTHandle, error) {
	astPkg, err := parser.ParseToHandle(ctx, []byte(flux))
	if err != nil {
		return nil, err
	}
	return astPkg, nil
}

func ParseToJSON(ctx context.Context, flux string) ([]byte, error) {
	h, err := Parse(ctx, flux)
	if err != nil {
		return nil, err
	}
	return parser.HandleToJSON(h)
}

func MergePackages(dst, src flux.ASTHandle) error {
	dstPkg, srcPkg := dst.(*libflux.ASTPkg), src.(*libflux.ASTPkg)
	return libflux.MergePackages(dstPkg, srcPkg)
}
