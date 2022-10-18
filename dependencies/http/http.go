package http

import (
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"syscall"
	"time"

	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/dependencies/url"
	"github.com/influxdata/flux/internal/errors"
)

// maxResponseBody is the maximum response body we will read before just discarding
// the rest. This allows sockets to be reused.
const maxResponseBody = 100 * 1024 * 1024 // 100 MB

type Client interface {
	Do(*http.Request) (*http.Response, error)
}

func LimitHTTPBody(client http.Client, size int64) *http.Client {
	// The client is already a struct so it was already copied
	// which makes this safe.
	if client.Transport == nil {
		client.Transport = http.DefaultTransport
	}
	client.Transport = roundTripLimiter{RoundTripper: client.Transport, size: size}
	return &client
}

type limitedReadCloser struct {
	io.Reader
	io.Closer
}

func limitReadCloser(rc io.ReadCloser, size int64) limitedReadCloser {
	return limitedReadCloser{
		Reader: io.LimitReader(rc, size),
		Closer: rc,
	}
}

type roundTripLimiter struct {
	http.RoundTripper
	size int64
}

func (l roundTripLimiter) RoundTrip(r *http.Request) (*http.Response, error) {
	response, err := l.RoundTripper.RoundTrip(r)
	if err != nil {
		return nil, err
	}
	// Check that the content length is not too large for us to handle.
	if response.ContentLength > l.size {
		return nil, errors.New(codes.FailedPrecondition, "http response body is too large, reduce the amount of data querying")
	}
	response.Body = limitReadCloser(response.Body, l.size)
	return response, nil
}

// NewDefaultClient creates a client with sane defaults.
func NewDefaultClient(urlValidator url.Validator) *http.Client {
	// Control is called after DNS lookup, but before the network connection is
	// initiated.
	control := func(network, address string, c syscall.RawConn) error {
		host, _, err := net.SplitHostPort(address)
		if err != nil {
			return err
		}

		ip := net.ParseIP(host)
		return urlValidator.ValidateIP(ip)
	}

	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		Control:   control,
		// DualStack is deprecated
	}

	// These defaults are copied from http.DefaultTransport.
	return &http.Client{
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			DialContext:           dialer.DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       10 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			// Fields below are NOT part of Go's defaults
			MaxIdleConnsPerHost: 100,
		},
	}
}

// NewLimitedDefaultClient creates a client with a limit on the response body size.
func NewLimitedDefaultClient(urlValidator url.Validator) *http.Client {
	cli := NewDefaultClient(urlValidator)
	return LimitHTTPBody(*cli, maxResponseBody)
}

func WithTimeout(c Client, t time.Duration) (Client, error) {
	cli, ok := c.(*http.Client)
	if !ok {
		return nil, errors.New(codes.Internal, "cannot set timeout on client")
	}
	// make shallow copy
	newClient := *cli
	newClient.Timeout = t
	return &newClient, nil
}
func WithTLSConfig(c Client, config *tls.Config) (Client, error) {
	cli, ok := c.(*http.Client)
	if !ok {
		return nil, errors.New(codes.Internal, "cannot set timeout on client")
	}
	// make shallow copy of client
	newClient := *cli

	// We control the clients so we can safely deconstruct the client
	// to change its transport config.
	switch t := newClient.Transport.(type) {
	case *http.Transport:
		newTransport := t.Clone()
		newTransport.TLSClientConfig = config
		newClient.Transport = newTransport
	case roundTripLimiter:
		transport, ok := t.RoundTripper.(*http.Transport)
		if !ok {
			return nil, errors.New(codes.Internal, "roundTripLimiter does not have http a known transport")
		}
		newTransport := transport.Clone()
		newTransport.TLSClientConfig = config
		t.RoundTripper = newTransport
		newClient.Transport = t
	default:
		return nil, errors.New(codes.Internal, "http client does not have http a known transport")
	}
	return &newClient, nil
}

// privateClient is an http client that obscures error messages that may contain
// sensitive information
type privateClient struct {
	client Client
}

func NewPrivateClient(c Client) Client {
	return &privateClient{client: c}
}

func (c *privateClient) Do(req *http.Request) (*http.Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, codes.Internal, "an internal error has occurred")
	}
	return resp, nil
}
