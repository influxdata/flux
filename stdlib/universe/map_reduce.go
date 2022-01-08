package universe

import (
	"context"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/compiler"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

const MapReduceKind = "mapReduce"

type MapReduceOpSpec struct {
	Fn       interpreter.ResolvedFunction `json:"fn"`
	Identity values.Object                `json:"identity"`
	MergeKey bool                         `json:"mergeKey"`
}

func init() {
	mapReduceSignature := runtime.MustLookupBuiltinType("universe", "mapReduce")

	runtime.RegisterPackageValue("universe", MapReduceKind, flux.MustValue(flux.FunctionValue(MapReduceKind, createMapReduceOpSpec, mapReduceSignature)))
	flux.RegisterOpSpec(MapReduceKind, newMapReduceOp)
	plan.RegisterProcedureSpec(MapReduceKind, newMapReduceProcedure, MapReduceKind)
	execute.RegisterTransformation(MapReduceKind, createMapReduceTransformation)
}

func createMapReduceOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(MapReduceOpSpec)

	if f, err := args.GetRequiredFunction("fn"); err != nil {
		return nil, err
	} else {
		fn, err := interpreter.ResolveFunction(f)
		if err != nil {
			return nil, err
		}
		spec.Fn = fn
	}

	if o, err := args.GetRequiredObject("identity"); err != nil {
		return nil, err
	} else {
		spec.Identity = o
	}

	if m, ok, err := args.GetBool("mergeKey"); err != nil {
		return nil, err
	} else if ok {
		spec.MergeKey = m
	} else {
		// deprecated parameter: default is now false.
		spec.MergeKey = false
	}
	return spec, nil
}

func newMapReduceOp() flux.OperationSpec {
	return new(MapReduceOpSpec)
}

func (s *MapReduceOpSpec) Kind() flux.OperationKind {
	return MapReduceKind
}

type MapReduceProcedureSpec struct {
	plan.DefaultCost
	Fn       interpreter.ResolvedFunction `json:"fn"`
	Identity values.Object
	MergeKey bool
}

func newMapReduceProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*MapReduceOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}

	return &MapReduceProcedureSpec{
		Fn:       spec.Fn,
		MergeKey: spec.MergeKey,
	}, nil
}

func (s *MapReduceProcedureSpec) Kind() plan.ProcedureKind {
	return MapReduceKind
}
func (s *MapReduceProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(MapReduceProcedureSpec)
	*ns = *s
	ns.Fn = s.Fn.Copy()
	return ns
}

func createMapReduceTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*MapReduceProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}
	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)
	t, err := NewMapReduceTransformation(a.Context(), s, d, cache)

	if err != nil {
		return nil, nil, err
	}
	return t, d, nil
}

type mapReduceTransformation struct {
	mapTransformation
	fn       *execute.RowMapReduceFn
	identity values.Object
}

func NewMapReduceTransformation(ctx context.Context, spec *MapReduceProcedureSpec, d execute.Dataset, cache execute.TableBuilderCache) (*mapReduceTransformation, error) {
	fn := execute.NewRowMapReduceFn(spec.Fn.Fn, compiler.ToScope(spec.Fn.Scope))
	return &mapReduceTransformation{
		mapTransformation: mapTransformation{
			d:        d,
			cache:    cache,
			ctx:      ctx,
			mergeKey: spec.MergeKey,
		},
		fn:       fn,
		identity: spec.Identity,
	}, nil
}

func (t *mapReduceTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	// Prepare the functions for the column types.
	cols := tbl.Cols()
	fn, err := t.fn.Prepare(cols, map[string]semantic.MonoType{"accumulator": t.identity.Type()})
	if err != nil {
		return err
	}

	const accumulatorParamName = "accumulator"
	const rowPropertyName = "row"
	params := map[string]values.Value{accumulatorParamName: t.identity}

	var on map[string]bool
	err = tbl.Do(func(cr flux.ColReader) error {
		l := cr.Len()
		for i := 0; i < l; i++ {
			mm, err := fn.Eval(t.ctx, i, cr, params)
			if err != nil {
				return errors.Wrap(err, codes.Invalid, "failed to evaluate mapReduce function")
			}

			// reduce
			accumulator, ok := mm.Get(accumulatorParamName)
			if !ok {
				return errors.Newf(codes.Invalid, "failed to retrieve property %s from the mapReduce functions return value", accumulatorParamName)
			}
			params[accumulatorParamName] = accumulator

			// map
			mValue, ok := mm.Get(rowPropertyName)
			if !ok {
				return errors.Newf(codes.Invalid, "failed to retrieve property %s from the mapReduce function's return value", rowPropertyName)
			}
			m, ok := mValue.(values.Object)
			if !ok {
				return errors.Newf(codes.Invalid, "property %s of the mapReduce function's return value is not an object", rowPropertyName)
			}

			// geropl: not sure what the code below does: took it from map.go:mapTransformation.Process.
			// 		Maybe collides with reduce-code below?

			// If we haven't determined the columns to group on, do that now.
			if on == nil {
				var err error
				on, err = t.groupOn(tbl.Key(), m.Type())
				if err != nil {
					return err
				}
			}

			key := groupKeyForObject(i, cr, m, on)
			builder, created := t.cache.TableBuilder(key)
			if created {
				if err := t.createSchema(m.Type(), builder, m); err != nil {
					return err
				}
			}

			for j, c := range builder.Cols() {
				v, ok := m.Get(c.Label)
				if !ok {
					if idx := execute.ColIdx(c.Label, tbl.Key().Cols()); t.mergeKey && idx >= 0 {
						v = tbl.Key().Value(idx)
					} else {
						// This should be unreachable
						return errors.Newf(codes.Internal, "could not find value for column %q", c.Label)
					}
				}
				if !v.IsNull() && c.Type.String() != v.Type().Nature().String() {
					return errors.Newf(codes.Invalid, "map regroups data such that column %q would include values"+
						" of two different data types: %v, %v",
						c.Label, c.Type, v.Type(),
					)
				}
				if err := builder.AppendValue(j, v); err != nil {
					return err
				}
			}
		}
		return nil
	})

	return err
}
