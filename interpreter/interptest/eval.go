package interptest

import (
	"context"

	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

func Eval(ctx context.Context, itrp *interpreter.Interpreter, scope values.Scope, importer interpreter.Importer, src string) ([]interpreter.SideEffect, error) {
	node, err := semantic.AnalyzeSource(src)
	if err != nil {
		return nil, errors.Wrap(err, codes.Inherit, "could not analyze program")
	}
	return itrp.Eval(ctx, node, scope, importer)
}
