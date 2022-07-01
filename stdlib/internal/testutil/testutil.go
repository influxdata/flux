package testutil

import (
	"context"
	"regexp"

	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

func init() {
	runtime.RegisterPackageValue("internal/testutil", "fail", values.NewFunction(
		"fail",
		runtime.MustLookupBuiltinType("internal/testutil", "fail"),
		func(ctx context.Context, args values.Object) (values.Value, error) {
			return nil, errors.New(codes.Aborted, "fail")
		},
		false,
	))
	runtime.RegisterPackageValue("internal/testutil", "yield", values.NewFunction(
		"yield",
		runtime.MustLookupBuiltinType("internal/testutil", "yield"),
		func(ctx context.Context, args values.Object) (values.Value, error) {
			in, ok := args.Get("v")
			if !ok {
				return nil, errors.Newf(codes.Invalid, "missing required keyword argument %q", "v")
			}
			return in, nil
		}, true))
	runtime.RegisterPackageValue("internal/testutil", "makeRecord", values.NewFunction(
		"makeRecord",
		runtime.MustLookupBuiltinType("internal/testutil", "makeRecord"),
		func(ctx context.Context, args values.Object) (values.Value, error) {
			in, ok := args.Get("o")
			if !ok {
				return nil, errors.Newf(codes.Invalid, "missing required keyword argument %q", "o")
			}
			return in, nil
		}, false))
	runtime.RegisterPackageValue("internal/testutil", "makeAny", values.NewFunction(
		"makeAny",
		runtime.MustLookupBuiltinType("internal/testutil", "makeAny"),
		func(ctx context.Context, args values.Object) (values.Value, error) {
			arg, ok := args.Get("typ")
			if !ok {
				panic(`could not find argument "typ"`)
			}
			switch arg.Str() {
			case semantic.String.String():
				return values.NewString("foo"), nil
			case semantic.Bytes.String():
				return values.NewBytes([]byte("foo")), nil
			case semantic.Int.String():
				return values.NewInt(10), nil
			case semantic.UInt.String():
				return values.NewUInt(10), nil
			case semantic.Float.String():
				return values.NewFloat(10.0), nil
			case semantic.Bool.String():
				return values.NewBool(false), nil
			case semantic.Time.String():
				return values.NewTime(100), nil
			case semantic.Duration.String():
				return values.NewDuration(values.Duration{}), nil
			case semantic.Regexp.String():
				return values.NewRegexp(regexp.MustCompile("a")), nil
			case semantic.Array.String():
				return values.NewArray(semantic.NewArrayType(semantic.BasicInt)), nil
			case semantic.Object.String():
				return values.NewObjectWithValues(map[string]values.Value{}), nil
			case semantic.Function.String():
				return values.NewFunction(
					"returnOne",
					semantic.NewFunctionType(semantic.BasicInt, nil),
					func(ctx context.Context, args values.Object) (values.Value, error) {
						return values.NewInt(1), nil
					},
					false,
				), nil
			case semantic.Dictionary.String():
				return values.NewEmptyDict(semantic.NewDictType(semantic.BasicInt, semantic.BasicString)), nil
			default:
				return values.Null, nil
			}
		},
		false,
	))
}
