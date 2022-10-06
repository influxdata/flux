package dynamic

import (
	"context"

	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

func init() {
	runtime.RegisterPackageValue("experimental/dynamic", "dynamic", dynamicConv)
	runtime.RegisterPackageValue("experimental/dynamic", "asArray", asArray)
	runtime.RegisterPackageValue("experimental/dynamic", "isType", isType)
}

var dynamicConv = values.NewFunction(
	"dynamic",
	runtime.MustLookupBuiltinType("experimental/dynamic", "dynamic"),
	func(ctx context.Context, args values.Object) (values.Value, error) {
		arguments := interpreter.NewArguments(args)
		v, err := arguments.GetRequired("v")
		if err != nil {
			return nil, err
		}
		// N.b. all types are accepted, but only a subset have ways to be
		// extracted.
		// Add checks here for the nature of `v` if you want to make certain
		// types a runtime error.
		return values.NewDynamic(v), nil
	},
	false,
)

var asArray = values.NewFunction(
	"asArray",
	runtime.MustLookupBuiltinType("experimental/dynamic", "asArray"),
	func(ctx context.Context, args values.Object) (values.Value, error) {
		arguments := interpreter.NewArguments(args)
		v, err := arguments.GetRequired("v")
		if err != nil {
			return nil, err
		}

		if v.IsNull() {
			return nil, errors.Newf(codes.Invalid, "unable to convert <null> to array")
		}

		d := v.Dynamic()
		inner := d.Inner()
		if inner.Type().Nature() != semantic.Array {
			return nil, errors.Newf(codes.Invalid, "unable to convert %s to array", inner.Type())
		}
		arr := inner.Array()
		elmType, err := arr.Type().ElemType()
		if err != nil {
			return nil, err
		}

		if elmType.Nature() != semantic.Dynamic {
			return nil, errors.Newf(codes.Internal, "expected array to have dynamic elements, got %s", elmType)
		}

		return arr, nil
	},
	false,
)

var isType = values.NewFunction(
	"isType",
	runtime.MustLookupBuiltinType("experimental/dynamic", "isType"),
	func(ctx context.Context, args values.Object) (values.Value, error) {
		arguments := interpreter.NewArguments(args)
		v, err := arguments.GetRequired("v")
		if err != nil {
			return nil, err
		}

		// Normally nulls would land in the default case since we don't have an
		// explicit check for `semantic.Invalid`, but this early return avoids
		// the panic we'd see when trying to access the inner value of the
		// dynamic (which is not valid when we get a true `null` rather than a
		// `dynamic(null)`).
		if v.IsNull() {
			return values.NewBool(false), nil
		}

		type_, err := arguments.GetRequiredString("type")
		if err != nil {
			return nil, err
		}
		result := false

		d := v.Dynamic()
		inner := d.Inner()
		switch inner.Type().Nature() {
		case semantic.String:
			result = type_ == "string"
		case semantic.Bytes:
			result = type_ == "bytes"
		case semantic.Int:
			result = type_ == "int"
		case semantic.UInt:
			result = type_ == "uint"
		case semantic.Float:
			result = type_ == "float"
		case semantic.Bool:
			result = type_ == "bool"
		case semantic.Time:
			result = type_ == "time"
		case semantic.Duration:
			result = type_ == "duration"
		case semantic.Regexp:
			result = type_ == "regexp"
		case semantic.Array:
			result = type_ == "array"
		case semantic.Object:
			result = type_ == "object"
		case semantic.Function:
			result = type_ == "function"
		case semantic.Dictionary:
			result = type_ == "dictionary"
		case semantic.Vector:
			result = type_ == "vector"
		default:
			result = false
		}
		return values.NewBool(result), nil
	},
	false,
)
