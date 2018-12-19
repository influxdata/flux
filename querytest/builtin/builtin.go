// Package querytest.builtin contains flux builtins and testing-specific function initialization.
// This should be imported only by tests that require end-to-end query capabilities.
package builtin

import (
	"github.com/influxdata/flux"
	_ "github.com/influxdata/flux/functions/inputs"          // Import the built-in inputs
	_ "github.com/influxdata/flux/functions/outputs"         // Import the built-in outputs
	_ "github.com/influxdata/flux/functions/tests"           // Import the built-in tests
	_ "github.com/influxdata/flux/functions/transformations" // Import the built-in functions
	_ "github.com/influxdata/flux/options"                   // Import the built-in options
	_ "github.com/influxdata/flux/querytest/functions"       // Import the built-in test functions
)

func init() {
	flux.FinalizeBuiltIns()
}
