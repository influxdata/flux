package json

import (
	"encoding/json"
	"time"

	flux "github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

func init() {
	flux.RegisterPackageValue("json", "encode", values.NewFunction(
		"encode",
		semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
			Parameters: map[string]semantic.PolyType{
				"v": semantic.Tvar(1),
			},
			Required: []string{"v"},
			Return:   semantic.Bytes,
		}),
		func(args values.Object) (values.Value, error) {
			v, ok := args.Get("v")
			if !ok {
				return nil, errors.New(codes.Invalid, "missing parameter \"v\"")
			}
			val, err := convertValue(v)
			if err != nil {
				return nil, err
			}
			bytes, err := json.Marshal(val)
			if err != nil {
				return nil, err
			}
			return values.NewBytes(bytes), nil
		},
		false,
	))
}

func convertValue(v values.Value) (interface{}, error) {
	if v.IsNull() {
		return nil, nil
	}
	switch n := v.Type().Nature(); n {
	case semantic.String:
		return v.Str(), nil
	case semantic.Bytes:
		return v.Bytes(), nil
	case semantic.Int:
		return v.Int(), nil
	case semantic.UInt:
		return v.UInt(), nil
	case semantic.Float:
		return v.Float(), nil
	case semantic.Bool:
		return v.Bool(), nil
	case semantic.Time:
		return v.Time().Time().Format(time.RFC3339Nano), nil
	case semantic.Duration:
		return v.Duration(), nil
	case semantic.Regexp:
		return v.Regexp().String(), nil
	case semantic.Array:
		arr := v.Array()
		a := make([]interface{}, arr.Len())
		var rangeErr error
		arr.Range(func(i int, v values.Value) {
			if rangeErr != nil {
				return //short circuit if we already hit an error
			}
			val, err := convertValue(v)
			if err != nil {
				rangeErr = err
				return
			}
			a[i] = val
		})
		if rangeErr != nil {
			return nil, rangeErr
		}
		return a, nil
	case semantic.Object:
		obj := v.Object()
		o := make(map[string]interface{}, obj.Len())
		var rangeErr error
		obj.Range(func(k string, v values.Value) {
			if rangeErr != nil {
				return //short circuit if we already hit an error
			}
			val, err := convertValue(v)
			if err != nil {
				rangeErr = err
				return
			}
			o[k] = val
		})
		if rangeErr != nil {
			return nil, rangeErr
		}
		return o, nil
	case semantic.Function:
		return nil, errors.New(codes.Invalid, "cannot encode a function value")
	default:
		return nil, errors.Newf(codes.Unknown, "unknown nature %v", n)
	}
}
