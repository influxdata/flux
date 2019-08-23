package http

import (
	"bytes"
	"context"
	"net/http"
	"net/url"

	flux "github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/dependencies"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

func init() {
	flux.RegisterPackageValue("http", "post", values.NewFunction(
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
		func(ctx context.Context, deps dependencies.Interface, args values.Object) (values.Value, error) {
			// Get and validate URL
			uV, ok := args.Get("url")
			if !ok {
				return nil, errors.New(codes.Invalid, "missing \"url\" parameter")
			}
			u, err := url.Parse(uV.Str())
			if err != nil {
				return nil, err
			}
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
			response, err := dc.Do(req)
			if err != nil {
				return nil, err
			}
			defer response.Body.Close()

			// return status code
			return values.NewInt(int64(response.StatusCode)), nil
		},
		true, // post has side-effects
	))
}
