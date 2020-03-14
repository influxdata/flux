package runtime

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/libflux/go/libflux"
	"github.com/influxdata/flux/semantic"
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
