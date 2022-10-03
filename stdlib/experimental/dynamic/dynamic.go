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

		// FIXME(onelson): wrap the value recursively, not just if v is an Array
		//  We want to produce a dynamic where every single value contained
		//  within it is also wrapped with a dynamic.
		if v.Type().Nature() == semantic.Array {
			arr := v.Array()
			elmType, err := arr.Type().ElemType()
			if err != nil {
				return nil, err
			}
			if elmType.Nature() == semantic.Dynamic {
				return values.NewDynamic(arr), nil
			} else {
				elems := make([]values.Value, arr.Len())
				for i := 0; i < arr.Len(); i++ {
					v := arr.Get(i)
					elems[i] = values.NewDynamic(v)
				}
				dynArr := values.NewArrayWithBacking(semantic.NewArrayType(semantic.NewDynamicType()), elems)
				return values.NewDynamic(dynArr), nil
			}
		}

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
