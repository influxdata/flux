package experimental

import (
	"github.com/apache/arrow/go/v7/arrow/memory"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/internal/execute/groupkey"
	"github.com/influxdata/flux/internal/execute/table"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/stdlib/influxdata/influxdb"
	"github.com/influxdata/flux/values"
)

const UnpivotKind = "experimental.unpivot"

type UnpivotOpSpec struct {
	otherColumns []string
}

func init() {
	unpivotSig := runtime.MustLookupBuiltinType("experimental", "unpivot")

	runtime.RegisterPackageValue("experimental", "unpivot", flux.MustValue(flux.FunctionValue(UnpivotKind, createUnpivotOpSpec, unpivotSig)))
	plan.RegisterProcedureSpec(UnpivotKind, newUnpivotProcedure, UnpivotKind)
	execute.RegisterTransformation(UnpivotKind, createUnpivotTransformation)
}

func createUnpivotOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(UnpivotOpSpec)

	if columns, ok, err := args.GetArray("otherColumns", semantic.String); err != nil {
		return nil, err
	} else if ok {
		spec.otherColumns, err = interpreter.ToStringArray(columns)
		if err != nil {
			return nil, err
		}
	} else {
		spec.otherColumns = []string{execute.DefaultTimeColLabel}
	}

	return spec, nil
}

func (s *UnpivotOpSpec) Kind() flux.OperationKind {
	return UnpivotKind
}

func newUnpivotProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	opSpec, ok := qs.(*UnpivotOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}

	return &UnpivotProcedureSpec{
		OtherColumns: opSpec.otherColumns,
	}, nil
}

func createUnpivotTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {

	s, ok := spec.(*UnpivotProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}

	return NewUnpivotTransformation(s, id, a.Allocator())
}

type UnpivotProcedureSpec struct {
	plan.DefaultCost
	OtherColumns []string
}

func (s *UnpivotProcedureSpec) Kind() plan.ProcedureKind {
	return UnpivotKind
}
func (s *UnpivotProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(UnpivotProcedureSpec)
	*ns = *s
	return ns
}

func NewUnpivotTransformation(spec *UnpivotProcedureSpec, id execute.DatasetID, alloc memory.Allocator) (execute.Transformation, execute.Dataset, error) {
	t := &unpivotTransformation{
		otherColumns: spec.OtherColumns,
	}
	return execute.NewNarrowTransformation(id, t, alloc)

}

type unpivotTransformation struct {
	execute.ExecutionNode
	otherColumns []string
}

func (t *unpivotTransformation) Close() error { return nil }

func (t *unpivotTransformation) Process(chunk table.Chunk, d *execute.TransportDataset, mem memory.Allocator) error {
	otherCols := make([]flux.ColMeta, len(t.otherColumns))
	for i, utc := range t.otherColumns {
		found := false
		for _, c := range chunk.Cols() {
			if c.Label == utc {
				otherCols[i] = c
				found = true
				break
			}
		}
		if !found {
			return errors.Newf(codes.Invalid, "unpivot could not find column named %q", utc)
		}
	}

	for i, c := range chunk.Cols() {
		if chunk.Key().HasCol(c.Label) || execute.HasCol(c.Label, otherCols) {
			continue
		}

		chunkValues := chunk.Values(i)
		chunkValues.Retain()

		newChunkLen := chunk.Len() - chunkValues.NullN()

		// The schema of the output table will be
		//
		// gk_col_0
		// gk_col_1
		// ...
		// gk_col_n
		// _field
		// other_col_0 (one of these may be _time)
		// other_col_1
		// ...
		// other_col_n
		// _value

		defaultValueColLen := 1
		defaultFieldColLen := 1
		if execute.HasCol(influxdb.DefaultFieldColLabel, otherCols) {
			defaultFieldColLen = 0
		}

		groupKey := chunk.Key()
		nCols := len(groupKey.Cols()) + len(otherCols) + defaultValueColLen + defaultFieldColLen

		columns := make([]flux.ColMeta, 0, nCols)
		columns = append(columns, groupKey.Cols()...)
		if defaultFieldColLen != 0 {
			columns = append(columns, flux.ColMeta{Label: influxdb.DefaultFieldColLabel, Type: flux.TString})
		}
		columns = append(columns, otherCols...)
		columns = append(columns, flux.ColMeta{Label: execute.DefaultValueColLabel, Type: c.Type})

		groupCols := make([]flux.ColMeta, 0, len(groupKey.Cols())+defaultFieldColLen)
		groupCols = append(groupCols, groupKey.Cols()...)
		if defaultFieldColLen != 0 {
			groupCols = append(groupCols, flux.ColMeta{Label: influxdb.DefaultFieldColLabel, Type: flux.TString})
		}

		groupValues := make([]values.Value, 0, len(groupKey.Cols())+defaultFieldColLen)
		groupValues = append(groupValues, groupKey.Values()...)
		if defaultFieldColLen != 0 {
			groupValues = append(groupValues, values.NewString(c.Label))
		}

		buffer := arrow.TableBuffer{
			GroupKey: groupkey.New(groupCols, groupValues),
			Columns:  columns,
			// we append to this below
			Values: make([]array.Array, 0, len(columns)),
		}

		// Copy group key columns
		for _, gkCol := range groupKey.Cols() {
			fromColumn := execute.ColIdx(gkCol.Label, chunk.Cols())

			oldValues := chunk.Values(fromColumn)
			var values array.Array
			if newChunkLen == chunk.Len() {
				values = oldValues
				values.Retain()
			} else {
				// We have nulls for some of the fields which we must exclude from the unpivoted
				// output, otherwise they show up as extra rows
				values = array.Slice(oldValues, 0, newChunkLen)
			}
			buffer.Values = append(buffer.Values, values)
		}
		// append the name of the value column into _field
		if defaultFieldColLen != 0 {
			buffer.Values = append(buffer.Values, array.StringRepeat(c.Label, newChunkLen, mem))
		}

		// Copy cols that are neither group columns nor value columns
		// Often _time will be in here, but sometimes others as well, in case of a group() transformation.
		for _, otherCol := range otherCols {
			if otherCol.Label == influxdb.DefaultFieldColLabel {
				buffer.Values = append(buffer.Values, array.StringRepeat(c.Label, newChunkLen, mem))
				continue
			}
			fromColumn := execute.ColIdx(otherCol.Label, chunk.Cols())

			// copy these cols but only when the value column does not have a null
			oldValues := chunk.Values(fromColumn)
			newValues := array.CopyValidValues(mem, oldValues, chunk.Values(i))
			buffer.Values = append(buffer.Values, newValues)
		}

		oldValues := chunk.Values(i)
		values := array.CopyValidValues(mem, oldValues, oldValues)
		buffer.Values = append(buffer.Values, values)

		out := table.ChunkFromBuffer(buffer)
		if err := d.Process(out); err != nil {
			return err
		}
	}

	return nil
}
