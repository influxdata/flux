package requests

import (
	"bytes"
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	fhttp "github.com/influxdata/flux/dependencies/http"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

// http get mirrors the http post originally completed for alerts & notifications
var do = values.NewFunction(
	"_do",
	runtime.MustLookupBuiltinType("http/requests", "_do"),
	func(ctx context.Context, args values.Object) (values.Value, error) {
		// Get URL
		uV, ok := args.Get("url")
		if !ok {
			return nil, errors.New(codes.Invalid, "missing \"url\" parameter")
		}
		u, err := url.Parse(uV.Str())
		if err != nil {
			return nil, err
		}
		// Get params
		var params url.Values
		if paramsV, ok := args.Get("params"); ok {
			params = make(url.Values)
			if paramsV.Type().Nature() != semantic.Dictionary {
				return nil, errors.Newf(codes.Invalid, "parameter \"params\" is not of type [string:string]: %v", paramsV.Type())
			}
			paramsDict := paramsV.Dict()
			keyType, _ := paramsDict.Type().KeyType()
			if keyType.Nature() != semantic.String {
				return nil, errors.Newf(codes.Invalid, "parameter \"params\"'s key type is not a string: %v", keyType)
			}
			valueType, _ := paramsDict.Type().ValueType()
			if valueType.Nature() != semantic.Array {
				return nil, errors.Newf(codes.Invalid, "parameter \"params\"'s value type is not a [string]: %v", valueType)
			}
			elementType, _ := valueType.ElemType()
			if elementType.Nature() != semantic.String {
				return nil, errors.Newf(codes.Invalid, "parameter \"params\"'s value type is not a string: %v", elementType)
			}
			paramsDict.Range(func(key, value values.Value) {
				k := key.Str()
				value.Array().Range(func(i int, value values.Value) {
					v := value.Str()
					params.Add(k, v)
				})
			})
		}
		if len(params) > 0 {
			u.RawQuery = params.Encode()
		}

		methodV, ok := args.Get("method")
		if !ok {
			return nil, errors.New(codes.Invalid, "missing \"method\" parameter")
		}
		if methodV.Type().Nature() != semantic.String {
			return nil, errors.Newf(codes.Invalid, "parameter \"method\" is not of type string: %v", methodV.Type())
		}
		method := methodV.Str()
		switch method {
		case "DELETE", "GET", "HEAD", "PATCH", "POST", "PUT":
		default:
			return nil, errors.Newf(codes.Invalid, "invalid HTTP method %q", method)
		}

		configV, ok := args.Get("config")
		if !ok {
			return nil, errors.New(codes.Invalid, "missing \"config\" parameter")
		}
		if configV.Type().Nature() != semantic.Object {
			return nil, errors.Newf(codes.Invalid, "parameter \"config\" is not of type record: %v", configV.Type())
		}
		config := configV.Object()

		var body io.Reader
		if bodyV, ok := args.Get("body"); ok {
			if bodyV.Type().Nature() != semantic.Bytes {
				return nil, errors.Newf(codes.Invalid, "parameter \"body\" is not of type bytes: %v", bodyV.Type())
			}
			body = bytes.NewReader(bodyV.Bytes())
		}

		// Construct HTTP request
		req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
		if err != nil {
			return nil, err
		}

		// Add headers to request
		headersV, ok := args.Get("headers")
		if ok && !headersV.IsNull() {
			if headersV.Type().Nature() != semantic.Dictionary {
				return nil, errors.Newf(codes.Invalid, "parameter \"headers\" is not of type [string:string] : %v", headersV.Type())
			}
			var rangeErr error
			headersV.Dict().Range(func(k values.Value, v values.Value) {
				if k.Type().Nature() == semantic.String &&
					v.Type().Nature() == semantic.String {
					req.Header.Set(k.Str(), v.Str())
				} else {
					rangeErr = errors.Newf(codes.Invalid, "header key and values must be a string: %q", k)
				}
			})
			if rangeErr != nil {
				return nil, rangeErr
			}
		}

		// Get Client and configure it
		deps := flux.GetDependencies(ctx)
		dc, err := deps.HTTPClient()
		if err != nil {
			return nil, errors.Wrap(err, codes.Aborted, "missing client in http.request")
		}

		timeoutV, ok := config.Get("timeout")
		if !ok {
			return nil, errors.New(codes.Invalid, "config is missing \"timeout\" property")
		}
		timeout := timeoutV.Duration()
		if timeout.IsMixed() {
			return nil, errors.New(codes.Invalid, "config timeout must not be a mixed duration")
		}
		dc, err = fhttp.WithTimeout(dc, timeout.Duration())
		if err != nil {
			return nil, err
		}

		insecureSkipVerifyV, ok := config.Get("insecureSkipVerify")
		if !ok {
			return nil, errors.New(codes.Invalid, "config is missing \"insecureSkipVerify\" property")
		}
		if insecureSkipVerifyV.Bool() {
			dc, err = fhttp.WithTLSConfig(dc, &tls.Config{
				InsecureSkipVerify: true,
			})
			if err != nil {
				return nil, err
			}
		}

		// Do request, using local anonymous functions to facilitate timing the request
		statusCode, responseBody, headers, duration, err := func(req *http.Request) (statusCode int, body []byte, headers values.Dictionary, duration time.Duration, err error) {
			startTime := time.Now()
			s, cctx := opentracing.StartSpanFromContext(req.Context(), "requests._do", opentracing.StartTime(startTime))
			s.SetTag("url", req.URL.String())
			defer func() {
				finishTime := time.Now()
				s.FinishWithOptions(opentracing.FinishOptions{
					FinishTime: finishTime,
				})
				// set duration to return
				duration = finishTime.Sub(startTime)
			}()

			req = req.WithContext(cctx)
			response, err := dc.Do(req)
			if err != nil {
				// Alias the DNS lookup error so as not to disclose the
				// DNS server address. This error is private in the net/http
				// package, so string matching is used.
				if strings.HasSuffix(err.Error(), "no such host") {
					err = errors.New(codes.Invalid, "no such host")
					return
				}
				return
			}
			body, err = io.ReadAll(response.Body)
			_ = response.Body.Close()
			if err != nil {
				return
			}
			s.LogFields(
				log.Int("statusCode", response.StatusCode),
				log.Int("responseSize", len(body)),
			)
			headers, err = headerToDict(response.Header)
			if err != nil {
				return
			}
			statusCode = response.StatusCode
			return
		}(req)
		if err != nil {
			return nil, err
		}

		return values.NewObjectWithValues(map[string]values.Value{
			"statusCode": values.NewInt(int64(statusCode)),
			"headers":    headers,
			"body":       values.NewBytes(responseBody),
			"duration":   values.NewDuration(values.ConvertDurationNsecs(duration)),
		}), nil

	},
	true, // _do has side-effects
)

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

func init() {
	runtime.RegisterPackageValue("http/requests", "_do", do)
}
