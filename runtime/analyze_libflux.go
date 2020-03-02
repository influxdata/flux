package runtime

import (
	"github.com/influxdata/flux/libflux/go/libflux"
	"github.com/influxdata/flux/semantic"
)

// AnalyzeSource parses and analyzes the given Flux source,
// using libflux.
func AnalyzeSource(fluxSrc string) (*semantic.Package, error) {
	ast := libflux.Parse(fluxSrc)
	return AnalyzePackage(ast)
}

func AnalyzePackage(astPkg *libflux.ASTPkg) (*semantic.Package, error) {
	defer astPkg.Free()
	sem, err := libflux.Analyze(astPkg)
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
