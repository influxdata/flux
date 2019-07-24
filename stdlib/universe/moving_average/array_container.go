package moving_average

import (
	"github.com/apache/arrow/go/arrow/array"
	"github.com/influxdata/flux/values"
)

type ArrayContainer struct {
	Array array.Interface
}

func (a *ArrayContainer) IsNull(i int) bool {
	return a.Array.IsNull(i)
}

func (a *ArrayContainer) IsValid(i int) bool {
	return a.Array.IsValid(i)
}

func (a *ArrayContainer) Len() int {
	return a.Array.Len()
}

func (a *ArrayContainer) Value(i int) values.Value {
	switch a.Array.(type) {
	case *array.Boolean:
		return values.New(a.Array.(*array.Boolean).Value(i))
	case *array.Int64:
		return values.New(float64(a.Array.(*array.Int64).Value(i)))
	case *array.Uint64:
		return values.New(float64(a.Array.(*array.Uint64).Value(i)))
	case *array.Float64:
		return values.New(float64(a.Array.(*array.Float64).Value(i)))
	case *array.Binary:
		return values.New(string(a.Array.(*array.Binary).Value(i)))
	default:
		return nil
	}
}

func (a *ArrayContainer) OrigValue(i int) interface{} {
	switch a.Array.(type) {
	case *array.Boolean:
		return a.Array.(*array.Boolean).Value(i)
	case *array.Int64:
		return a.Array.(*array.Int64).Value(i)
	case *array.Uint64:
		return a.Array.(*array.Uint64).Value(i)
	case *array.Float64:
		return a.Array.(*array.Float64).Value(i)
	case *array.Binary:
		return string(a.Array.(*array.Binary).Value(i))
	default:
		return nil
	}
}
