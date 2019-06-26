package universe

import (
	"fmt"
	"sort"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
	"github.com/pkg/errors"
)

const MapKind = "map"

type MapOpSpec struct {
	Fn *semantic.FunctionExpression `json:"fn"`
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
	Fn *semantic.FunctionExpression
}

func newMapProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*MapOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	return &MapProcedureSpec{
		Fn: spec.Fn,
	}, nil
}

func (s *MapProcedureSpec) Kind() plan.ProcedureKind {
	return MapKind
}
func (s *MapProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(MapProcedureSpec)
	*ns = *s
	ns.Fn = s.Fn.Copy().(*semantic.FunctionExpression)
	return ns
}

func createMapTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*MapProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", spec)
	}
	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)
	t, err := NewMapTransformation(d, cache, s)
	if err != nil {
		return nil, nil, err
	}
	return t, d, nil
}

type mapTransformation struct {
	d     execute.Dataset
	cache execute.TableBuilderCache

	fn *execute.RowMapFn
}

func NewMapTransformation(d execute.Dataset, cache execute.TableBuilderCache, spec *MapProcedureSpec) (*mapTransformation, error) {
	fn, err := execute.NewRowMapFn(spec.Fn)
	if err != nil {
		return nil, err
	}
	return &mapTransformation{
		d:     d,
		cache: cache,
		fn:    fn,
	}, nil
}

func (t *mapTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *mapTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	// Prepare the functions for the column types.
	cols := tbl.Cols()
	err := t.fn.Prepare(cols)
	if err != nil {
		// TODO(nathanielc): Should we not fail the query for failed compilation?
		return err
	}
	// Determine keys return from function
	var properties map[string]semantic.Type
	var keys []string
	var on map[string]bool

	return tbl.Do(func(cr flux.ColReader) error {
		l := cr.Len()
		for i := 0; i < l; i++ {
			m, err := t.fn.Eval(i, cr)
			if err != nil {
				return errors.Wrap(err, "failed to evaluate map function")
			}
			if i == 0 {
				properties = m.Type().Properties()
				keys = make([]string, 0, len(properties))
				for k := range properties {
					keys = append(keys, k)
				}
				sort.Strings(keys)

				// Determine on which cols to group
				on = make(map[string]bool, len(tbl.Key().Cols()))
				for _, c := range tbl.Key().Cols() {
					on[c.Label] = execute.ContainsStr(keys, c.Label)
				}
			}

			key := groupKeyForObject(i, cr, m, on)
			builder, created := t.cache.TableBuilder(key)
			if created {
				// Add columns from function in sorted order
				for _, k := range keys {
					if _, err := builder.AddCol(flux.ColMeta{
						Label: k,
						Type:  execute.ConvertFromKind(properties[k].Nature()),
					}); err != nil {
						return err
					}
				}
			}
			for j, c := range builder.Cols() {
				v, ok := m.Get(c.Label)
				if !ok {
					// This should be unreachable
					return fmt.Errorf("could not find value for column %q", c.Label)
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
