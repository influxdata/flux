// +build libflux

package semantic

import (
	"github.com/influxdata/flux/libflux/go/libflux"
)

// AnalyzeSource parses and analyzes the given Flux source,
// using libflux.
func AnalyzeSource(fluxSrc string) (*Package, error) {
	data, err := libflux.Analyze(fluxSrc)
	if err != nil {
		return nil, err
	}
	return DeserializeFromFlatBuffer(data)
}
