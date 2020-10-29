package execute

import (
	"context"
	"time"

	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/metadata"
	"go.uber.org/zap"
)

type key int

const executionDependenciesKey key = iota

type ExecutionOptions struct {
	OperatorProfiler *OperatorProfiler
	Profilers        []Profiler
}

// ExecutionDependencies represents the dependencies that a function call
// executed by the Interpreter needs in order to trigger the execution of a flux.Program
type ExecutionDependencies struct {
	// Must be set
	Allocator *memory.Allocator
	Now       *time.Time

	// Allowed to be nil
	Logger *zap.Logger

	// Metadata is passed up from any invocations of execution up to the parent
	// execution, and out through the statistics.
	Metadata metadata.Metadata

	ExecutionOptions *ExecutionOptions
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

// Create some execution dependencies. Any arg may be nil, this will choose
// some suitable defaults.
func NewExecutionDependencies(allocator *memory.Allocator, now *time.Time, logger *zap.Logger) ExecutionDependencies {
	if allocator == nil {
		allocator = new(memory.Allocator)
	}

	if now == nil {
		nowVar := time.Now()
		now = &nowVar
	}
	return ExecutionDependencies{
		Allocator:        allocator,
		Now:              now,
		Logger:           logger,
		Metadata:         make(metadata.Metadata),
		ExecutionOptions: &ExecutionOptions{},
	}
}

func DefaultExecutionDependencies() ExecutionDependencies {
	return NewExecutionDependencies(nil, nil, nil)
}
