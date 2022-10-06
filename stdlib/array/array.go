package array

import (
	"context"

	"github.com/influxdata/flux/compiler"
	"github.com/influxdata/flux/internal/function"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

const packagePath = "array"

func registerFunctions(b *function.Builder) {
	b.Register("concat", Concat)
	b.RegisterContext("map", Map)
	b.RegisterContext("filter", Filter)
}

func Concat(args *function.Arguments) (values.Value, error) {
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
}

func Map(ctx context.Context, args *function.Arguments) (values.Value, error) {
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
	f, err := compiler.Compile(ctx, compiler.ToScope(fn.Scope), fn.Fn, inputType)
	if err != nil {
		return nil, err
	}

	var evalErr error
	elements := make([]values.Value, arr.Len())
	input := values.NewObject(inputType)
	arr.Range(func(i int, v values.Value) {
		if evalErr != nil {
			return
		}
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
}

func Filter(ctx context.Context, args *function.Arguments) (values.Value, error) {
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
	f, err := compiler.Compile(ctx, compiler.ToScope(fn.Scope), fn.Fn, inputType)
	if err != nil {
		return nil, err
	}

	var evalErr error
	elements := make([]values.Value, 0, arr.Len())
	input := values.NewObject(inputType)
	arr.Range(func(i int, v values.Value) {
		if evalErr != nil {
			return
		}
		input.Set("x", v)
		tValue, err := f.Eval(ctx, input)
		if err != nil {
			evalErr = err
			return
		}
		if tValue.Bool() {
			elements = append(elements, v)
		}
	})
	if evalErr != nil {
		return nil, evalErr
	}

	return values.NewArrayWithBacking(semantic.NewArrayType(elementType), elements), nil
}

func init() {
	b := function.ForPackage(packagePath)
	registerFunctions(&b)
	registerSource(&b)
}
