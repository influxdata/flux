package semantic

import (
	"github.com/influxdata/flux/libflux/go/libflux"
)

// AnalyzeSource parses and analyzes the given Flux source,
// using libflux.
func AnalyzeSource(fluxSrc string) (*Package, error) {
	ast := libflux.Parse(fluxSrc)
	return AnalyzePackage(ast)
}

func AnalyzePackage(astPkg *libflux.ASTPkg) (*Package, error) {
	defer astPkg.Free()
	sem, err := libflux.Analyze(astPkg)
	if err != nil {
		return nil, err
	}
	defer sem.Free()
	mbuf, err := sem.MarshalFB()
	if err != nil {
		return nil, err
	}
	return DeserializeFromFlatBuffer(mbuf.Buffer, mbuf.Offset)
}
