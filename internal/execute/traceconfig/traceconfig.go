package traceconfig

var experimentalTracingEnabled = false

func EnableExperimentalTracing() {
	experimentalTracingEnabled = true
}

func IsExperimentalTracingEnabled() bool {
	return experimentalTracingEnabled
}
