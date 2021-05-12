package jaeger

import (
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
)

// InfoFromSpan returns the traceID and if it was sampled from the span, given
// it is a jaeger span. It returns whether a span associated with the context
// has been found.
func InfoFromSpan(span opentracing.Span) (traceID string, sampled bool, found bool) {
	type ctxWithInfo interface {
		TraceID() jaeger.TraceID
		IsSampled() bool
	}
	if ctx, ok := span.Context().(ctxWithInfo); ok {
		traceID = ctx.TraceID().String()
		sampled = ctx.IsSampled()
		found = true
	}
	return
}
