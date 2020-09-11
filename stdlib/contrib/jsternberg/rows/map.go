package rows

import (
	"context"

	"github.com/apache/arrow/go/arrow/array"
	"github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/compiler"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/internal/execute/table"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/values"
)

const pkgpath = "contrib/jsternberg/rows"

const MapKind = pkgpath + ".map"

func init() {
	runtime.RegisterPackageValue(pkgpath, "map", flux.MustValue(flux.FunctionValue(
		"map",
		createMapOpSpec,
		runtime.MustLookupBuiltinType(pkgpath, "map"),
	)))
	plan.RegisterProcedureSpec(MapKind, newMapProcedure, MapKind)
	execute.RegisterTransformation(MapKind, createMapTransformation)
}

type MapOpSpec struct {
	Fn interpreter.ResolvedFunction
}

func createMapOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(MapOpSpec)

	if fn, err := args.GetRequiredFunction("fn"); err != nil {
		return nil, err
	} else if fn, err := interpreter.ResolveFunction(fn); err != nil {
		return nil, err
	} else {
		spec.Fn = fn
	}

	return spec, nil
}

func (a *MapOpSpec) Kind() flux.OperationKind {
	return MapKind
}

type MapProcedureSpec struct {
	plan.DefaultCost
	Fn interpreter.ResolvedFunction
}

func newMapProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*MapOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
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
	ns.Fn = s.Fn.Copy()
	return ns
}

func createMapTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*MapProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}
	return NewMapTransformation(a.Context(), s, id, a.Allocator())
}

type mapTransformation struct {
	d   *execute.PassthroughDataset
	ctx context.Context
	fn  *execute.RowMapFn
	mem memory.Allocator
}

func NewMapTransformation(ctx context.Context, spec *MapProcedureSpec, id execute.DatasetID, mem memory.Allocator) (execute.Transformation, execute.Dataset, error) {
	fn := execute.NewRowMapFn(spec.Fn.Fn, compiler.ToScope(spec.Fn.Scope))
	t := &mapTransformation{
		d:   execute.NewPassthroughDataset(id),
		ctx: ctx,
		fn:  fn,
		mem: mem,
	}
	return t, t.d, nil
}

func (t *mapTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *mapTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	// Prepare the function for the column types.
	cols := tbl.Cols()
	fn, err := t.fn.Prepare(cols)
	if err != nil {
		return err
	}
	return t.yield(t.mapTable(tbl, fn))
}

func (t *mapTransformation) createSchema(in flux.Table, fn *execute.RowMapPreparedFn) ([]flux.ColMeta, error) {
	mt := fn.Type()
	nproperties, err := mt.NumProperties()
	if err != nil {
		return nil, err
	}

	key := in.Key()
	cols := make([]flux.ColMeta, 0, nproperties+len(key.Cols()))

	for i := 0; i < nproperties; i++ {
		prop, err := mt.RecordProperty(i)
		if err != nil {
			return nil, err
		}
		name := prop.Name()
		if execute.ColIdx(name, cols) >= 0 {
			// A column with this name was already added so
			// this was overwritten.
			continue
		}
		typ, err := prop.TypeOf()
		if err != nil {
			return nil, err
		}

		colType := flux.ColumnType(typ)
		if colType == flux.TInvalid {
			return nil, errors.Newf(codes.FailedPrecondition, "output column %q is an invalid type; check that all inputs exist on all series", name)
		}
		cols = append(cols, flux.ColMeta{
			Label: name,
			Type:  colType,
		})
	}

	for _, c := range key.Cols() {
		if execute.ColIdx(c.Label, cols) >= 0 {
			continue
		}
		cols = append(cols, c)
	}
	return cols, nil
}

func (t *mapTransformation) mapTable(in flux.Table, fn *execute.RowMapPreparedFn) (flux.Table, error) {
	cols, err := t.createSchema(in, fn)
	if err != nil {
		return nil, err
	}

	return &mapTable{
		Table: in,
		ctx:   t.ctx,
		cols:  cols,
		fn:    fn,
		mem:   t.mem,
	}, nil
}

func (t *mapTransformation) yield(tbl flux.Table, err error) error {
	if err != nil {
		return err
	}
	return t.d.Process(tbl)
}

func (t *mapTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}

func (t *mapTransformation) UpdateProcessingTime(id execute.DatasetID, ts execute.Time) error {
	return t.d.UpdateProcessingTime(ts)
}

func (t *mapTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}

type mapTable struct {
	flux.Table
	ctx  context.Context
	cols []flux.ColMeta
	fn   *execute.RowMapPreparedFn
	mem  memory.Allocator
}

func (t *mapTable) Cols() []flux.ColMeta {
	return t.cols
}

func (t *mapTable) Do(f func(cr flux.ColReader) error) error {
	return t.Table.Do(func(cr flux.ColReader) error {
		buffer := &arrow.TableBuffer{
			GroupKey: t.Table.Key(),
			Columns:  t.cols,
		}
		if err := t.mapValues(buffer, cr, t.fn); err != nil {
			return err
		}
		return f(buffer)
	})
}

func (t *mapTable) mapValues(w *arrow.TableBuffer, cr flux.ColReader, fn *execute.RowMapPreparedFn) error {
	cols := w.Cols()
	builders := make([]array.Builder, len(cols))
	for i, c := range cols {
		// If part of the group key, do not create
		// a builder. We are going to retain the original
		// column.
		if execute.ColIdx(c.Label, cr.Key().Cols()) >= 0 {
			continue
		}
		builders[i] = arrow.NewBuilder(c.Type, t.mem)
		builders[i].Resize(cr.Len())
	}

	// Iterate over each row.
	for i, n := 0, cr.Len(); i < n; i++ {
		obj, err := fn.Eval(t.ctx, i, cr)
		if err != nil {
			return err
		}

		// Append each of the returned values to the
		// appropriate column.
		obj.Range(func(name string, v values.Value) {
			if err != nil {
				return
			}

			idx := execute.ColIdx(name, w.Cols())
			if builders[idx] == nil {
				// Part of the group key. Discard the value.
				// Checking for equality is too expensive so we do not
				// attempt to verify that it wasn't changed and instead
				// opt to ignore changes.
				return
			}
			err = arrow.AppendValue(builders[idx], v)
		})

		if err != nil {
			return err
		}
	}

	w.Values = make([]array.Interface, len(cols))
	for i, c := range cols {
		if builders[i] == nil {
			idx := execute.ColIdx(c.Label, cr.Cols())
			w.Values[i] = table.Values(cr, idx)
			w.Values[i].Retain()
		} else {
			w.Values[i] = builders[i].NewArray()
		}
	}
	return nil
}
