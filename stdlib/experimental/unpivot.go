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
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/values"
)

const UnpivotKind = "experimental.unpivot"

type UnpivotOpSpec struct{}

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

	return new(UnpivotOpSpec), nil
}

func (s *UnpivotOpSpec) Kind() flux.OperationKind {
	return UnpivotKind
}

func newUnpivotProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	_, ok := qs.(*UnpivotOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}

	return &UnpivotProcedureSpec{}, nil
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
	t := &unpivotTransformation{}
	return execute.NewNarrowTransformation(id, t, alloc)

}

type unpivotTransformation struct {
	execute.ExecutionNode
}

func (t *unpivotTransformation) Close() error { return nil }

func (t *unpivotTransformation) Process(chunk table.Chunk, d *execute.TransportDataset, mem memory.Allocator) error {

	timeColumn := -1
	for j, c := range chunk.Cols() {
		if c.Label == "_time" {
			timeColumn = j
			break
		}
	}

	if timeColumn == -1 || chunk.Cols()[timeColumn].Type != flux.TTime {
		return errors.Newf(codes.Internal, "Expected a `_time` column in the input")
	}

	for i, c := range chunk.Cols() {
		if chunk.Key().HasCol(c.Label) || c.Label == "_time" {
			continue
		}

		groupKey := chunk.Key()
		columns := groupKey.Cols()
		columns = append(columns,
			flux.ColMeta{Label: "_field", Type: flux.TString},
			flux.ColMeta{Label: "_time", Type: flux.TTime},
			flux.ColMeta{Label: "_value", Type: c.Type},
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
		for toColumn, groupColumn := range groupKey.Cols() {
			fromColumn := -1
			for j, c := range chunk.Cols() {
				if c.Label == groupColumn.Label {
					fromColumn = j
					break
				}
			}
			values := chunk.Values(fromColumn)
			values.Retain()
			buffer.Values[toColumn] = values
		}

		buffer.Values[len(buffer.Values)-3] = array.StringRepeat(c.Label, chunk.Len(), mem)

		times := chunk.Values(timeColumn)
		times.Retain()
		buffer.Values[len(buffer.Values)-2] = times

		values := chunk.Values(i)
		values.Retain()
		buffer.Values[len(buffer.Values)-1] = values

		out := table.ChunkFromBuffer(buffer)
		if err := d.Process(out); err != nil {
			return err
		}
	}

	return nil
}
