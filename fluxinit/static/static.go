// The fluxinit/static package can be imported in test cases and other uses
// cases where it is okay to always initialize flux.
package static

import (
	"github.com/InfluxCommunity/flux/fluxinit"
)

func init() {
	fluxinit.FluxInit()
}
