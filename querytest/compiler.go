package querytest

import (
	"context"

	"github.com/influxdata/flux"
)

// FromCSVCompiler wraps a compiler and replaces all From operations with FromCSV
type FromCSVCompiler struct {
	flux.Compiler
	InputFile string
}

// FromJSONCompiler wraps a compiler and replaces all From operations with FromJSON
type FromJSONCompiler struct {
	flux.Compiler
	InputFile string
}

func (c FromCSVCompiler) Compile(ctx context.Context) (*flux.Spec, error) {
	spec, err := c.Compiler.Compile(ctx)
	if err != nil {
		return nil, err
	}
	ReplaceFromSpec(spec, c.InputFile)
	return spec, nil
}

func (c FromJSONCompiler) Compile(ctx context.Context) (*flux.Spec, error) {
	spec, err := c.Compiler.Compile(ctx)
	if err != nil {
		return nil, err
	}
	ReplaceFromWithFromJSONSpec(spec, c.InputFile)
	return spec, nil
}
