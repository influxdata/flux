package experimental_test

import (
	"testing"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

func addFail(scope values.Scope) {
	scope.Set("fail", values.NewFunction(
		"fail",
		semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
			Return: semantic.Bool,
		}),
		func(args values.Object) (values.Value, error) {
			return nil, errors.New(codes.Aborted, "fail")
		},
		false,
	))
}

func TestObjectKeys(t *testing.T) {
	script := `
import "experimental"

o = {a: 1, b: 2, c: 3}
experimental.objectKeys(o: o) == ["a", "b", "c"] or fail()
`
	if _, _, err := flux.Eval(script, addFail); err != nil {
		t.Fatal("evaluation of objectKeys failed: ", err)
	}
}
