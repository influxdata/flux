package http

import (
	"bytes"
	"context"
	"io"
	"net"
	"net/http"
	"net/url"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/iocounter"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

func init() {
	runtime.RegisterPackageValue("http", "post", values.NewFunction(
		"post",
		runtime.MustLookupBuiltinType("http", "post"),
		func(ctx context.Context, args values.Object) (values.Value, error) {
			// Get URL
			uV, ok := args.Get("url")
			if !ok {
				return nil, errors.New(codes.Invalid, "missing \"url\" parameter")
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
			deps := flux.GetDependencies(ctx)
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
					// If an error is returned during a request (from the control
					// function where our IP validator runs), the original error
					// is attached to a field of OpError which is wrapped in a
					// url.Error. Attempt to unwrap these to get the original
					// cause.
					if urlErr, ok := err.(*url.Error); ok {
						if urlErr.Err != nil {
							if opErr, ok := urlErr.Err.(*net.OpError); ok {
								if opErr.Err != nil {
									return 0, opErr.Err
								}
							}
						}
					}
					return 0, err
				}

				wc := iocounter.Writer{Writer: io.Discard}
				_, _ = io.Copy(&wc, response.Body)
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
	))
}
