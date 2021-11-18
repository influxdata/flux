package types

import (
	"context"

	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

const IsTypeKind = "isType"

func init() {
	runtime.RegisterPackageValue("types", IsTypeKind, IsType())
}

func IsType() values.Function {
	return values.NewFunction(
		IsTypeKind,
		runtime.MustLookupBuiltinType("types", IsTypeKind),
		func(ctx context.Context, args values.Object) (values.Value, error) {
			return interpreter.DoFunctionCallContext(func(ctx context.Context, args interpreter.Arguments) (values.Value, error) {
				v, err := args.GetRequired("v")
				if err != nil {
					return nil, err
				}
				type_, err := args.GetRequiredString("type")
				if err != nil {
					return nil, err
				}
				result := false
				switch v.Type().Nature() {
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
				// We explicitly only support the primitive types as we do not want callers to
				// confuse `[int]` and `[string]` etc
				// case semantic.Array:
				// 	result = type_ == "array"
				// case semantic.Object:
				// 	result = type_ == "object"
				// case semantic.Function;
				// 	result = type_ == "function"
				// case semantic.Dictionary:
				// 	result = type_ == "dictionary"
				// case semantic.Vector:
				// 	result = type_ == "vector"
				default:
					result = false
				}

				return values.NewBool(result), nil
			}, ctx, args)
		}, false,
	)
}
