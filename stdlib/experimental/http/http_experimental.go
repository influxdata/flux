package http

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

// http get mirrors the http post originally completed for alerts & notifications
var get = values.NewFunction(
	"get",
	runtime.MustLookupBuiltinType("experimental/http", "get"),
	func(ctx context.Context, args values.Object) (values.Value, error) {
		// Get and validate URL
		uV, ok := args.Get("url")
		if !ok {
			return nil, errors.New(codes.Invalid, "missing \"url\" parameter")
		}
		u, err := url.Parse(uV.Str())
		if err != nil {
			return nil, err
		}
		deps := flux.GetDependencies(ctx)
		validator, err := deps.URLValidator()
		if err != nil {
			return nil, err
		}
		if err := validator.Validate(u); err != nil {
			return nil, errors.New(codes.Invalid, "no such host")
		}

		// http.NewDefaultClient() does default to 30
		var theTimeout = values.ConvertDurationNsecs(30 * time.Second)
		tv, ok := args.Get("timeout")
		if !ok {
			// default timeout
		} else if tv.Type().Nature() != semantic.Duration {
			return nil, fmt.Errorf("expected argument %q to be of type %v, got type %v", tv, semantic.Int, tv.Type().Nature())
		} else {
			theTimeout = tv.Duration()
		}

		// Construct HTTP request
		req, err := http.NewRequest("GET", uV.Str(), nil)
		if err != nil {
			return nil, err
		}

		// Add headers to request
		header, ok := args.Get("headers")
		if ok && !header.IsNull() {
			var rangeErr error
			header.Object().Range(func(k string, v values.Value) {
				if v.Type().Nature() == semantic.String {
					req.Header.Set(k, v.Str())
				} else {
					rangeErr = errors.Newf(codes.Invalid, "header value %q must be a string", k)
				}
			})
			if rangeErr != nil {
				return nil, rangeErr
			}
		}

		// Perform request
		dc, err := deps.HTTPClient()
		if err != nil {
			return nil, errors.Wrap(err, codes.Aborted, "missing client in http.get")
		}

		statusCode, body, headers, err := func(req *http.Request) (int, []byte, values.Object, error) {
			s, cctx := opentracing.StartSpanFromContext(ctx, "http.get")
			s.SetTag("url", req.URL.String())
			defer s.Finish()

			ccctx, cncl := context.WithTimeout(cctx, theTimeout.Duration())
			defer cncl()

			req = req.WithContext(ccctx)
			response, err := dc.Do(req)
			if err != nil {
				// Alias the DNS lookup error so as not to disclose the
				// DNS server address. This error is private in the net/http
				// package, so string matching is used.
				if strings.HasSuffix(err.Error(), "no such host") {
					return 0, nil, nil, errors.New(codes.Invalid, "no such host")
				}
				return 0, nil, nil, err
			}
			body, err := ioutil.ReadAll(response.Body)
			_ = response.Body.Close()
			if err != nil {
				return 0, nil, nil, err
			}
			s.LogFields(
				log.Int("statusCode", response.StatusCode),
				log.Int("responseSize", len(body)),
			)
			return response.StatusCode, body, headerToObject(response.Header), nil
		}(req)
		if err != nil {
			return nil, err
		}

		return values.NewObjectWithValues(map[string]values.Value{
			"statusCode": values.NewInt(int64(statusCode)),
			"headers":    headers,
			"body":       values.NewBytes(body)}), nil

	},
	true, // get has side-effects
)

func headerToObject(header http.Header) (headerObj values.Object) {
	m := make(map[string]values.Value)
	for name, thevalues := range header {
		for _, onevalue := range thevalues {
			m[name] = values.New(onevalue)
		}
	}
	return values.NewObjectWithValues(m)
}

func init() {
	runtime.RegisterPackageValue("experimental/http", "get", get)

}
