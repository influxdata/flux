package runtime

import (
	"github.com/mvn-trinhnguyen2-dn/flux"
	"github.com/mvn-trinhnguyen2-dn/flux/libflux/go/libflux"
	"github.com/mvn-trinhnguyen2-dn/flux/semantic"
)

// AnalyzeSource parses and analyzes the given Flux source,
// using libflux.
func AnalyzeSource(fluxSrc string) (*semantic.Package, error) {
	ast := libflux.ParseString(fluxSrc)
	return AnalyzePackage(ast)
}

func AnalyzePackage(astPkg flux.ASTHandle) (*semantic.Package, error) {
	hdl := astPkg.(*libflux.ASTPkg)
	defer hdl.Free()
	sem, err := libflux.Analyze(hdl)
	if err != nil {
		return nil, err
	}
	defer sem.Free()
	bs, err := sem.MarshalFB()
	if err != nil {
		return nil, err
	}
	return semantic.DeserializeFromFlatBuffer(bs)
}
