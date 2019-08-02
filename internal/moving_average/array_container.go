package moving_average

import (
	"github.com/apache/arrow/go/arrow/array"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/values"
)

type ArrayContainer struct {
	array array.Interface
}

func NewArrayContainer(a array.Interface) *ArrayContainer {
	return &ArrayContainer{a}
}

func (a *ArrayContainer) IsNull(i int) bool {
	return a.array.IsNull(i)
}

func (a *ArrayContainer) IsValid(i int) bool {
	return a.array.IsValid(i)
}

func (a *ArrayContainer) Len() int {
	return a.array.Len()
}

func (a *ArrayContainer) Value(i int) values.Value {
	switch a.array.(type) {
	case *array.Boolean:
		return values.New(a.array.(*array.Boolean).Value(i))
	case *array.Int64:
		return values.New(float64(a.array.(*array.Int64).Value(i)))
	case *array.Uint64:
		return values.New(float64(a.array.(*array.Uint64).Value(i)))
	case *array.Float64:
		return values.New(float64(a.array.(*array.Float64).Value(i)))
	case *array.Binary:
		return values.New(string(a.array.(*array.Binary).Value(i)))
	default:
		return nil
	}
}

func (a *ArrayContainer) OrigValue(i int) interface{} {
	switch a.array.(type) {
	case *array.Boolean:
		return a.array.(*array.Boolean).Value(i)
	case *array.Int64:
		return a.array.(*array.Int64).Value(i)
	case *array.Uint64:
		return a.array.(*array.Uint64).Value(i)
	case *array.Float64:
		return a.array.(*array.Float64).Value(i)
	case *array.Binary:
		return string(a.array.(*array.Binary).Value(i))
	default:
		return nil
	}
}

func (a *ArrayContainer) Slice(i int, j int) *ArrayContainer {
	slice := &ArrayContainer{}
	switch a.array.(type) {
	case *array.Boolean:
		slice.array = arrow.BoolSlice(a.array.(*array.Boolean), i, j)
	case *array.Int64:
		slice.array = arrow.IntSlice(a.array.(*array.Int64), i, j)
	case *array.Uint64:
		slice.array = arrow.UintSlice(a.array.(*array.Uint64), i, j)
	case *array.Float64:
		slice.array = arrow.FloatSlice(a.array.(*array.Float64), i, j)
	case *array.Binary:
		slice.array = arrow.StringSlice(a.array.(*array.Binary), i, j)
	default:
		slice.array = nil
	}
	return slice
}

func (a *ArrayContainer) Release() {
	a.array.Release()
}
