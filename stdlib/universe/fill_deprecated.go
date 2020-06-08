package universe

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/plan"
)

func createDeprecatedFillTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*FillProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}
	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)
	t := NewDeprecatedFillTransformation(d, cache, s)
	return t, d, nil
}

type deprecatedFillTransformation struct {
	d     execute.Dataset
	cache execute.TableBuilderCache

	spec *FillProcedureSpec
}

func NewDeprecatedFillTransformation(d execute.Dataset, cache execute.TableBuilderCache, spec *FillProcedureSpec) *deprecatedFillTransformation {
	return &deprecatedFillTransformation{
		d:     d,
		cache: cache,
		spec:  spec,
	}
}

func (t *deprecatedFillTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *deprecatedFillTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	key := tbl.Key()
	if idx := execute.ColIdx(t.spec.Column, key.Cols()); idx >= 0 && key.IsNull(idx) {
		var err error
		gkb := execute.NewGroupKeyBuilder(key)

		gkb.SetKeyValue(t.spec.Column, t.spec.Value)
		key, err = gkb.Build()
		if err != nil {
			return err
		}
	}

	builder, created := t.cache.TableBuilder(key)
	if created {
		if err := execute.AddTableCols(tbl, builder); err != nil {
			return err
		}
	}
	idx := execute.ColIdx(t.spec.Column, builder.Cols())
	if idx < 0 {
		return errors.Newf(codes.FailedPrecondition, "fill column not found: %s", t.spec.Column)
	}

	prevNonNull := t.spec.Value
	if !t.spec.UsePrevious {
		if builder.Cols()[idx].Type != flux.ColumnType(prevNonNull.Type()) {
			return errors.Newf(codes.FailedPrecondition, "fill column type mismatch: %s/%s", builder.Cols()[idx].Type.String(), flux.ColumnType(prevNonNull.Type()).String())
		}
	}
	return tbl.Do(func(cr flux.ColReader) error {
		for j := range cr.Cols() {
			if j == idx {
				continue
			}
			if err := execute.AppendCol(j, j, cr, builder); err != nil {
				return err
			}
		}
		// Set new value
		l := cr.Len()

		if l > 0 {
			if t.spec.UsePrevious {
				prevNonNull = execute.ValueForRow(cr, 0, idx)
			}

			for i := 0; i < l; i++ {
				v := execute.ValueForRow(cr, i, idx)
				if v.IsNull() {
					if err := builder.AppendValue(idx, prevNonNull); err != nil {
						return err
					}
				} else {
					if err := builder.AppendValue(idx, v); err != nil {
						return err
					}
					if t.spec.UsePrevious {
						prevNonNull = v
					}
				}
			}
		}
		return nil
	})
}

func (t *deprecatedFillTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}
func (t *deprecatedFillTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}
func (t *deprecatedFillTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}
