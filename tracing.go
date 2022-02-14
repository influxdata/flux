package flux

import (
	"context"
)

// WithQueryTracingEnabled will return a child context
// that will turn on experimental query tracing.
//
// Deprecated: Query tracing has been removed because generated traces
// would be too large and impractical. The query profiler is meant to be
// used instead.
func WithQueryTracingEnabled(parentCtx context.Context) context.Context {
	return parentCtx
}

// IsQueryTracingEnabled will return true if the context
// contains a key indicating that experimental tracing is enabled.
//
// Deprecated: See note about in WithQueryTracingEnabled.
func IsQueryTracingEnabled(ctx context.Context) bool {
	return false
}
