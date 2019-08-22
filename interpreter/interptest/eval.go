package interptest

import (
	"context"

	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/dependencies"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/parser"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

func Eval(ctx context.Context, deps dependencies.Interface, itrp *interpreter.Interpreter, scope values.Scope, importer interpreter.Importer, src string) ([]interpreter.SideEffect, error) {
	pkg := parser.ParseSource(src)
	if ast.Check(pkg) > 0 {
		return nil, ast.GetError(pkg)
	}
	node, err := semantic.New(pkg)
	if err != nil {
		return nil, err
	}
	return itrp.Eval(ctx, deps, node, scope, importer)
}
