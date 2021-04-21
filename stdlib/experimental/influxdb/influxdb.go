package influxdb

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	neturl "net/url"
	"strings"
	"time"

	"github.com/influxdata/flux"
	fluxhttp "github.com/influxdata/flux/dependencies/http"
	fluxurl "github.com/influxdata/flux/dependencies/url"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

const (
	APIFuncName = "api"
	PackagePath = "experimental/influxdb"
	APIKind     = PackagePath + "." + APIFuncName
)

var APISignature = runtime.MustLookupBuiltinType(PackagePath, APIFuncName)

func init() {
	runtime.RegisterPackageValue(PackagePath, APIFuncName,
		values.NewFunction(APIFuncName, APISignature, api, true),
	)
}

// api submits an HTTP request to the specified API path.
// Returns HTTP status code, response headers, and body as a byte array.
func api(ctx context.Context, a values.Object) (values.Value, error) {
	var (
		deps      = flux.GetDependencies(ctx)
		validator fluxurl.Validator
		client    fluxhttp.Client
		err       error
	)
	{
		if validator, err = deps.URLValidator(); err != nil {
			return nil, errors.New("internal error")
		}
		if client, err = deps.HTTPClient(); err != nil {
			return nil, errors.New("internal error")
		}
	}

	var (
		args    = interpreter.NewArguments(a)
		method  string
		path    string
		host    string
		token   string
		headers values.Dictionary
		query   values.Dictionary
		timeout values.Duration
		body    []byte
	)
	{
		if method, err = args.GetRequiredString("method"); err != nil {
			return nil, err
		}

		if path, err = args.GetRequiredString("path"); err != nil {
			return nil, err
		}

		if host, err = args.GetRequiredString("host"); err != nil {
			return nil, err
		}

		if token, err = args.GetRequiredString("token"); err != nil {
			return nil, err
		}

		if raw, ok := args.Get("timeout"); !ok {
			timeout = values.ConvertDurationNsecs(30 * time.Second)
		} else if raw.Type().Nature() != semantic.Duration {
			return nil, errors.New("timeout argument must be a duration")
		} else {
			timeout = raw.Duration()
		}

		if q, ok := args.Get("query"); ok {
			query = q.Dict()
		}

		if h, ok := args.Get("headers"); ok {
			headers = h.Dict()
		}

		if b, ok := args.Get("body"); ok {
			body = b.Bytes()
		}
	}

	var url string
	{
		u, err := neturl.Parse(host + path)
		if err != nil {
			return nil, err
		}

		if query != nil {
			q := make(neturl.Values, query.Len())
			query.Range(func(k values.Value, v values.Value) {
				q.Set(k.Str(), v.Str())
			})
			u.RawQuery = q.Encode()
		}

		if err := validator.Validate(u); err != nil {
			return nil, err
		}

		url = u.String()
	}

	var req *http.Request
	{
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout.Duration())
		defer cancel()

		req, err = http.NewRequestWithContext(ctx, strings.ToUpper(method), url, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}

		if headers != nil {
			headers.Range(func(k values.Value, v values.Value) {
				req.Header.Set(k.Str(), v.Str())
			})
		}
		req.Header.Set("Authorization", "Token "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		// Following the pattern set by the  experimental/http.get implementation:
		// Alias the DNS lookup error so as not to disclose the DNS server address.
		// This error is private in the net/http package, so string matching is used.
		if strings.HasSuffix(err.Error(), "no such host") {
			return nil, errors.New("no such host")
		}
		return nil, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return values.NewObjectWithValues(map[string]values.Value{
		"statusCode": values.NewInt(int64(resp.StatusCode)),
		"headers":    headerToObject(resp.Header),
		"body":       values.NewBytes(b)}), nil
}

// headerToObject constructs a values.Object from a map of header keys and values.
func headerToObject(header http.Header) (headerObj values.Object) {
	m := make(map[string]values.Value)
	for name, thevalues := range header {
		for _, onevalue := range thevalues {
			m[name] = values.New(onevalue)
		}
	}
	return values.NewObjectWithValues(m)
}
