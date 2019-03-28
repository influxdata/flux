package mock

import (
	"context"

	"github.com/influxdata/flux"
)

type Compiler struct {
	CompileFn func(ctx context.Context) (flux.Program, error)
	Type      flux.CompilerType
}

func (c Compiler) Compile(ctx context.Context) (flux.Program, error) {
	return c.CompileFn(ctx)
}
func (c Compiler) CompilerType() flux.CompilerType {
	if c.Type == "" {
		return "mockCompiler"
	}
	return c.Type
}
