package feature

type Metrics interface {
	Inc(key string, value interface{})
}

type discardMetrics struct{}

func (discardMetrics) Inc(key string, value interface{}) {}

var metrics Metrics = discardMetrics{}

// SetMetrics sets the metric store for feature flags.
func SetMetrics(m Metrics) {
	if m == nil {
		metrics = discardMetrics{}
	} else {
		metrics = m
	}
}
