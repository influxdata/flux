package flux

var experimentalTracingEnabled = false

// EnableExperimentalTracing will enable any experimental
// tracing in the flux binary. Experimental tracing may provide
// more insight, but it indicates that we have not tested that the
// tracing doesn't have negative side effects when run in production.
//
// Traces that are enabled this way may be removed or may be enabled
// by default in the future.
func EnableExperimentalTracing() {
	experimentalTracingEnabled = true
}

func IsExperimentalTracingEnabled() bool {
	return experimentalTracingEnabled
}
