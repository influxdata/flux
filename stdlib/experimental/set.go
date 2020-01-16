package experimental

import (
	"fmt"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

const SetKind = "setExperimental"

type SetOpSpec struct {
	Object values.Object `json:"object"`
}

func init() {
	setSignature := semantic.MustLookupBuiltinType("experimental", "set")

	flux.RegisterPackageValue("experimental", "set", flux.MustValue(flux.FunctionValue(SetKind, createSetOpSpec, setSignature)))
	flux.RegisterOpSpec(SetKind, newSetOp)
	plan.RegisterProcedureSpec(SetKind, newSetProcedure, SetKind)
	execute.RegisterTransformation(SetKind, createSetTransformation)
}

func createSetOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(SetOpSpec)
	o, err := args.GetRequiredObject("o")
	if err != nil {
		return nil, err
	}
	spec.Object = o
	return spec, nil
}

func newSetOp() flux.OperationSpec {
	return new(SetOpSpec)
}

func (s *SetOpSpec) Kind() flux.OperationKind {
	return SetKind
}

type SetProcedureSpec struct {
	plan.DefaultCost
	Object values.Object
}

func newSetProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	s, ok := qs.(*SetOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}
	p := &SetProcedureSpec{
		Object: s.Object,
	}
	return p, nil
}

func (s *SetProcedureSpec) Kind() plan.ProcedureKind {
	return SetKind
}
func (s *SetProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(SetProcedureSpec)
	ns.Object = values.NewObject(s.Object.Type())
	s.Object.Range(func(k string, v values.Value) {
		ns.Object.Set(k, v)
	})
	return ns
}

func createSetTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*SetProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", spec)
	}
	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)
	t := NewSetTransformation(d, cache, s)
	return t, d, nil
}

type setTransformation struct {
	d     execute.Dataset
	cache execute.TableBuilderCache

	object values.Object
}

func NewSetTransformation(
	d execute.Dataset,
	cache execute.TableBuilderCache,
	spec *SetProcedureSpec,
) execute.Transformation {
	return &setTransformation{
		d:      d,
		cache:  cache,
		object: spec.Object,
	}
}

func (t *setTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	// TODO
	return nil
}

func (t *setTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	key := tbl.Key()
	updateKey := false
	t.object.Range(func(k string, _ values.Value) {
		updateKey = updateKey || execute.HasCol(k, key.Cols())
	})
	if updateKey {
		// Update key
		cols := make([]flux.ColMeta, len(key.Cols()))
		vs := make([]values.Value, len(key.Cols()))
		for j, c := range key.Cols() {
			cols[j] = c
			vs[j] = key.Value(j)
			if v, ok := t.object.Get(c.Label); ok {
				cols[j] = flux.ColMeta{
					Label: c.Label,
					Type:  flux.ColumnType(v.Type()),
				}
				vs[j] = v
			}
		}
		key = execute.NewGroupKey(cols, vs)
	}
	var colMap []int
	builder, created := t.cache.TableBuilder(key)
	if created {
		// Add existing columns from input
		for _, c := range tbl.Cols() {
			if v, ok := t.object.Get(c.Label); !ok {
				builder.AddCol(c)
			} else {
				if _, err := builder.AddCol(flux.ColMeta{
					Label: c.Label,
					Type:  flux.ColumnType(v.Type()),
				}); err != nil {
					return err
				}
			}
		}

		// Add new columns from object
		var rangeErr error
		t.object.Range(func(k string, v values.Value) {
			if rangeErr != nil {
				return
			}
			if !execute.HasCol(k, builder.Cols()) {
				if _, err := builder.AddCol(flux.ColMeta{
					Label: k,
					Type:  flux.ColumnType(v.Type()),
				}); err != nil {
					rangeErr = err
				}
			}
		})
		if rangeErr != nil {
			return rangeErr
		}
	}
	colMap = execute.ColMap(colMap, builder, tbl.Cols())

	return tbl.Do(func(cr flux.ColReader) error {
		for j, c := range builder.Cols() {
			v, ok := t.object.Get(c.Label)
			if !ok {
				// copy column from input to output
				cj := colMap[j]
				if err := execute.AppendCol(j, cj, cr, builder); err != nil {
					return err
				}
			} else {
				// set new value on output
				l := cr.Len()
				for i := 0; i < l; i++ {
					if err := builder.AppendValue(j, v); err != nil {
						return err
					}
				}
			}
		}
		return nil
	})
}

func (t *setTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}
func (t *setTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}
func (t *setTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}
