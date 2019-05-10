package interptest

import (
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/parser"
	"github.com/influxdata/flux/semantic"
)

func Eval(itrp *interpreter.Interpreter, scope interpreter.Scope, importer interpreter.Importer, src string) ([]interpreter.SideEffect, error) {
	pkg := parser.ParseSource(src)
	if ast.Check(pkg) > 0 {
		return nil, ast.GetError(pkg)
	}
	node, err := semantic.New(pkg)
	if err != nil {
		return nil, err
	}
	return itrp.Eval(node, scope, importer)
}
