package http

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewDefaultClient(t *testing.T) {
	c := NewDefaultClient()
	if c == nil {
		t.Fail()
	}
}

func TestLimitedDefaultClient(t *testing.T) {
	t.Run("response larger than given size", func(t *testing.T) {
		var size int64 = 1
		c := LimitHTTPBody(*NewDefaultClient(), size)
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

func TestInvalidRedirects(t *testing.T) {
	t.Run("redirects to localhost are rejected", func(t *testing.T) {
		client := NewDefaultClient()

		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
			http.Redirect(w, request, "http://localhost", http.StatusMovedPermanently)
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
		if !strings.HasSuffix(err.Error(), "url is not valid, it connects to a private IP") {
			t.Fatal(err)
		}

	})
}
