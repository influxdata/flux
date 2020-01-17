package json_test

import (
	"context"
	"testing"

	"github.com/influxdata/flux"
	_ "github.com/influxdata/flux/builtin"
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
json.encode(v: o) == bytes(v:"{\"a\":1,\"b\":{\"x\":[1,2],\"y\":\"string\",\"z\":\"1m\"},\"c\":1.1,\"d\":false,\"e\":\".*\",\"f\":\"2019-08-14T10:03:12Z\"}")  or fail()
`
	ctx := dependenciestest.Default().Inject(context.Background())
	if _, _, err := flux.Eval(ctx, script, addFail); err != nil {
		t.Fatal("evaluation of json.encode failed: ", err)
	}
}
