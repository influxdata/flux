package math

import (
	"context"

	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

const pkgpath = "contrib/jsternberg/math"

func init() {
	runtime.RegisterPackageValue(pkgpath, "minIndex", values.NewFunction(
		"minIndex",
		runtime.MustLookupBuiltinType(pkgpath, "minIndex"),
		func(ctx context.Context, args values.Object) (values.Value, error) {
			return interpreter.DoFunctionCall(MinIndex, args)
		},
		false,
	))

	runtime.RegisterPackageValue(pkgpath, "maxIndex", values.NewFunction(
		"maxIndex",
		runtime.MustLookupBuiltinType(pkgpath, "maxIndex"),
		func(ctx context.Context, args values.Object) (values.Value, error) {
			return interpreter.DoFunctionCall(MaxIndex, args)
		},
		false,
	))

	runtime.RegisterPackageValue(pkgpath, "sum", values.NewFunction(
		"sum",
		runtime.MustLookupBuiltinType(pkgpath, "sum"),
		func(ctx context.Context, args values.Object) (values.Value, error) {
			return interpreter.DoFunctionCall(Sum, args)
		},
		false,
	))
}

func MinIndex(args interpreter.Arguments) (values.Value, error) {
	arr, err := args.GetRequired("values")
	if err != nil {
		return nil, err
	}

	arrType := arr.Type()
	elemType, err := arrType.ElemType()
	if err != nil {
		return nil, err
	}

	index, err := func() (int64, error) {
		switch elemType.Nature() {
		case semantic.Int:
			return intMinIndex(arr.Array()), nil
		case semantic.UInt:
			return uintMinIndex(arr.Array()), nil
		case semantic.Float:
			return floatMinIndex(arr.Array()), nil
		default:
			return -1, errors.New(codes.Unimplemented)
		}
	}()
	if err != nil {
		return nil, err
	}
	return values.NewInt(index), nil
}

func floatMinIndex(arr values.Array) int64 {
	var min float64
	index := -1
	arr.Range(func(i int, v values.Value) {
		if index < 0 || v.Float() < min {
			min, index = v.Float(), i
		}
	})
	return int64(index)
}

func intMinIndex(arr values.Array) int64 {
	var min int64
	index := -1
	arr.Range(func(i int, v values.Value) {
		if index < 0 || v.Int() < min {
			min, index = v.Int(), i
		}
	})
	return int64(index)
}

func uintMinIndex(arr values.Array) int64 {
	var min uint64
	index := -1
	arr.Range(func(i int, v values.Value) {
		if index < 0 || v.UInt() < min {
			min, index = v.UInt(), i
		}
	})
	return int64(index)
}

func MaxIndex(args interpreter.Arguments) (values.Value, error) {
	arr, err := args.GetRequired("values")
	if err != nil {
		return nil, err
	}

	arrType := arr.Type()
	elemType, err := arrType.ElemType()
	if err != nil {
		return nil, err
	}

	index, err := func() (int64, error) {
		switch elemType.Nature() {
		case semantic.Int:
			return intMaxIndex(arr.Array()), nil
		case semantic.UInt:
			return uintMaxIndex(arr.Array()), nil
		case semantic.Float:
			return floatMaxIndex(arr.Array()), nil
		default:
			return -1, errors.New(codes.Unimplemented)
		}
	}()
	if err != nil {
		return nil, err
	}
	return values.NewInt(index), nil
}

func floatMaxIndex(arr values.Array) int64 {
	var max float64
	index := -1
	arr.Range(func(i int, v values.Value) {
		if index < 0 || v.Float() > max {
			max, index = v.Float(), i
		}
	})
	return int64(index)
}

func intMaxIndex(arr values.Array) int64 {
	var max int64
	index := -1
	arr.Range(func(i int, v values.Value) {
		if index < 0 || v.Int() > max {
			max, index = v.Int(), i
		}
	})
	return int64(index)
}

func uintMaxIndex(arr values.Array) int64 {
	var max uint64
	index := -1
	arr.Range(func(i int, v values.Value) {
		if index < 0 || v.UInt() > max {
			max, index = v.UInt(), i
		}
	})
	return int64(index)
}

func Sum(args interpreter.Arguments) (values.Value, error) {
	arr, err := args.GetRequired("values")
	if err != nil {
		return nil, err
	}

	arrType := arr.Type()
	elemType, err := arrType.ElemType()
	if err != nil {
		return nil, err
	}

	switch elemType.Nature() {
	case semantic.Int:
		return values.NewInt(intSum(arr.Array())), nil
	case semantic.UInt:
		return values.NewUInt(uintSum(arr.Array())), nil
	case semantic.Float:
		return values.NewFloat(floatSum(arr.Array())), nil
	default:
		return nil, errors.New(codes.Unimplemented)
	}
}

func floatSum(arr values.Array) float64 {
	var sum float64
	arr.Range(func(i int, v values.Value) {
		sum += v.Float()
	})
	return sum
}

func intSum(arr values.Array) int64 {
	var sum int64
	arr.Range(func(i int, v values.Value) {
		sum += v.Int()
	})
	return sum
}

func uintSum(arr values.Array) uint64 {
	var sum uint64
	arr.Range(func(i int, v values.Value) {
		sum += v.UInt()
	})
	return sum
}
