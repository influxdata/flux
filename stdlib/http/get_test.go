package http_test

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	_ "github.com/influxdata/flux/builtin"
	_ "github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/dependencies/url"
	_ "github.com/influxdata/flux/internal/errors"
	_ "github.com/influxdata/flux/semantic"
	_ "github.com/influxdata/flux/values"
)

func TestGet(t *testing.T) {
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

status = http.get(url:"%s/path/a/b/c", headers: {x:"a",y:"b",z:"c"})
status == 204 or fail()
`, ts.URL)

	ctx := flux.NewDefaultDependencies().Inject(context.Background())
	if _, _, err := flux.Eval(ctx, script, addFail); err != nil {
		t.Fatal("evaluation of http.get failed: ", err)
	}
	if want, got := "/path/a/b/c", req.URL.Path; want != got {
		t.Errorf("unexpected url want: %q got: %q", want, got)
	}
	if want, got := "GET", req.Method; want != got {
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

func TestGet_ValidationFail(t *testing.T) {
	script := `
import "http"

http.get(url:"http://127.1.1.1/path/a/b/c", headers: {x:"a",y:"b",z:"c"})
`

	deps := flux.NewDefaultDependencies()
	deps.Deps.HTTPClient = http.DefaultClient
	deps.Deps.URLValidator = url.PrivateIPValidator{}
	ctx := deps.Inject(context.Background())
	_, _, err := flux.Eval(ctx, script, addFail)
	if err == nil {
		t.Fatal("expected failure")
	}
	if !strings.Contains(err.Error(), "url is not valid") {
		t.Errorf("unexpected cause of failure, got err: %v", err)
	}
}
