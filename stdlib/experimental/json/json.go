package json

import (
	"context"
	"encoding/json"

	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

func init() {
	runtime.RegisterPackageValue("experimental/json", "parse", parse)
}

var parse = values.NewFunction(
	"parse",
	runtime.MustLookupBuiltinType("experimental/json", "parse"),
	func(ctx context.Context, args values.Object) (values.Value, error) {
		arguments := interpreter.NewArguments(args)
		data, err := arguments.GetRequired("data")
		if err != nil {
			return nil, err
		}
		if data.Type().Nature() != semantic.Bytes {
			return nil, errors.New(codes.Internal, "parse data is not of type bytes")
		}

		bytes := data.Bytes()
		return unmarshalToValue(bytes)
	},
	false,
)

func unmarshalToValue(bytes []byte) (values.Value, error) {
	var i interface{}
	err := json.Unmarshal(bytes, &i)
	if err != nil {
		return nil, err
	}
	return toValue(i)
}

// toValue converts a Go value that can be produced by json.Unmarshal into its corresponding Flux value.
func toValue(i interface{}) (values.Value, error) {
	switch t := i.(type) {
	case string:
		return values.NewString(t), nil
	case bool:
		return values.NewBool(t), nil
	case float64:
		return values.NewFloat(t), nil
	case []interface{}:
		vals := make([]values.Value, len(t))
		var elemTyp semantic.MonoType
		for i, v := range t {
			val, err := toValue(v)
			if err != nil {
				return nil, err
			}
			if elemTyp.Nature() == semantic.Invalid {
				elemTyp = val.Type()
			}
			if !val.Type().Equal(elemTyp) {
				return nil, errors.New(codes.Invalid, "array values must all be the same type")
			}
			vals[i] = val
		}
		return values.NewArrayWithBacking(semantic.NewArrayType(elemTyp), vals), nil
	case map[string]interface{}:
		vals := make(map[string]values.Value, len(t))
		for k, v := range t {
			val, err := toValue(v)
			if err != nil {
				return nil, err
			}
			vals[k] = val
		}
		return values.NewObjectWithValues(vals), nil
	}
	if i == nil {
		return values.Null, nil
	}
	return nil, errors.Newf(codes.Internal, "unsupported json type %T", i)
}
