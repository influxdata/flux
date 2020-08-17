package url_test

import (
	nurl "net/url"
	"testing"

	"github.com/influxdata/flux/dependencies/url"
)

func TestPassValidator(t *testing.T) {
	if err := (url.PassValidator{}).Validate(nil); err != nil {
		t.Error(err)
	}
}
func TestPrivateIPValidator(t *testing.T) {
	v := url.PrivateIPValidator{}
	testCases := []struct {
		url   string
		valid bool
	}{
		{
			url:   "http://localhost",
			valid: false,
		},
		{
			url:   "http://127.0.0.1:1234",
			valid: false,
		},
		{
			url:   "http://example.com:80",
			valid: true,
		},
		{
			url:   "http://10.10.10.10",
			valid: false,
		},
		{
			url:   "http://192.168.0.0",
			valid: false,
		},
		{
			url:   "http://169.254.0.0",
			valid: false,
		},
		{
			url:   "http://1.1.1.1",
			valid: true,
		},
		{
			url:   "http://thisdnsnamedoesnotexistasitdoesnothavearootandhaslotsofentropy",
			valid: false,
		},
		{
			url:   "http://127.0.0.1:8093/debug/pprof",
			valid: false,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.url, func(t *testing.T) {
			u, err := nurl.Parse(tc.url)
			if err != nil {
				t.Fatal(err)
			}
			err = v.Validate(u)
			if tc.valid && err != nil || !tc.valid && err == nil {
				if tc.valid {
					t.Errorf("unexpected validation error: %v", err)
				} else {
					t.Errorf("expected validation error got nil")
				}
			}
		})
	}
}
