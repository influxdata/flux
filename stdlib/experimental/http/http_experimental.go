package http

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

// maxResponseBody is the maximum response body we will read before just discarding
// the rest. This allows sockets to be reused.
const maxResponseBody = 512 * 1024 // 512 KB

// http get mirrors the http post originally completed for alerts & notifications
var get = values.NewFunction(
	"get",
	semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
		Parameters: map[string]semantic.PolyType{
			"url":             semantic.String,
			"headers":         semantic.Tvar(1),
			"responseHeaders": semantic.Bool,
		},
		Required: []string{"url"},
		Return:   semantic.NewObjectPolyType(map[string]semantic.PolyType{"statusCode": semantic.Int, "body": semantic.Bytes, "headers": semantic.String}, semantic.LabelSet{"status", "body"}, semantic.LabelSet{"status", "body", "headers"}),
	}),
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
			return nil, err
		}

		// Get and validate responseHeaders
		var includeResHeaders bool
		responseHeaders, ok := args.Get("responseHeaders")
		if !ok || responseHeaders.Type() != semantic.Bool {
			includeResHeaders = false
		} else if responseHeaders.Bool() {
			includeResHeaders = true
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
				if v.Type() == semantic.String {
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

		statusCode, body, headers, err := func(req *http.Request) (int, []byte, string, error) {
			s, cctx := opentracing.StartSpanFromContext(ctx, "http.get")
			s.SetTag("url", req.URL.String())
			defer s.Finish()

			req = req.WithContext(cctx)
			response, err := dc.Do(req)
			if err != nil {
				return 0, nil, "", err
			}

			// Read the response body but limit how much we will read.
			// Allows socket to be reused after it is closed. (Only reusable if response emptied)
			// maxResponseBody const is defined in http.go
			limitedReader := &io.LimitedReader{R: response.Body, N: maxResponseBody}
			body, err := ioutil.ReadAll(limitedReader)
			_ = response.Body.Close()
			if err != nil {
				return 0, nil, "", err
			}
			s.LogFields(
				log.Int("statusCode", response.StatusCode),
				log.Int("responseSize", len(body)),
			)

			if includeResHeaders {
				return response.StatusCode, body, strings.Join(HeaderToArray(response.Header), " "), nil
			} else {
				return response.StatusCode, body, "", nil
			}

		}(req)
		if err != nil {
			return nil, err
		}

		// return the NewObjectPolyMap
		if headers == "" {
			return values.NewObjectWithValues(map[string]values.Value{"statusCode": values.NewInt(int64(statusCode)), "body": values.NewBytes(body)}), nil
		} else {
			return values.NewObjectWithValues(map[string]values.Value{"statusCode": values.NewInt(int64(statusCode)), "body": values.NewBytes(body), "headers": values.NewString(headers)}), nil
		}

	},
	true, // get has side-effects
)

func HeaderToArray(header http.Header) (res []string) {
	for name, values := range header {
		for _, value := range values {
			res = append(res, fmt.Sprintf("%s: %s", name, value))
		}
	}
	return
}

func init() {
	flux.RegisterPackageValue("experimental/http", "get", get)

}
