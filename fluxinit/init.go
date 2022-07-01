// Package fluxinit is used to initialize the flux library for compilation and
// execution of Flux. The FluxInit function should be called exactly once in a
// process.
package fluxinit

import (
	"github.com/influxdata/flux/runtime"
	_ "github.com/influxdata/flux/stdlib"
)

// The FluxInit() function prepares the runtime for compilation and execution
// of Flux. This is a costly step and should only be performed if the intention
// is to compile and execute flux code.
//
// This package imports the standard library. These modules register themselves
// in go init() functions. This package must ensure all required standard
// library functions are imported.
//
// As a convenience, the fluxinit/static package can be imported for use cases
// where static initialization is okay, such as tests.

func FluxInit() {
	runtime.FinalizeBuiltIns()
}
