package interptest

import (
	"context"

	"github.com/InfluxCommunity/flux/codes"
	"github.com/InfluxCommunity/flux/internal/errors"
	"github.com/InfluxCommunity/flux/interpreter"
	"github.com/InfluxCommunity/flux/runtime"
	"github.com/InfluxCommunity/flux/values"
)

func Eval(ctx context.Context, itrp *interpreter.Interpreter, scope values.Scope, importer interpreter.Importer, src string) ([]interpreter.SideEffect, error) {
	node, err := runtime.AnalyzeSource(ctx, src)
	if err != nil {
		return nil, errors.Wrap(err, codes.Inherit, "could not analyze program")
	}
	return itrp.Eval(ctx, node, scope, importer)
}
