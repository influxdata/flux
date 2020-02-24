package json_test

import (
	"context"
	"testing"

	_ "github.com/influxdata/flux/builtin"
	"github.com/influxdata/flux/dependencies/dependenciestest"
	"github.com/influxdata/flux/runtime"
)

func TestJSONEncode(t *testing.T) {
	script := `
import "json"
import "internal/testutil"

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
json.encode(v: o) == bytes(v:"{\"a\":1,\"b\":{\"x\":[1,2],\"y\":\"string\",\"z\":\"1m\"},\"c\":1.1,\"d\":false,\"e\":\".*\",\"f\":\"2019-08-14T10:03:12Z\"}")  or testutil.fail()
`
	ctx := dependenciestest.Default().Inject(context.Background())
	if _, _, err := runtime.Eval(ctx, script); err != nil {
		t.Fatal("evaluation of json.encode failed: ", err)
	}
}
