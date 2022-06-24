package interptest

import (
	"context"

	"github.com/mvn-trinhnguyen2-dn/flux/codes"
	"github.com/mvn-trinhnguyen2-dn/flux/internal/errors"
	"github.com/mvn-trinhnguyen2-dn/flux/interpreter"
	"github.com/mvn-trinhnguyen2-dn/flux/runtime"
	"github.com/mvn-trinhnguyen2-dn/flux/values"
)

func Eval(ctx context.Context, itrp *interpreter.Interpreter, scope values.Scope, importer interpreter.Importer, src string) ([]interpreter.SideEffect, error) {
	node, err := runtime.AnalyzeSource(src)
	if err != nil {
		return nil, errors.Wrap(err, codes.Inherit, "could not analyze program")
	}
	return itrp.Eval(ctx, node, scope, importer)
}
