package http

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/codes"
	depsUrl "github.com/influxdata/flux/dependencies/url"
	"github.com/influxdata/flux/internal/errors"
)

func TestNewDefaultClient(t *testing.T) {
	c := NewDefaultClient(depsUrl.PassValidator{})
	if c == nil {
		t.Fail()
	}
}

func TestLimitedDefaultClient(t *testing.T) {
	t.Run("response larger than given size", func(t *testing.T) {
		var size int64 = 1
		c := LimitHTTPBody(*NewDefaultClient(depsUrl.PassValidator{}), size)
		body := "hello"

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
			_, err := w.Write([]byte(body))
			if err != nil {
				t.Fatalf("error in test server: %v", err)
			}
		}))
		defer ts.Close()

		req, err := http.NewRequest("GET", ts.URL, bytes.NewReader([]byte{}))
		if err != nil {
			t.Fatal(err)
		}
		resp, err := c.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		bs, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(body[:size], string(bs)); diff != "" {
			t.Fatalf("got unexpected response body:\n\t%s", diff)
		}
	})
}

// The TestValidator will let everything through _except_ a specific url, which
// is redirected to from a http test server.
type TestValidator struct{}

func (TestValidator) Validate(anUrl *url.URL) error {
	if anUrl.Host == "test-validator.example.com" {
		return errors.New(codes.Invalid, "url validation error, it connects to a private IP")
	}
	return nil
}

func TestInvalidRedirects(t *testing.T) {
	t.Run("redirects to localhost are rejected", func(t *testing.T) {
		client := NewDefaultClient(TestValidator{})

		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
			http.Redirect(w, request, "http://test-validator.example.com", http.StatusMovedPermanently)
		}))
		defer testServer.Close()

		req, err := http.NewRequest("GET", testServer.URL, bytes.NewReader([]byte{}))
		if err != nil {
			t.Fatal(err)
		}
		_, err = client.Do(req)
		if err == nil {
			t.Fatal("Client did not error")
		}
		if !strings.HasSuffix(err.Error(), "url validation error, it connects to a private IP") {
			t.Fatal(err)
		}

	})
}
