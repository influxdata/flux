package http

import (
	"bytes"
	"net/http"

	flux "github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
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
		func(args values.Object) (values.Value, error) {
			url, ok := args.Get("url")
			if !ok {
				return nil, errors.New(codes.Invalid, "missing \"url\" parameter")
			}
			var data []byte
			dataV, ok := args.Get("data")
			if ok {
				data = dataV.Bytes()
			}

			// Construct HTTP request
			req, err := http.NewRequest("POST", url.Str(), bytes.NewReader(data))
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
			response, err := http.DefaultClient.Do(req)
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
