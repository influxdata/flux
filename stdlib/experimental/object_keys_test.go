package experimental_test

import (
	"context"
	"testing"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/dependencies/dependenciestest"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

func addFail(scope values.Scope) {
	scope.Set("fail", values.NewFunction(
		"fail",
		semantic.NewFunctionType(semantic.BasicBool, nil),
		func(ctx context.Context, args values.Object) (values.Value, error) {
			return nil, errors.New(codes.Aborted, "fail")
		},
		false,
	))
}

func TestObjectKeys(t *testing.T) {
	t.Skip("https://github.com/influxdata/flux/issues/2402")
	script := `
import "experimental"

o = {a: 1, b: 2, c: 3}
experimental.objectKeys(o: o) == ["a", "b", "c"] or fail()
`
	ctx := dependenciestest.Default().Inject(context.Background())
	if _, _, err := flux.Eval(ctx, script, addFail); err != nil {
		t.Fatal("evaluation of objectKeys failed: ", err)
	}
}
