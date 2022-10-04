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

		// Nothing to do if the incoming value is already a dynamic
		if v.Type().Nature() == semantic.Dynamic {
			return v, nil
		}
		return wrapValue(v)
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

func wrapValue(v values.Value) (values.Dynamic, error) {
	if v.IsNull() {
		return values.NewDynamic(v), nil
	}
	switch n := v.Type().Nature(); n {
	case semantic.Dynamic:
		return v.Dynamic(), nil // Return as-is

	// Basic types wrap plainly.
	case semantic.String,
		semantic.Bytes,
		semantic.Int,
		semantic.UInt,
		semantic.Float,
		semantic.Bool,
		semantic.Time,
		semantic.Duration:
		return values.NewDynamic(v), nil

	// The composite types need to recurse.
	case semantic.Array:
		arr := v.Array()
		elems := make([]values.Value, arr.Len())
		var rangeErr error
		arr.Range(func(i int, v values.Value) {
			if rangeErr != nil {
				return // short circuit if we already hit an error
			}
			val, err := wrapValue(v)
			if err != nil {
				rangeErr = err
				return
			}
			elems[i] = val
		})
		if rangeErr != nil {
			return nil, rangeErr
		}
		return values.NewDynamic(
			values.NewArrayWithBacking(
				semantic.NewArrayType(semantic.NewDynamicType()),
				elems,
			)), nil
	case semantic.Object:
		obj := v.Object()
		o := make(map[string]values.Value, obj.Len())
		var rangeErr error
		obj.Range(func(k string, v values.Value) {
			if rangeErr != nil {
				return // short circuit if we already hit an error
			}
			val, err := wrapValue(v)
			if err != nil {
				rangeErr = err
				return
			}
			o[k] = val
		})
		if rangeErr != nil {
			return nil, rangeErr
		}
		return values.NewDynamic(values.NewObjectWithValues(o)), nil

	case semantic.Dictionary:
		// FIXME(onelson): should dict work? Seems like member expressions should work as-is on dict.
		panic("TODO")
	// FIXME(onelson): stands to reason we should only allow wrapping of types that can be cast INTO
	case semantic.Regexp:
	case semantic.Function:
	case semantic.Stream:
	default:
		return nil, errors.Newf(codes.Unknown, "unknown nature %v", n)
	}

	return nil, errors.Newf(codes.Invalid, "unsupported type for dynamic %s", v.Type())
}
