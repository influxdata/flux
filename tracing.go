package flux

import (
	"context"
)

const queryTracingContextKey = "query-tracing-enabled"

// WithExperimentalTracingEnabled will return a child context
// that will turn on experimental query tracing.
func WithExperimentalTracingEnabled(parentCtx context.Context) context.Context {
	return context.WithValue(parentCtx, queryTracingContextKey, true)
}

// IsExperimentalTracingEnabled will return true if the context
// contains a key indicating that experimental tracing is enabled.
func IsExperimentalTracingEnabled(ctx context.Context) bool {
	v := ctx.Value(queryTracingContextKey)
	if v == nil {
		return false
	}
	b, ok := v.(bool)
	if !ok {
		return false
	}
	return b
}
