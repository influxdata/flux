package execute

import (
	"context"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/compiler"
	"github.com/influxdata/flux/execute/table"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

type VectorMapFn struct {
	dynamicFn
}

func NewVectorMapFn(fn *semantic.FunctionExpression, scope compiler.Scope) *VectorMapFn {
	return &VectorMapFn{
		dynamicFn: newDynamicFn(fn, scope),
	}
}

func (f *VectorMapFn) Prepare(cols []flux.ColMeta) (*VectorMapPreparedFn, error) {
	fn, err := f.prepare(cols, nil, true)
	if err != nil {
		return nil, err
	} else if k := fn.returnType().Nature(); k != semantic.Object {
		return nil, errors.Newf(codes.Invalid, "map function must return an object, got %s", k.String())
	}
	return &VectorMapPreparedFn{
		vectorFn: vectorFn{preparedFn: fn},
	}, nil
}

type VectorMapPreparedFn struct {
	vectorFn
}

func (f *VectorMapPreparedFn) Type() semantic.MonoType {
	return f.fn.Type()
}

type vectorFn struct {
	preparedFn
}

func (f *vectorFn) Eval(ctx context.Context, chunk table.Chunk) ([]array.Interface, error) {
	for j, col := range chunk.Cols() {
		arr := chunk.Values(j)
		arr.Retain()
		v := values.NewVectorValue(arr, flux.SemanticType(col.Type))
		f.arg0.Set(col.Label, v)
	}

	res, err := f.fn.Eval(ctx, f.args)
	if err != nil {
		return nil, err
	}

	// Map the return object to the expected order from type inference.
	// The compiler should have done this by itself, but it doesn't at the moment.
	// When the compiler gets refactored so it returns records in the same order
	// as type inference, we can remove this and just do a copy by index.
	retType := f.returnType()
	n := res.Object().Len()
	vs := make([]array.Interface, n)
	for i := 0; i < n; i++ {
		prop, err := retType.RecordProperty(i)
		if err != nil {
			return nil, err
		}

		vec, ok := res.Object().Get(prop.Name())
		if !ok {
			return nil, errors.Newf(codes.Internal, "column %s is not valid", prop.Name())
		}
		vs[i] = vec.(values.Vector).Arr()
		vs[i].Retain()
	}
	res.Release()
	return vs, nil
}
