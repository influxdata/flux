package flux

import (
	"context"
	"fmt"
	"time"

	"github.com/influxdata/flux/memory"

	"go.uber.org/zap"
)

// Compiler produces a specification for the query.
type Compiler interface {
	// Compile produces a specification for the query.
	Compile(ctx context.Context) (Program, error)
	CompilerType() CompilerType
}

// CompilerType is the name of a query compiler.
type CompilerType string
type CreateCompiler func() Compiler
type CompilerMappings map[CompilerType]CreateCompiler

func (m CompilerMappings) Add(t CompilerType, c CreateCompiler) error {
	if _, ok := m[t]; ok {
		return fmt.Errorf("duplicate compiler mapping for %q", t)
	}
	m[t] = c
	return nil
}

// ExecutionDependencies represents the provided dependencies to the execution environment.
// The dependencies is opaque.
type ExecutionDependencies map[string]interface{}

type ExecutionContext struct {
	Context      context.Context
	Now          time.Time
	Allocator    *memory.Allocator
	Dependencies ExecutionDependencies
	Logger       *zap.Logger
}

func NewDefaultExecutionContext() *ExecutionContext {
	return &ExecutionContext{
		Context:      context.Background(),
		Now:          time.Now(),
		Allocator:    new(memory.Allocator),
		Dependencies: make(map[string]interface{}),
		Logger:       zap.NewNop(),
	}
}

// Program defines a Flux script which has been compiled.
type Program interface {
	// Start begins execution of the program and returns immediately.
	// As results are produced they arrive on the channel.
	// The program is finished once the result channel is closed and all results have been consumed.
	Start(ec *ExecutionContext) (Query, error)
}
