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
	"github.com/influxdata/flux/codes"
	fluxhttp "github.com/influxdata/flux/dependencies/http"
	fluxurl "github.com/influxdata/flux/dependencies/url"
	fluxerrors "github.com/influxdata/flux/internal/errors"
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
			return nil, fluxerrors.New(codes.Internal, "missing dependencies")
		}
		if client, err = deps.HTTPClient(); err != nil {
			return nil, fluxerrors.New(codes.Internal, "missing dependencies")
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
		ok      bool
	)
	{
		if method, err = args.GetRequiredString("method"); err != nil {
			return nil, err
		}

		if path, err = args.GetRequiredString("path"); err != nil {
			return nil, err
		}

		if host, ok, err = args.GetString("host"); err != nil {
			return nil, err
		} else if !ok {
			return nil, fluxerrors.New(codes.Invalid, `keyword argument "host" is required when executing outside InfluxDB`)
		}

		if token, ok, err = args.GetString("token"); err != nil {
			return nil, err
		} else if !ok {
			return nil, fluxerrors.New(codes.Invalid, `keyword argument "token" is required when executing outside InfluxDB`)
		}

		if raw, ok := args.Get("timeout"); !ok {
			timeout = values.ConvertDurationNsecs(30 * time.Second)
		} else if raw.Type().Nature() != semantic.Duration {
			return nil, fluxerrors.New(codes.Invalid, `keyword argument "timeout" must be a duration type`)
		} else {
			timeout = raw.Duration()
		}

		if query, _, err = args.GetDictionary("query"); err != nil {
			return nil, err
		}

		if headers, _, err = args.GetDictionary("headers"); err != nil {
			return nil, err
		}

		if b, ok := args.Get("body"); ok {
			if b.Type().Nature() != semantic.Bytes {
				return nil, fluxerrors.New(codes.Invalid, `keyword argument "body" must be a bytes type`)
			}
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

	responseHeaders, err := headerToDict(resp.Header)
	if err != nil {
		return nil, err
	}

	return values.NewObjectWithValues(map[string]values.Value{
		"statusCode": values.NewInt(int64(resp.StatusCode)),
		"headers":    responseHeaders,
		"body":       values.NewBytes(b)}), nil
}

// headerToDict constructs a values.Dictionary from a map of header keys and values.
func headerToDict(header http.Header) (values.Dictionary, error) {
	builder := values.NewDictBuilder(semantic.NewDictType(semantic.BasicString, semantic.BasicString))
	for name, thevalues := range header {
		for _, onevalue := range thevalues {
			if err := builder.Insert(values.NewString(name), values.NewString(onevalue)); err != nil {
				return nil, err
			}
		}
	}
	return builder.Dict(), nil
}
