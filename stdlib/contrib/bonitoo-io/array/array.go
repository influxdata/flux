package array

import (
	"context"

	"github.com/influxdata/flux/compiler"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

var SpecialFns map[string]values.Function

const packagePath = "contrib/bonitoo-io/array"

func init() {
	SpecialFns = map[string]values.Function{
		"concat": values.NewFunction(
			"concat",
			runtime.MustLookupBuiltinType(packagePath, "concat"),
			func(ctx context.Context, args values.Object) (values.Value, error) {
				return interpreter.DoFunctionCallContext(func(ctx context.Context, args interpreter.Arguments) (values.Value, error) {
					_v, err := args.GetRequired("v")
					if err != nil {
						return nil, err
					}
					v := _v.Array()

					if v.Len() == 0 {
						_arr, err := args.GetRequired("arr")
						if err != nil {
							return nil, err
						}
						return _arr, nil
					}

					elementType, err := v.Type().ElemType()
					if err != nil {
						return nil, err
					}

					arr, err := args.GetRequiredArrayAllowEmpty("arr", elementType.Nature())
					if err != nil {
						return nil, err
					}

					m := arr.Len()
					n := v.Len()
					elements := make([]values.Value, m+n)
					arr.Range(func(i int, v values.Value) {
						elements[i] = v
					})
					v.Range(func(i int, v values.Value) {
						elements[m+i] = v
					})

					return values.NewArrayWithBacking(semantic.NewArrayType(elementType), elements), nil
				}, ctx, args)
			}, false,
		),
		"map": values.NewFunction(
			"map",
			runtime.MustLookupBuiltinType(packagePath, "map"),
			func(ctx context.Context, args values.Object) (values.Value, error) {
				return interpreter.DoFunctionCallContext(func(ctx context.Context, args interpreter.Arguments) (values.Value, error) {
					_fn, err := args.GetRequiredFunction("fn")
					if err != nil {
						return nil, err
					}
					fn, err := interpreter.ResolveFunction(_fn)
					if err != nil {
						return nil, err
					}

					_arr, err := args.GetRequired("arr")
					if err != nil {
						return nil, err
					}
					arr := _arr.Array()

					if arr.Len() == 0 {
						return arr, nil
					}

					elementType, err := arr.Type().ElemType()
					if err != nil {
						return nil, err
					}
					inputType := semantic.NewObjectType([]semantic.PropertyType{
						{Key: []byte("x"), Value: elementType},
					})
					f, err := compiler.Compile(compiler.ToScope(fn.Scope), fn.Fn, inputType)
					if err != nil {
						return nil, err
					}

					var evalErr error
					elements := make([]values.Value, arr.Len())
					input := values.NewObject(inputType)
					arr.Range(func(i int, v values.Value) {
						input.Set("x", v)
						tValue, err := f.Eval(ctx, input)
						if err != nil {
							evalErr = err
							return
						}
						elements[i] = tValue
					})
					if evalErr != nil {
						return nil, evalErr
					}

					return values.NewArrayWithBacking(semantic.NewArrayType(elements[0].Type()), elements), nil
				}, ctx, args)
			}, false,
		),
	}

	runtime.RegisterPackageValue(packagePath, "concat", SpecialFns["concat"])
	runtime.RegisterPackageValue(packagePath, "map", SpecialFns["map"])
}
