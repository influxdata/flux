//go:build !fipsonly

package sql

import (
	"context"
	"errors"
	"net"
	"net/http"
	"testing"

	"github.com/influxdata/flux"
	fhttp "github.com/influxdata/flux/dependencies/http"
)

// snowflakeDeps overrides both HTTPClient and Dialer for snowflake tests.
type snowflakeDeps struct {
	flux.Deps
	httpClient fhttp.Client
	dialer     flux.Dialer
}

func (d snowflakeDeps) HTTPClient() (fhttp.Client, error) {
	return d.httpClient, nil
}

func (d snowflakeDeps) Dialer() (flux.Dialer, error) {
	return d.dialer, nil
}

// Test the path where deps.HTTPClient() returns an *http.Client,
// so its Transport is copied to the snowflake config.
func TestSnowflakeOpenFunctionDialer_HTTPClient(t *testing.T) {
	expectErr := errors.New("test dial error via http client")
	dialf := func(_ context.Context, _, _ string) (net.Conn, error) {
		return nil, expectErr
	}

	deps := snowflakeDeps{
		Deps: flux.NewDefaultDependencies(),
		httpClient: &http.Client{
			Transport: fhttp.NewTransport(dialf),
		},
	}

	openFn := snowflakeOpenFunction("user:password@accountname/dbname?loginTimeout=1")
	db, err := openFn(deps)
	if err != nil {
		t.Fatalf("unexpected error from open function: %v", err)
	}
	defer db.Close()

	// Ping triggers a real connection attempt through the transport.
	err = db.Ping()
	if err == nil {
		t.Fatal("expected error from Ping, got nil")
	}
	if !errors.Is(err, expectErr) {
		t.Fatalf("expected error %q, got: %v", expectErr, err)
	}
}

// nonStdHTTPClient implements fhttp.Client but is NOT *http.Client,
// so snowflakeOpenFunction falls through to deps.Dialer().
type nonStdHTTPClient struct{}

func (nonStdHTTPClient) Do(_ *http.Request) (*http.Response, error) {
	return nil, errors.New("should not be called")
}

// Test the path where deps.HTTPClient() returns a non-*http.Client,
// so a new Transport is created from deps.Dialer().
func TestSnowflakeOpenFunctionDialer_Dialer(t *testing.T) {
	expectErr := errors.New("test dial error via dialer")

	deps := snowflakeDeps{
		Deps:       flux.NewDefaultDependencies(),
		httpClient: nonStdHTTPClient{},
		dialer:     &mockDialer{err: expectErr},
	}

	openFn := snowflakeOpenFunction("user:password@accountname/dbname?loginTimeout=1")
	db, err := openFn(deps)
	if err != nil {
		t.Fatalf("unexpected error from open function: %v", err)
	}
	defer db.Close()

	// Ping triggers a real connection attempt through the dialer.
	err = db.Ping()
	if err == nil {
		t.Fatal("expected error from Ping, got nil")
	}
	if !errors.Is(err, expectErr) {
		t.Fatalf("expected error %q, got: %v", expectErr, err)
	}
}
