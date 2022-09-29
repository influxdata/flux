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
	runtime.RegisterPackageValue("experimental/dynamic", "_equal", equal)
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

		// FIXME(onelson): do we need to special case nulls here?
		//   Ex: if the input is null, should the output be null or a
		//   Dynamic with a null in it? My sense is the latter.
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
			panic("TODO")
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

		// The contract for this function says it accepts a dynamic and produces
		// an array of dynamic.
		// Note that just because we've verified the argument `v` was a dynamic
		// value holding an array, it doesn't mean the elements inside that array
		// are wrapped in dynamic.
		// Therefore, check to see if the elements are dynamic and wrap them if they aren't.
		if elmType.Nature() == semantic.Dynamic {
			arr.Retain()
			return arr, nil
		} else {
			elems := make([]values.Value, arr.Len())
			for i := 0; i < arr.Len(); i++ {
				v := arr.Get(i)
				v.Retain()
				elems[i] = values.NewDynamic(v)
			}
			dynArr := values.NewArrayWithBacking(semantic.NewArrayType(semantic.NewDynamicType()), elems)
			return dynArr, nil
		}
	},
	false,
)

var equal = values.NewFunction(
	"_equal",
	runtime.MustLookupBuiltinType("experimental/dynamic", "_equal"),
	func(ctx context.Context, args values.Object) (values.Value, error) {
		arguments := interpreter.NewArguments(args)
		a, err := arguments.GetRequired("a")
		if err != nil {
			return nil, err
		}
		if a.IsNull() {
			return values.NewBool(false), nil
		}
		if a.Type().Nature() != semantic.Dynamic {
			return nil, errors.Newf(codes.Invalid, "expected dynamic value for argument a, got %s", a.Type())
		}
		b, err := arguments.GetRequired("b")
		if err != nil {
			return nil, err
		}
		if b.IsNull() {
			return values.NewBool(false), nil
		}
		if b.Type().Nature() != semantic.Dynamic {
			return nil, errors.Newf(codes.Invalid, "expected dynamic value for argument b, got %s", b.Type())
		}
		return values.NewBool(a.Dynamic().Equal(b.Dynamic())), nil
	},
	false,
)
