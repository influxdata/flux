package dynamic

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

func init() {
	runtime.RegisterPackageValue("experimental/dynamic", "jsonParse", jsonParse)
	runtime.RegisterPackageValue("experimental/dynamic", "jsonEncode", jsonEncode)
}

var jsonParse = values.NewFunction(
	"jsonParse",
	runtime.MustLookupBuiltinType("experimental/dynamic", "jsonParse"),
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
	inner, err := toValue(i)
	if err != nil {
		return nil, err
	}
	return values.NewDynamic(inner), nil
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
		return toArray(t)
	case map[string]interface{}:
		return toObject(t)
	}
	if i == nil {
		return values.Null, nil
	}
	return nil, errors.Newf(codes.Internal, "unsupported json type %T", i)
}

func toArray(a []interface{}) (values.Array, error) {
	vals := make([]values.Value, len(a))
	for i, v := range a {
		val, err := toValue(v)
		if err != nil {
			return nil, err
		}

		vals[i] = values.NewDynamic(val)
	}
	arrayTyp := semantic.NewArrayType(semantic.NewDynamicType())
	return values.NewArrayWithBacking(arrayTyp, vals), nil
}

func toObject(m map[string]interface{}) (values.Value, error) {
	vals := make(map[string]values.Value, len(m))
	for k, v := range m {
		val, err := toValue(v)
		if err != nil {
			return nil, err
		}
		vals[k] = values.NewDynamic(val)
	}
	return values.NewObjectWithValues(vals), nil
}

var jsonEncode = values.NewFunction(
	"encode",
	runtime.MustLookupBuiltinType("experimental/dynamic", "jsonEncode"),
	func(ctx context.Context, args values.Object) (values.Value, error) {
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
)

// convertValue takes a Flux value and converts it to Go native values
// that Go's standard JSON serialization can work with.
//
// This code is only slightly modified from the non-experimental
// version in stdlib/json/encode.go.
func convertValue(v values.Value) (interface{}, error) {
	if v.IsNull() {
		return nil, nil
	}
	switch n := v.Type().Nature(); n {
	case semantic.Dynamic:
		return convertValue(v.Dynamic().Inner())
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
				return // short circuit if we already hit an error
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
	case semantic.Stream:
		return nil, errors.New(
			codes.Invalid,
			"got table stream instead of array. "+
				"Try using tableFind() or findRecord() to extract data from stream")
	case semantic.Object:
		obj := v.Object()
		o := make(map[string]interface{}, obj.Len())
		var rangeErr error
		obj.Range(func(k string, v values.Value) {
			if rangeErr != nil {
				return // short circuit if we already hit an error
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
	case semantic.Dictionary:
		dict := v.Dict()
		// Go JSON encoder requires that map key type is either a primitive type or implements encoding.TextMarshaler interface.
		// Since Go maps are encoded as JSON objects with string keys (https://www.json.org/json-en.html), and dictionary keys
		// are primitive Flux types, we can safely convert Flux dictionary to Go map with string keys.
		d := make(map[string]interface{}, dict.Len())
		var rangeErr error
		var b strings.Builder
		dict.Range(func(k, v values.Value) {
			if rangeErr != nil {
				return // short circuit if we already hit an error
			}
			b.Reset()
			err := values.Display(&b, k)
			if err != nil {
				rangeErr = err
				return
			}
			key := b.String()
			val, err := convertValue(v)
			if err != nil {
				rangeErr = err
				return
			}
			d[key] = val
		})
		if rangeErr != nil {
			return nil, rangeErr
		}
		return d, nil
	default:
		return nil, errors.Newf(codes.Unknown, "unknown nature %v", n)
	}
}
