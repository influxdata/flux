package universe

import (
	"context"
	"fmt"
	"sort"

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

const MapKind = "map"

type MapOpSpec struct {
	Fn       interpreter.ResolvedFunction `json:"fn"`
	MergeKey bool                         `json:"mergeKey"`
}

func init() {
	mapSignature := runtime.MustLookupBuiltinType("universe", "map")

	runtime.RegisterPackageValue("universe", MapKind, flux.MustValue(flux.FunctionValue(MapKind, createMapOpSpec, mapSignature)))
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
	fn := execute.NewRowMapFn(spec.Fn.Fn, compiler.ToScope(spec.Fn.Scope))
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
	fn, err := t.fn.Prepare(cols)
	if err != nil {
		// TODO(nathanielc): Should we not fail the query for failed compilation?
		return err
	}

	// TODO(jsternberg): Use type inference. The original algorithm
	// didn't really use type inference at all so I removed its usage
	// in favor of the real returned type.

	var on map[string]bool
	return tbl.Do(func(cr flux.ColReader) error {
		l := cr.Len()
		for i := 0; i < l; i++ {
			m, err := fn.Eval(t.ctx, i, cr)
			if err != nil {
				return errors.Wrap(err, codes.Inherit, "failed to evaluate map function")
			}

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
				if err := t.createSchema(fn, builder, m); err != nil {
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
				if err := builder.AppendValue(j, v); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (t *mapTransformation) groupOn(key flux.GroupKey, m semantic.MonoType) (map[string]bool, error) {
	on := make(map[string]bool, len(key.Cols()))
	for _, c := range key.Cols() {
		if t.mergeKey {
			on[c.Label] = true
			continue
		}

		// If the label isn't included in the properties,
		// then it wasn't returned by the eval.
		n, err := m.NumProperties()
		if err != nil {
			return nil, err
		}

		for i := 0; i < n; i++ {
			Record, err := m.RecordProperty(i)
			if err != nil {
				return nil, err
			}

			if Record.Name() == c.Label {
				on[c.Label] = true
				break
			}
		}
	}
	return on, nil
}

// createSchema will create the schema for a table based on the object.
// This should only be called when a table is created anew.
//
// TODO(jsternberg): I am pretty sure this method and its usage don't
// match with the spec, but it is a faithful reproduction of the current
// map behavior. When we get around to rewriting portions of map, this
// should be rewritten to use the inferred type from type inference
// and it should be capable of consolidating schemas from non-uniform
// tables.
func (t *mapTransformation) createSchema(fn *execute.RowMapPreparedFn, b execute.TableBuilder, m values.Object) error {
	if t.mergeKey {
		if err := execute.AddTableKeyCols(b.Key(), b); err != nil {
			return err
		}
	}

	returnType := fn.Type()

	numProps, err := returnType.NumProperties()
	if err != nil {
		return err
	}

	props := make(map[string]semantic.Nature, numProps)
	// Deduplicate the properties in the return type.
	// Scan properties in reverse order to ensure we only
	// add visible properties to the list.
	for i := numProps - 1; i >= 0; i-- {
		prop, err := returnType.RecordProperty(i)
		if err != nil {
			return err
		}
		typ, err := prop.TypeOf()
		if err != nil {
			return err
		}
		props[prop.Name()] = typ.Nature()
	}

	// Add columns from function in sorted order.
	n, err := m.Type().NumProperties()
	if err != nil {
		return err
	}

	keys := make([]string, 0, n)
	for i := 0; i < n; i++ {
		Record, err := m.Type().RecordProperty(i)
		if err != nil {
			return err
		}
		keys = append(keys, Record.Name())
	}
	sort.Strings(keys)

	for _, k := range keys {
		if t.mergeKey && b.Key().HasCol(k) {
			continue
		}

		v, ok := m.Get(k)
		if !ok {
			continue
		}

		nature := v.Type().Nature()

		if kind, ok := props[k]; ok && kind != semantic.Invalid {
			nature = kind
		}
		if nature == semantic.Invalid {
			continue
		}
		ty := execute.ConvertFromKind(nature)
		if ty == flux.TInvalid {
			return fmt.Errorf(`map object property "%s" is %v type which is not supported in a flux table`, k, nature)
		}
		if _, err := b.AddCol(flux.ColMeta{
			Label: k,
			Type:  ty,
		}); err != nil {
			return err
		}
	}
	return nil
}

func groupKeyForObject(i int, cr flux.ColReader, obj values.Object, on map[string]bool) flux.GroupKey {
	cols := make([]flux.ColMeta, 0, len(on))
	vs := make([]values.Value, 0, len(on))
	for j, c := range cr.Cols() {
		if !on[c.Label] {
			continue
		}
		v, ok := obj.Get(c.Label)
		if ok {
			vs = append(vs, v)
			cols = append(cols, flux.ColMeta{
				Label: c.Label,
				Type:  flux.ColumnType(v.Type()),
			})
		} else {
			vs = append(vs, execute.ValueForRow(cr, i, j))
			cols = append(cols, c)
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
