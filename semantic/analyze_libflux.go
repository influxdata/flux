// +build libflux

package semantic

import (
	"github.com/influxdata/flux/libflux/go/libflux"
)

// AnalyzeSource parses and analyzes the given Flux source,
// using libflux.
func AnalyzeSource(fluxSrc string) (*Package, error) {
	ast := libflux.Parse(fluxSrc)
	defer ast.Free()
	sem, err := libflux.Analyze(ast)
	if err != nil {
		return nil, err
	}
	defer sem.Free()
	fb, err := sem.MarshalFB()
	if err != nil {
		return nil, err
	}
	return DeserializeFromFlatBuffer(fb)
}
