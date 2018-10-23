// Package builtin contains all packages related to Flux built-ins are imported and initialized.
// This should only be imported from main or test packages.
// It is a mistake to import it from any other package.
package builtin

import (
	"github.com/influxdata/flux"
	_ "github.com/influxdata/flux/functions/inputs"
	_ "github.com/influxdata/flux/functions/outputs"
	_ "github.com/influxdata/flux/functions/transformations" // Import the built-in functions
	_ "github.com/influxdata/flux/options"                   // Import the built-in options
)

func init() {
	flux.FinalizeBuiltIns()
}
