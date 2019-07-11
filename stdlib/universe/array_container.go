package universe

import (
	"github.com/apache/arrow/go/arrow/array"
	"github.com/influxdata/flux/values"
)

type arrayContainer struct {
	array array.Interface
}

func (a *arrayContainer) IsNull(i int) bool {
	return a.array.IsNull(i)
}

func (a *arrayContainer) IsValid(i int) bool {
	return a.array.IsValid(i)
}

func (a *arrayContainer) Len() int {
	return a.array.Len()
}

func (a *arrayContainer) Value(i int) values.Value {
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

func (a *arrayContainer) OrigValue(i int) interface{} {
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
