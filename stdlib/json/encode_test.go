package json_test

import (
	"context"
	"github.com/influxdata/flux/dependencies/dependenciestest"
	"testing"

	"github.com/influxdata/flux"
	_ "github.com/influxdata/flux/builtin"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/dependencies"
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
		func(ctx context.Context, deps dependencies.Interface, args values.Object) (values.Value, error) {
			return nil, errors.New(codes.Aborted, "fail")
		},
		false,
	))
}

func TestJSONEncode(t *testing.T) {
	script := `
import "json"

o = {
    a:1,
    b: {
        x: [1,2],
        y: "string",
        z: 1m
    },
    c: 1.1,
    d: false,
    e: /.*/,
	f: 2019-08-14T10:03:12Z,
}
json.encode(v: o) == bytes(v:"{\"a\":1,\"b\":{\"x\":[1,2],\"y\":\"string\",\"z\":60000000000},\"c\":1.1,\"d\":false,\"e\":\".*\",\"f\":\"2019-08-14T10:03:12Z\"}")  or fail()
`
	if _, _, err := flux.Eval(context.Background(), dependenciestest.NewTestDependenciesInterface(), script, addFail); err != nil {
		t.Fatal("evaluation of json.encode failed: ", err)
	}
}
