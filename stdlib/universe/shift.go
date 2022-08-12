package universe

import (
	"time"

	"github.com/apache/arrow/go/v7/arrow/memory"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/table"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

const ShiftKind = "timeShift"

type ShiftOpSpec struct {
	Shift   flux.Duration `json:"duration"`
	Columns []string      `json:"columns"`
}

func init() {
	shiftSignature := runtime.MustLookupBuiltinType("universe", "timeShift")

	runtime.RegisterPackageValue("universe", ShiftKind, flux.MustValue(flux.FunctionValue(ShiftKind, createShiftOpSpec, shiftSignature)))
	plan.RegisterProcedureSpec(ShiftKind, newShiftProcedure, ShiftKind)
	execute.RegisterTransformation(ShiftKind, createShiftTransformation)
}

func createShiftOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(ShiftOpSpec)

	if shift, err := args.GetRequiredDuration("duration"); err != nil {
		return nil, err
	} else {
		spec.Shift = shift
	}

	if cols, ok, err := args.GetArray("columns", semantic.String); err != nil {
		return nil, err
	} else if ok {
		columns, err := interpreter.ToStringArray(cols)
		if err != nil {
			return nil, err
		}
		spec.Columns = columns
	} else {
		spec.Columns = []string{
			execute.DefaultTimeColLabel,
			execute.DefaultStopColLabel,
			execute.DefaultStartColLabel,
		}
	}
	return spec, nil
}

func (s *ShiftOpSpec) Kind() flux.OperationKind {
	return ShiftKind
}

type ShiftProcedureSpec struct {
	plan.DefaultCost
	Shift   flux.Duration
	Columns []string
	Now     time.Time
}

// TimeBounds implements plan.BoundsAwareProcedureSpec
func (s *ShiftProcedureSpec) TimeBounds(predecessorBounds *plan.Bounds) *plan.Bounds {
	if predecessorBounds != nil {
		return predecessorBounds.Shift(values.Duration(s.Shift))
	}
	return nil
}

func newShiftProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*ShiftOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}

	return &ShiftProcedureSpec{
		Shift:   spec.Shift,
		Columns: spec.Columns,
		Now:     pa.Now(),
	}, nil
}

func (s *ShiftProcedureSpec) Kind() plan.ProcedureKind {
	return ShiftKind
}

func (s *ShiftProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(ShiftProcedureSpec)
	*ns = *s

	if s.Columns != nil {
		ns.Columns = make([]string, len(s.Columns))
		copy(ns.Columns, s.Columns)
	}
	return ns
}

// TriggerSpec implements plan.TriggerAwareProcedureSpec
func (s *ShiftProcedureSpec) TriggerSpec() plan.TriggerSpec {
	return plan.NarrowTransformationTriggerSpec{}
}

func createShiftTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*ShiftProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}
	return NewShiftTransformation(id, s, a.Allocator())
}

type shiftTransformation struct {
	columns []string
	shift   execute.Duration
}

func NewShiftTransformation(id execute.DatasetID, spec *ShiftProcedureSpec, mem memory.Allocator) (execute.Transformation, execute.Dataset, error) {
	tr := &shiftTransformation{
		columns: spec.Columns,
		shift:   spec.Shift,
	}
	return execute.NewNarrowTransformation(id, tr, mem)
}

func (s *shiftTransformation) Process(chunk table.Chunk, d *execute.TransportDataset, mem memory.Allocator) error {
	key := chunk.Key()
	for _, c := range key.Cols() {
		if execute.ContainsStr(s.columns, c.Label) {
			k, err := s.regenerateKey(key)
			if err != nil {
				return err
			}
			key = k
			break
		}
	}

	buffer := arrow.TableBuffer{
		GroupKey: key,
		Columns:  chunk.Cols(),
		Values:   make([]array.Array, chunk.NCols()),
	}
	for j, c := range chunk.Cols() {
		vs := chunk.Values(j)
		if execute.ContainsStr(s.columns, c.Label) {
			if c.Type != flux.TTime {
				return errors.Newf(codes.FailedPrecondition, "column %q is not of type time", c.Label)
			}
			buffer.Values[j] = s.shiftTimes(vs.(*array.Int), mem)
		} else {
			vs.Retain()
			buffer.Values[j] = vs
		}
	}

	out := table.ChunkFromBuffer(buffer)
	return d.Process(out)
}

func (s *shiftTransformation) regenerateKey(key flux.GroupKey) (flux.GroupKey, error) {
	cols := key.Cols()
	vals := make([]values.Value, len(cols))
	for j, c := range cols {
		if execute.ContainsStr(s.columns, c.Label) {
			if c.Type != flux.TTime {
				return nil, errors.Newf(codes.FailedPrecondition, "column %q is not of type time", c.Label)
			}
			vals[j] = values.NewTime(key.ValueTime(j).Add(s.shift))
		} else {
			vals[j] = key.Value(j)
		}
	}
	return execute.NewGroupKey(cols, vals), nil
}

func (s *shiftTransformation) shiftTimes(vs *array.Int, mem memory.Allocator) *array.Int {
	b := array.NewIntBuilder(mem)
	b.Resize(vs.Len())
	for i, n := 0, vs.Len(); i < n; i++ {
		if vs.IsNull(i) {
			b.AppendNull()
			continue
		}

		ts := execute.Time(vs.Value(i)).Add(s.shift)
		b.Append(int64(ts))
	}
	return b.NewIntArray()
}

func (s *shiftTransformation) Close() error { return nil }
