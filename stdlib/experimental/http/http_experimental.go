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
			"url":          semantic.String,
			"headers":      semantic.Tvar(1),
			"responseType": semantic.String,
		},
		Required: []string{"url"},
		Return:   semantic.NewObjectPolyType(map[string]semantic.PolyType{"statusCode": semantic.Int, "body": semantic.String, "response": semantic.String}, semantic.LabelSet{"status"}, semantic.LabelSet{"status", "body", "res"}),
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

		// Get and validate responseType
		var responseTypeValue string
		responseType, ok := args.Get("responseType")
		if !ok {
			responseTypeValue = "PING"
		} else if responseType.Str() == "BODY" {
			responseTypeValue = "BODY"
		} else if responseType.Str() == "ALL" {
			responseTypeValue = "ALL"
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

		statusCode, body, theRes, err := func(req *http.Request) (int, string, string, error) {
			s, cctx := opentracing.StartSpanFromContext(ctx, "http.get")
			s.SetTag("url", req.URL.String())
			defer s.Finish()

			req = req.WithContext(cctx)
			response, err := dc.Do(req)
			if err != nil {
				return 0, "", "", err
			}

			// Read the response body but limit how much we will read.
			// Allows socket to be reused after it is closed. (Only reusable if response emptied)
			// maxResponseBody const is defined in http.go
			limitedReader := &io.LimitedReader{R: response.Body, N: maxResponseBody}
			body, err := ioutil.ReadAll(limitedReader)
			_ = response.Body.Close()
			if err != nil {
				return 0, "", "", err
			}
			s.LogFields(
				log.Int("statusCode", response.StatusCode),
				log.Int("responseSize", len(body)),
			)

			if responseTypeValue == "ALL" {
				return response.StatusCode, string(body), strings.Join(HeaderToArray(response.Header), " "), nil
			} else if responseTypeValue == "BODY" {
				return response.StatusCode, string(body), "BODY", nil
			} else {
				return response.StatusCode, "PING", "PING", nil
			}

		}(req)
		if err != nil {
			return nil, err
		}

		// return the NewObjectPolyMap
		if theRes == "PING" {
			return values.NewObjectWithValues(map[string]values.Value{"statusCode": values.NewInt(int64(statusCode))}), nil
		} else if theRes == "BODY" {
			return values.NewObjectWithValues(map[string]values.Value{"statusCode": values.NewInt(int64(statusCode)), "body": values.NewString(body)}), nil
		}
		return values.NewObjectWithValues(map[string]values.Value{"statusCode": values.NewInt(int64(statusCode)), "body": values.NewString(body), "response": values.NewString(theRes)}), nil
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
