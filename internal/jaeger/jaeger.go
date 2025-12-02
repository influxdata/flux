package jaeger

import (
	"go.opentelemetry.io/otel/trace"
)

// InfoFromSpan returns the traceID and if it was sampled from the span.
// It returns whether a span associated with the context has been found.
func InfoFromSpan(span trace.Span) (traceID string, sampled bool, found bool) {
	sc := span.SpanContext()
	if sc.IsValid() {
		traceID = sc.TraceID().String()
		sampled = sc.IsSampled()
		found = true
	}
	return
}
