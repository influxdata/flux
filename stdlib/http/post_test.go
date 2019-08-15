package http_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	_ "github.com/influxdata/flux/builtin"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

func addFail(scope interpreter.Scope) {
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

func TestPost(t *testing.T) {
	var req *http.Request
	var body []byte
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req = r
		var err error
		body, err = ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
		}
		w.WriteHeader(204)
	}))
	defer ts.Close()

	script := fmt.Sprintf(`
import "http"

status = http.post(url:"%s/path/a/b/c", headers: {x:"a",y:"b",z:"c"}, data: bytes(v: "body"))
status == 204 or fail()
`, ts.URL)

	if _, _, err := flux.Eval(script, addFail); err != nil {
		t.Fatal("evaluation of http.post failed: ", err)
	}
	if want, got := "/path/a/b/c", req.URL.Path; want != got {
		t.Errorf("unexpected url want: %q got: %q", want, got)
	}
	if want, got := "POST", req.Method; want != got {
		t.Errorf("unexpected method want: %q got: %q", want, got)
	}
	header := make(http.Header)
	header.Set("x", "a")
	header.Set("y", "b")
	header.Set("z", "c")
	header.Set("Accept-Encoding", "gzip")
	header.Set("Content-Length", "4")
	header.Set("User-Agent", "Go-http-client/1.1")
	if !cmp.Equal(header, req.Header) {
		t.Errorf("unexpected header -want/+got\n%s", cmp.Diff(header, req.Header))
	}

	expBody := []byte("body")
	if !bytes.Equal(body, expBody) {
		t.Errorf("unexpected body want: %q got: %q", string(expBody), string(body))
	}
}
