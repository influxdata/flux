package lang

import (
	"context"

	"github.com/influxdata/flux/memory"
	"go.uber.org/zap"
)

type dependencyKey int

const (
	executionDependenciesKey dependencyKey = iota
	compileOptionsKey
)

// ExecutionDependencies represents the dependencies that a function call
// executed by the Interpreter needs in order to trigger the execution of a flux.Program.
type ExecutionDependencies struct {
	Allocator *memory.Allocator
	Logger    *zap.Logger
}

func (d ExecutionDependencies) Inject(ctx context.Context) context.Context {
	return context.WithValue(ctx, executionDependenciesKey, d)
}

func GetExecutionDependencies(ctx context.Context) ExecutionDependencies {
	return ctx.Value(executionDependenciesKey).(ExecutionDependencies)
}

// CompileOptions represents multiple `lang.CompileOption` objects.
// It can be injected in the context and is used by the `flux.Compiler`s in package lang when compiling.
type CompileOptions []CompileOption

func (c CompileOptions) Inject(ctx context.Context) context.Context {
	return context.WithValue(ctx, compileOptionsKey, c)
}

func getCompileOptions(ctx context.Context) CompileOptions {
	return ctx.Value(compileOptionsKey).(CompileOptions)
}
