package querytest

import (
	"context"

	"github.com/influxdata/flux"
)

// ReplaceOpInSpecFunction is a function that takes an Operation and returns a new OperationSpec for substitution,
// or nil, if nothing has to be changed.
type ReplaceOpInSpecFunction func(op *flux.Operation) flux.OperationSpec

type ReplaceSpecCompiler struct {
	flux.Compiler
	fn ReplaceOpInSpecFunction
}

func NewReplaceSpecCompiler(compiler flux.Compiler, fn ReplaceOpInSpecFunction) *ReplaceSpecCompiler {
	return &ReplaceSpecCompiler{Compiler: compiler, fn: fn}
}

func (c *ReplaceSpecCompiler) Compile(ctx context.Context) (*flux.Spec, error) {
	spec, err := c.Compiler.Compile(ctx)
	if err != nil {
		return nil, err
	}
	ReplaceOpInSpec(spec, c.fn)
	return spec, nil
}

func ReplaceOpInSpec(q *flux.Spec, replaceFn ReplaceOpInSpecFunction) {
	for _, op := range q.Operations {
		newOpSpec := replaceFn(op)
		if newOpSpec != nil {
			op.Spec = newOpSpec
		}
	}
}
