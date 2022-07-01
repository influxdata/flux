package moving_average

import (
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/values"
)

type ArrayContainer struct {
	array array.Array
}

func NewArrayContainer(a array.Array) *ArrayContainer {
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
	case *array.Int:
		return values.New(float64(a.array.(*array.Int).Value(i)))
	case *array.Uint:
		return values.New(float64(a.array.(*array.Uint).Value(i)))
	case *array.Float:
		return values.New(float64(a.array.(*array.Float).Value(i)))
	case *array.String:
		return values.New(string(a.array.(*array.String).Value(i)))
	default:
		return nil
	}
}

func (a *ArrayContainer) OrigValue(i int) interface{} {
	switch a.array.(type) {
	case *array.Boolean:
		return a.array.(*array.Boolean).Value(i)
	case *array.Int:
		return a.array.(*array.Int).Value(i)
	case *array.Uint:
		return a.array.(*array.Uint).Value(i)
	case *array.Float:
		return a.array.(*array.Float).Value(i)
	case *array.String:
		return string(a.array.(*array.String).Value(i))
	default:
		return nil
	}
}

func (a *ArrayContainer) Slice(i int, j int) *ArrayContainer {
	slice := &ArrayContainer{}
	switch a.array.(type) {
	case *array.Boolean:
		slice.array = arrow.BoolSlice(a.array.(*array.Boolean), i, j)
	case *array.Int:
		slice.array = arrow.IntSlice(a.array.(*array.Int), i, j)
	case *array.Uint:
		slice.array = arrow.UintSlice(a.array.(*array.Uint), i, j)
	case *array.Float:
		slice.array = arrow.FloatSlice(a.array.(*array.Float), i, j)
	case *array.String:
		slice.array = arrow.StringSlice(a.array.(*array.String), i, j)
	default:
		slice.array = nil
	}
	return slice
}

func (a *ArrayContainer) Array() array.Array {
	return a.array
}

func (a *ArrayContainer) Release() {
	a.array.Release()
}
