package http

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
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
