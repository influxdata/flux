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
	"github.com/influxdata/flux/values"
)

const UnpivotKind = "experimental.unpivot"

type UnpivotOpSpec struct {
	ungroupedTagColumns []string
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

	if columns, ok, err := args.GetArray("ungroupedTagColumns", semantic.String); err != nil {
		return nil, err
	} else if ok {
		spec.ungroupedTagColumns, err = interpreter.ToStringArray(columns)
		if err != nil {
			return nil, err
		}
	} else {
		spec.ungroupedTagColumns = []string{}
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
		ungroupedTagColumns: opSpec.ungroupedTagColumns,
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
	ungroupedTagColumns []string
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
		ungroupedTagColumns: spec.ungroupedTagColumns,
	}
	return execute.NewNarrowTransformation(id, t, alloc)

}

type unpivotTransformation struct {
	execute.ExecutionNode
	ungroupedTagColumns []string
}

func (t *unpivotTransformation) Close() error { return nil }

func (t *unpivotTransformation) Process(chunk table.Chunk, d *execute.TransportDataset, mem memory.Allocator) error {
	timeColumn := -1
	for j, c := range chunk.Cols() {
		if c.Label == execute.DefaultTimeColLabel {
			timeColumn = j
			break
		}
	}

	ungroupedTagCols := make([]flux.ColMeta, len(t.ungroupedTagColumns))
	for i, utc := range t.ungroupedTagColumns {
		found := false
		for _, c := range chunk.Cols() {
			if c.Label == utc {
				ungroupedTagCols[i] = c
				found = true
				break
			}
		}
		if !found {
			return errors.Newf(codes.Internal, "unpivot could not find column named %q", utc)
		}
	}

	for i, c := range chunk.Cols() {
		if chunk.Key().HasCol(c.Label) || c.Label == execute.DefaultTimeColLabel || execute.HasCol(c.Label, ungroupedTagCols) {
			continue
		}

		chunkValues := chunk.Values(i)
		chunkValues.Retain()

		newChunkLen := chunk.Len() - chunkValues.NullN()

		groupKey := chunk.Key()
		columns := groupKey.Cols()
		if timeColumn != -1 {
			columns = append(columns, flux.ColMeta{Label: execute.DefaultTimeColLabel, Type: flux.TTime})
		}

		for _, ungroupedTagCol := range ungroupedTagCols {
			columns = append(columns, ungroupedTagCol)
		}
		columns = append(columns,
			flux.ColMeta{Label: "_field", Type: flux.TString},
			flux.ColMeta{Label: execute.DefaultValueColLabel, Type: c.Type},
		)

		groupCols := []flux.ColMeta{{Label: "_field", Type: flux.TString}}
		groupValues := []values.Value{values.NewString(c.Label)}
		buffer := arrow.TableBuffer{
			GroupKey: groupkey.New(
				append(groupCols, groupKey.Cols()...),
				append(groupValues, groupKey.Values()...),
			),
			Columns: columns,
			Values:  make([]array.Array, len(columns)),
		}

		// Copy group key columns
		for toColumn, _ := range groupKey.Cols() {
			fromColumn := execute.ColIdx(c.Label, chunk.Cols())

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
			buffer.Values[toColumn] = values
		}

		oldValues := chunk.Values(i)

		if timeColumn != -1 {
			// The time array does not contain any nulls so we use the value array for that information
			times := array.CopyValidValues(mem, chunk.Values(timeColumn), oldValues)
			buffer.Values[len(buffer.Values)-3] = times
		}

		// Copy tag cols that are not in the group key because of a group transformation
		for idx, ungroupedTagCol := range ungroupedTagCols {
			fromColumn := execute.ColIdx(ungroupedTagCol.Label, chunk.Cols())

			// copy these cols but only when the value column does not have a null
			oldValues := chunk.Values(fromColumn)
			newValues := array.CopyValuesWithMask(mem, oldValues, chunk.Values(i))
			newIdx := len(groupKey.Cols()) + idx
			buffer.Values[newIdx] = newValues
		}

		buffer.Values[len(buffer.Values)-3] = array.StringRepeat(c.Label, newChunkLen, mem)

		times := array.CopyValuesWithMask(mem, chunk.Values(timeColumn), chunk.Values(i))
		buffer.Values[len(buffer.Values)-2] = times

		buffer.Values[len(buffer.Values)-2] = array.StringRepeat(c.Label, newChunkLen, mem)

		values := array.CopyValidValues(mem, oldValues, oldValues)
		buffer.Values[len(buffer.Values)-1] = values

		out := table.ChunkFromBuffer(buffer)
		if err := d.Process(out); err != nil {
			return err
		}
	}

	return nil
}
