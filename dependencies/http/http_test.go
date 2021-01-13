package http

import (
	"bytes"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/foxcpp/go-mockdns"
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

// Mocking DNS to resolve to blocked IP ranges and ensuring the connections
// fail. ADT: Would be nice to also test some pass cases, but it is unclear how
// to prevent the actual network connection out to the public internet. It's
// the dialer control function we are testing, so we can't mock the dialer to
// achieve that. Can really only replace the validator to pass a connetion to
// localhost. Doing that below the redirect test.
func TestIpValidation(t *testing.T) {
	bad := map[string]mockdns.Zone{
		"bad01.example.org.": {A: []string{"127.0.0.1"}},
		"bad02.example.org.": {A: []string{"127.10.20.30"}},
		"bad03.example.org.": {A: []string{"172.23.1.2"}},
		"bad04.example.org.": {A: []string{"192.168.0.0"}},
		"bad05.example.org.": {A: []string{"169.254.0.1"}},
		"bad06.example.org.": {AAAA: []string{"0000:0000:0000:0000:0000:0000:0000:0001"}},
		"bad07.example.org.": {AAAA: []string{"fc80:0000:0000:0000:0000:0000:0000:0001"}},
		"bad08.example.org.": {AAAA: []string{"fc00:0000:0000:0000:0000:0000:0000:0001"}},
	}

	// Mock a dns server.
	dnsLogger := log.New(ioutil.Discard, "mockdns server: ", log.LstdFlags)
	srv, _ := mockdns.NewServerWithLogger(bad, dnsLogger, false)
	defer srv.Close()

	// Hack the default resolver to use this server.
	srv.PatchNet(net.DefaultResolver)
	defer mockdns.UnpatchNet(net.DefaultResolver)

	client := NewDefaultClient(depsUrl.PrivateIPValidator{})

	for k := range bad {
		req, err := http.NewRequest("POST", ("http://" + k + "/path"), bytes.NewReader([]byte{}))
		if err != nil {
			t.Fatal(err)
		}

		_, err = client.Do(req)
		if err == nil {
			t.Fatal("expected private IP validation error, but client.Do succeeded")
		} else {
			if !strings.HasSuffix(err.Error(), "it connects to a private IP") {
				t.Fatalf("expected private IP validation error, but got %v", err)
			}
		}
	}
}

// The TestValidator will let everything through _except_ a specific IP, which
// is redirected to from an http test server. We must mock the validator
// because the original connection goes to localhost and that must be
// permitted, so it can send the redirect.
type TestValidator struct{}

func (TestValidator) Validate(anUrl *url.URL) error {
	return nil
}

func (TestValidator) ValidateIP(ip net.IP) error {
	if ip.Equal(net.ParseIP("127.6.6.6")) {
		return errors.New(codes.Invalid, "url validation error, it connects to a private IP")
	}
	return nil
}

func TestInvalidRedirects(t *testing.T) {
	bad := map[string]mockdns.Zone{
		"bad01.example.com.": {A: []string{"127.6.6.6"}},
	}

	// Mock a dns server.
	dnsLogger := log.New(ioutil.Discard, "mockdns server: ", log.LstdFlags)
	srv, _ := mockdns.NewServerWithLogger(bad, dnsLogger, false)
	defer srv.Close()

	// Hack the default resolver to use this server.
	srv.PatchNet(net.DefaultResolver)
	defer mockdns.UnpatchNet(net.DefaultResolver)

	t.Run("redirects to localhost are rejected", func(t *testing.T) {
		client := NewDefaultClient(TestValidator{})

		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
			http.Redirect(w, request, "http://bad01.example.com", http.StatusMovedPermanently)
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
