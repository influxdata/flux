package lang

import (
	"context"

	"github.com/influxdata/flux/memory"

	"go.uber.org/zap"
)

type key int

const executionDependenciesKey key = iota

// ExecutionDependencies represents the dependencies that a function call
// executed by the Interpreter needs in order to trigger the execution of a flux.Program
type ExecutionDependencies struct {
	Allocator *memory.Allocator
	Logger    *zap.Logger
}

func (d ExecutionDependencies) Inject(ctx context.Context) context.Context {
	return context.WithValue(ctx, executionDependenciesKey, d)
}

func HaveExecutionDependencies(ctx context.Context) bool {
	return ctx.Value(executionDependenciesKey) != nil
}

func GetExecutionDependencies(ctx context.Context) ExecutionDependencies {
	return ctx.Value(executionDependenciesKey).(ExecutionDependencies)
}
