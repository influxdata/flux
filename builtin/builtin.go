// Package builtin contains all packages related to Flux built-ins are imported and initialized.
// This should only be imported from main or test packages.
// It is a mistake to import it from any other package.
package builtin

import (
	"github.com/influxdata/flux"
	_ "github.com/influxdata/flux/options" // Import the built-in options
	_ "github.com/influxdata/flux/stdlib"  // Import the stdlib

	// TODO(nathanielc): Remove this line once the tests are full fledged package built-ins
	_ "github.com/influxdata/flux/stdlib/tests" // Import the built-in functions
)

func init() {
	flux.FinalizeBuiltIns()
}
