package requests_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/InfluxCommunity/flux"
	fhttp "github.com/InfluxCommunity/flux/dependencies/http"
	"github.com/InfluxCommunity/flux/dependencies/url"
	_ "github.com/InfluxCommunity/flux/fluxinit/static"
	"github.com/InfluxCommunity/flux/runtime"
	"github.com/InfluxCommunity/flux/semantic"
	"github.com/InfluxCommunity/flux/values"
	"github.com/google/go-cmp/cmp"
)

func TestDo(t *testing.T) {
	var req *http.Request

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		req = request
		w.Header().Add("h1", "v1")
		w.Write([]byte("response: "))
		io.Copy(w, request.Body)
	}))
	defer ts.Close()

	script := fmt.Sprintf(`
import "http/requests"

resp = requests.do(
    method: "GET",
    url: "%s/path/a/b/c",
    params: ["p1":["l", "m"],"p2":["n"],"p3":["o"]],
    headers: ["x":"a","y":"b","z":"c"],
    body: bytes(v: "body"),
)
`, ts.URL)

	ctx := flux.NewDefaultDependencies().Inject(context.Background())
	if _, scope, err := runtime.Eval(ctx, script); err != nil {
		t.Fatal("evaluation of http.get failed: ", err)
	} else {
		respV, ok := scope.Lookup("resp")
		if !ok {
			t.Fatal("no resp in scope")
		}
		if respV.Type().Nature() != semantic.Object {
			t.Fatal("resp in not a record")
		}
		resp := respV.Object()
		if statusCode, ok := resp.Get("statusCode"); !ok {
			t.Error("no statusCode found in response")
		} else {
			if want, got := int64(200), statusCode.Int(); want != got {
				t.Errorf("unexpected status code want: %q got: %q", want, got)
			}
		}
		if body, ok := resp.Get("body"); !ok {
			t.Error("no body found in response")
		} else {
			if want, got := []byte("response: body"), body.Bytes(); !bytes.Equal(want, got) {
				t.Errorf("unexpected body want: %q got: %q", string(want), string(got))
			}
		}
		if headersV, ok := resp.Get("headers"); !ok {
			t.Error("no headers found in response")
		} else {
			headers := headersV.Dict()
			v := headers.Get(values.NewString("H1"), values.NewString(""))
			if want, got := "v1", v.Str(); want != got {
				t.Errorf("unexpected header H1 want: %q got: %q", want, got)
			}
		}
		if durationV, ok := resp.Get("duration"); !ok {
			t.Error("no duration found in response")
		} else {
			duration := durationV.Duration()
			got := duration.Duration()
			if got <= 0 {
				t.Errorf("unexpected duration want: > 0  got: %q", got)
			}
		}
	}
	if want, got := "/path/a/b/c", req.URL.Path; want != got {
		t.Errorf("unexpected url want: %q got: %q", want, got)
	}
	if want, got := "p1=l&p1=m&p2=n&p3=o", req.URL.RawQuery; want != got {
		t.Errorf("unexpected url query want: %q got: %q", want, got)
	}
	if want, got := "GET", req.Method; want != got {
		t.Errorf("unexpected method want: %q got: %q", want, got)
	}
	header := make(http.Header)
	header.Set("x", "a")
	header.Set("y", "b")
	header.Set("z", "c")
	header.Set("Accept-Encoding", "gzip")
	header.Set("User-Agent", "Go-http-client/1.1")
	header.Set("Content-Length", "4")
	if !cmp.Equal(header, req.Header) {
		t.Errorf("unexpected header -want/+got\n%s", cmp.Diff(header, req.Header))
	}
}

func TestDo_ValidationFail(t *testing.T) {
	script := `
import "http/requests"

requests.do(method: "GET", url:"http://127.1.1.1:8888/path/a/b/c", headers: ["x":"a","y":"b","z":"c"])
`

	deps := flux.NewDefaultDependencies()
	urlValidator := url.PrivateIPValidator{}
	deps.Deps.HTTPClient = fhttp.NewLimitedDefaultClient(urlValidator)
	deps.Deps.URLValidator = urlValidator
	ctx := deps.Inject(context.Background())
	_, _, err := runtime.Eval(ctx, script)
	if err == nil {
		t.Fatal("expected failure")
	}
	if !strings.Contains(err.Error(), "no such host") {
		t.Errorf("unexpected cause of failure, got err: %v", err)
	}
}

func TestDo_DNSFail(t *testing.T) {
	script := `
import "http/requests"

requests.do(method: "GET", url:"http://notarealaddressatall/path/a/b/c", headers: ["x":"a","y":"b","z":"c"])
`

	deps := flux.NewDefaultDependencies()
	deps.Deps.HTTPClient = http.DefaultClient
	deps.Deps.URLValidator = url.PrivateIPValidator{}
	ctx := deps.Inject(context.Background())
	_, _, err := runtime.Eval(ctx, script)
	if err == nil {
		t.Fatal("expected failure")
	}
	if !strings.Contains(err.Error(), "no such host") && !strings.Contains(err.Error(), "Temporary failure in name resolution") {
		t.Errorf("unexpected cause of failure, got err: %v", err)
	}
}

func TestDo_Timeout(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		// Sleep for 1s
		time.Sleep(time.Second)
		w.WriteHeader(204)
	}))
	defer ts.Close()

	script := fmt.Sprintf(`
import "http/requests"

// syntax doesn't allow for {http.DefaultConfig with ...} so we rebind it
// See https://github.com/InfluxCommunity/flux/issues/3655
c = requests.defaultConfig
config = {c with timeout: 10ms}
requests.do(method: "GET", url:"%s/path/a/b/c", config: config)
`, ts.URL)

	ctx := flux.NewDefaultDependencies().Inject(context.Background())
	_, _, err := runtime.Eval(ctx, script)
	if err == nil {
		t.Fatal("expected timeout failure")
	}
	if !strings.Contains(err.Error(), "Client.Timeout exceeded") && !strings.Contains(err.Error(), "context deadline exceeded") {
		t.Errorf("unexpected cause of failure, got err: %v", err)
	}
}
func TestDo_VerifyTLS_Pass(t *testing.T) {
	var req *http.Request

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		req = request
		w.WriteHeader(204)
	}))
	defer ts.Close()

	script := fmt.Sprintf(`
import "http/requests"

// syntax doesn't allow for {http.DefaultConfig with ...} so we rebind it
// See https://github.com/InfluxCommunity/flux/issues/3655
c = requests.defaultConfig
config = {c with insecureSkipVerify: true}
requests.do(method: "GET", url:"%s/path/a/b/c", headers: ["x":"a","y":"b","z":"c"], config: config)
`, ts.URL)

	ctx := flux.NewDefaultDependencies().Inject(context.Background())
	if _, _, err := runtime.Eval(ctx, script); err != nil {
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
	header.Set("User-Agent", "Go-http-client/1.1")
	if !cmp.Equal(header, req.Header) {
		t.Errorf("unexpected header -want/+got\n%s", cmp.Diff(header, req.Header))
	}

}
func TestDo_VerifyTLS_Fail(t *testing.T) {

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		w.WriteHeader(204)
	}))
	defer ts.Close()

	script := fmt.Sprintf(`
import "http/requests"

requests.do(method: "GET", url:"%s/path/a/b/c")
`, ts.URL)

	ctx := flux.NewDefaultDependencies().Inject(context.Background())
	_, _, err := runtime.Eval(ctx, script)
	if err == nil {
		t.Fatal("expected TLS failure")
	}
	if !strings.Contains(err.Error(), "unknown authority") && !strings.Contains(err.Error(), "certificate is not trusted") {
		t.Errorf("unexpected cause of failure, got err: %v", err)
	}
}
