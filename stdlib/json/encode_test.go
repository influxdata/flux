package json_test

import (
	"context"
	"testing"

	"github.com/influxdata/flux/dependencies/dependenciestest"
	_ "github.com/influxdata/flux/fluxinit/static"
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
    g: [1: "hi", 2: "there"]
}
json.encode(v: o) == bytes(v:"{\"a\":1,\"b\":{\"x\":[1,2],\"y\":\"string\",\"z\":\"1m\"},\"c\":1.1,\"d\":false,\"e\":\".*\",\"f\":\"2019-08-14T10:03:12Z\",\"g\":{\"1\":\"hi\",\"2\":\"there\"}}") or testutil.fail()
`
	ctx := dependenciestest.Default().Inject(context.Background())
	if _, _, err := runtime.Eval(ctx, script); err != nil {
		t.Fatal("evaluation of json.encode failed: ", err)
	}
}

func TestJSONEncode_ReceiveTableObjectIsError(t *testing.T) {
	src := `import "array"
			import "json"
			json.encode(v: array.from(rows: [{}]))`
	_, _, err := runtime.Eval(context.Background(), src)
	if err == nil {
		t.Fatal("expected error, got none")
	}

	if want, got := "error calling function \"encode\" @3:4-3:42: got table stream instead of array. Try using tableFind() or findRecord() to extract data from stream", err.Error(); want != got {
		t.Errorf("wanted error %q, got %q", want, got)
	}
}
