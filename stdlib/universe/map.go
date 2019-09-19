package universe

import (
	"context"
	"sort"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/compiler"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

const MapKind = "map"

type MapOpSpec struct {
	Fn       interpreter.ResolvedFunction `json:"fn"`
	MergeKey bool                         `json:"mergeKey"`
}

func init() {
	mapSignature := flux.FunctionSignature(
		map[string]semantic.PolyType{
			"fn": semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
				Parameters: map[string]semantic.PolyType{
					"r": semantic.Tvar(1),
				},
				Required: semantic.LabelSet{"r"},
				Return:   semantic.Tvar(2),
			}),
			"mergeKey": semantic.Bool,
		},
		[]string{"fn"},
	)

	flux.RegisterPackageValue("universe", MapKind, flux.FunctionValue(MapKind, createMapOpSpec, mapSignature))
	flux.RegisterOpSpec(MapKind, newMapOp)
	plan.RegisterProcedureSpec(MapKind, newMapProcedure, MapKind)
	execute.RegisterTransformation(MapKind, createMapTransformation)
}

func createMapOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(MapOpSpec)

	if f, err := args.GetRequiredFunction("fn"); err != nil {
		return nil, err
	} else {
		fn, err := interpreter.ResolveFunction(f)
		if err != nil {
			return nil, err
		}
		spec.Fn = fn
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

func newMapOp() flux.OperationSpec {
	return new(MapOpSpec)
}

func (s *MapOpSpec) Kind() flux.OperationKind {
	return MapKind
}

type MapProcedureSpec struct {
	plan.DefaultCost
	Fn       interpreter.ResolvedFunction `json:"fn"`
	MergeKey bool
}

func newMapProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*MapOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}

	return &MapProcedureSpec{
		Fn:       spec.Fn,
		MergeKey: spec.MergeKey,
	}, nil
}

func (s *MapProcedureSpec) Kind() plan.ProcedureKind {
	return MapKind
}
func (s *MapProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(MapProcedureSpec)
	*ns = *s
	ns.Fn = s.Fn.Copy()
	return ns
}

func createMapTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*MapProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}
	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)
	t, err := NewMapTransformation(a.Context(), s, d, cache)

	if err != nil {
		return nil, nil, err
	}
	return t, d, nil
}

type mapTransformation struct {
	d        execute.Dataset
	cache    execute.TableBuilderCache
	ctx      context.Context
	fn       *execute.RowMapFn
	mergeKey bool
}

func NewMapTransformation(ctx context.Context, spec *MapProcedureSpec, d execute.Dataset, cache execute.TableBuilderCache) (*mapTransformation, error) {
	fn, err := execute.NewRowMapFn(spec.Fn.Fn, compiler.ToScope(spec.Fn.Scope))
	if err != nil {
		return nil, err
	}
	return &mapTransformation{
		d:        d,
		cache:    cache,
		fn:       fn,
		ctx:      ctx,
		mergeKey: spec.MergeKey,
	}, nil
}

func (t *mapTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *mapTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	// Prepare the functions for the column types.
	cols := tbl.Cols()
	if err := t.fn.Prepare(cols); err != nil {
		// TODO(nathanielc): Should we not fail the query for failed compilation?
		return err
	}

	// Determine the properties from type inference.
	// TODO(jsternberg): Type inference doesn't currently get all of the properties
	// when with is used, so this will likely return an incomplete set that we have
	// to complete when we read the first column. This pass should get the columns
	// that are directly referenced which are the ones capable of changing and the
	// only ones that can be a null value.
	// TODO (adam): these are a pointer to a property list that is definitely shared by the type checking code (semantic/types.go NewObjectType)
	unsafeProperties := t.fn.Type().Properties()
	properties := make(map[string]semantic.Type)

	var on map[string]bool
	return tbl.Do(func(cr flux.ColReader) error {
		l := cr.Len()
		for i := 0; i < l; i++ {
			m, err := t.fn.Eval(t.ctx, i, cr)
			if err != nil {
				return errors.Wrap(err, codes.Inherit, "failed to evaluate map function")
			}

			// Merge in the types that we may have missed because type inference omitted them.
			if i == 0 {
				// Merge in the missing properties from the type.
				// This will catch all of the non-referenced keys that are
				// merged in from the merge key or with.
				for k, typ := range m.Type().Properties() {
					// TODO(jsternberg): Type inference can sometimes tell us something is nil
					// when it actually has a value because of a bug in type inference.
					// If the property is missing or null, then take whatever the first value
					// is. This won't catch all situations, but we'll have to consider it good
					// enough until type inference works in this scenario.
					if t, ok := unsafeProperties[k]; !ok || t == semantic.Nil {
						properties[k] = typ
					} else {
						properties[k] = unsafeProperties[k]
					}
				}
			}

			// If we haven't determined the columns to group on, do that now.
			if on == nil {
				on = make(map[string]bool, len(tbl.Key().Cols()))
				for _, c := range tbl.Key().Cols() {
					if !t.mergeKey {
						// If the label isn't included in the properties,
						// then it wasn't returned by the eval.
						if _, ok := properties[c.Label]; !ok {
							continue
						}
					}
					on[c.Label] = true
				}
			}

			key := groupKeyForObject(i, cr, m, on)
			builder, created := t.cache.TableBuilder(key)
			if created {
				if t.mergeKey {
					if err := execute.AddTableKeyCols(tbl.Key(), builder); err != nil {
						return err
					}
				}

				// Add columns from function in sorted order.
				keys := make([]string, 0, len(properties))
				for k := range properties {
					keys = append(keys, k)
				}
				sort.Strings(keys)

				for _, k := range keys {
					if t.mergeKey && tbl.Key().HasCol(k) {
						continue
					}

					n := properties[k].Nature()
					if n == semantic.Nil {
						// If the column is null, then do not add it as a column.
						continue
					}

					if _, err := builder.AddCol(flux.ColMeta{
						Label: k,
						Type:  execute.ConvertFromKind(n),
					}); err != nil {
						return err
					}
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
				if err := builder.AppendValue(j, v); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func groupKeyForObject(i int, cr flux.ColReader, obj values.Object, on map[string]bool) flux.GroupKey {
	cols := make([]flux.ColMeta, 0, len(on))
	vs := make([]values.Value, 0, len(on))
	for j, c := range cr.Cols() {
		if !on[c.Label] {
			continue
		}
		cols = append(cols, c)
		v, ok := obj.Get(c.Label)
		if ok {
			vs = append(vs, v)
		} else {
			vs = append(vs, execute.ValueForRow(cr, i, j))
		}
	}
	return execute.NewGroupKey(cols, vs)
}

func (t *mapTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}
func (t *mapTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}
func (t *mapTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}
