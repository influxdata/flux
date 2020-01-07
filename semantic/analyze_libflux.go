// +build libflux

package semantic

import (
	"github.com/influxdata/flux/libflux/go/libflux"
)

// AnalyzeSource parses and analyzes the given Flux source,
// using libflux.
func AnalyzeSource(fluxSrc string) (*Package, error) {
	ast := libflux.Parse(fluxSrc)
	sem, err := libflux.Analyze(ast)
	fb, err := sem.MarshalFB()

	if err != nil {
		return nil, err
	}
	return DeserializeFromFlatBuffer(fb)
}
