package http

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/iocounter"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

// maxResponseBody is the maximum response body we will read before just discarding
// the rest. This allows sockets to be reused.
const maxResponseBody = 512 * 1024 // 512 KB

var getWithBody = values.NewFunction(
	"getBody",
	semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
		Parameters: map[string]semantic.PolyType{
			"url":     semantic.String,
			"headers": semantic.Tvar(1),
		},
		Required: []string{"url"},
		Return:   semantic.String,
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

		statusCode, body, err := func(req *http.Request) (int, []byte, error) {
			s, cctx := opentracing.StartSpanFromContext(ctx, "http.getBasedOnResponseType")
			s.SetTag("url", req.URL.String())
			defer s.Finish()

			req = req.WithContext(cctx)
			response, err := dc.Do(req)
			if err != nil {
				return 0, nil, err
			}

			// Read the response body but limit how much we will read.
			// This is to allow a socket to be reused after it is closed. (Only reusable if emptied)

			limitedReader := &io.LimitedReader{R: response.Body, N: maxResponseBody}
			body, err := ioutil.ReadAll(limitedReader)
			_ = response.Body.Close()
			s.LogFields(
				log.Int("statusCode", response.StatusCode),
				log.Int("responseSize", len(body)),
			)

			return response.StatusCode, body, nil
		}(req)
		if err != nil {
			return nil, err
		}

		var resultValue []values.Value
		resultValue = append(resultValue, values.NewInt(int64(statusCode)))
		if body != nil {
			resultValue = append(resultValue, values.NewString(string(body)))
		}
		// return status code and body
		return values.NewArrayWithBacking(semantic.String, resultValue), nil
	},
	true, // get has side-effects
)

// http get mirrors the http post originally completed for alerts & notifications
var get = values.NewFunction(
	"get",
	semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
		Parameters: map[string]semantic.PolyType{
			"url":     semantic.String,
			"headers": semantic.Tvar(1),
		},
		Required: []string{"url"},
		Return:   semantic.Int,
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

		statusCode, err := func(req *http.Request) (int, error) {
			s, cctx := opentracing.StartSpanFromContext(ctx, "http.get")
			s.SetTag("url", req.URL.String())
			defer s.Finish()

			req = req.WithContext(cctx)
			response, err := dc.Do(req)
			if err != nil {
				return 0, err
			}

			// Read the response body but limit how much we will read.
			// This is to allow a socket to be reused after it is closed.
			wc := iocounter.Writer{Writer: ioutil.Discard}
			r := io.LimitedReader{
				R: response.Body,
				N: maxResponseBody,
			}
			_, _ = io.Copy(&wc, &r)
			_ = response.Body.Close()
			s.LogFields(
				log.Int("statusCode", response.StatusCode),
				log.Int64("responseSize", wc.Count()),
			)

			return response.StatusCode, nil
		}(req)
		if err != nil {
			return nil, err
		}

		// return status code
		return values.NewInt(int64(statusCode)), nil
	},
	true, // get has side-effects
)

var post = values.NewFunction(
	"post",
	semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
		Parameters: map[string]semantic.PolyType{
			"url":     semantic.String,
			"headers": semantic.Tvar(1),
			"data":    semantic.Bytes,
		},
		Required: []string{"url"},
		Return:   semantic.Int,
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

		// Construct data
		var data []byte
		dataV, ok := args.Get("data")
		if ok {
			data = dataV.Bytes()
		}

		// Construct HTTP request
		req, err := http.NewRequest("POST", uV.Str(), bytes.NewReader(data))
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
			return nil, errors.Wrap(err, codes.Aborted, "missing client in http.post")
		}

		statusCode, err := func(req *http.Request) (int, error) {
			s, cctx := opentracing.StartSpanFromContext(ctx, "http.post")
			s.SetTag("url", req.URL.String())
			defer s.Finish()

			req = req.WithContext(cctx)
			response, err := dc.Do(req)
			if err != nil {
				return 0, err
			}

			// Read the response body but limit how much we will read.
			// This is to allow a socket to be reused after it is closed.
			wc := iocounter.Writer{Writer: ioutil.Discard}
			r := io.LimitedReader{
				R: response.Body,
				N: maxResponseBody,
			}
			_, _ = io.Copy(&wc, &r)
			_ = response.Body.Close()
			s.LogFields(
				log.Int("statusCode", response.StatusCode),
				log.Int64("responseSize", wc.Count()),
			)
			return response.StatusCode, nil
		}(req)
		if err != nil {
			return nil, err
		}

		// return status code
		return values.NewInt(int64(statusCode)), nil
	},
	true, // post has side-effects
)

func init() {
	flux.RegisterPackageValue("http", "post", post)
	flux.RegisterPackageValue("http", "get", get)
	flux.RegisterPackageValue("http", "getWithBody", getWithBody)
}
