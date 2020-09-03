package universe

import (
	"context"

	"github.com/apache/arrow/go/arrow/array"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/internal/execute/table"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
)

const LimitKind = "limit"

// LimitOpSpec limits the number of rows returned per table.
type LimitOpSpec struct {
	N      int64 `json:"n"`
	Offset int64 `json:"offset"`
}

func init() {
	limitSignature := runtime.MustLookupBuiltinType("universe", "limit")

	runtime.RegisterPackageValue("universe", LimitKind, flux.MustValue(flux.FunctionValue(LimitKind, createLimitOpSpec, limitSignature)))
	flux.RegisterOpSpec(LimitKind, newLimitOp)
	plan.RegisterProcedureSpec(LimitKind, newLimitProcedure, LimitKind)
	// TODO register a range transformation. Currently range is only supported if it is pushed down into a select procedure.
	execute.RegisterTransformation(LimitKind, createLimitTransformation)
}

func createLimitOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(LimitOpSpec)

	n, err := args.GetRequiredInt("n")
	if err != nil {
		return nil, err
	}
	spec.N = n

	if offset, ok, err := args.GetInt("offset"); err != nil {
		return nil, err
	} else if ok {
		spec.Offset = offset
	}

	return spec, nil
}

func newLimitOp() flux.OperationSpec {
	return new(LimitOpSpec)
}

func (s *LimitOpSpec) Kind() flux.OperationKind {
	return LimitKind
}

type LimitProcedureSpec struct {
	plan.DefaultCost
	N      int64 `json:"n"`
	Offset int64 `json:"offset"`
}

func newLimitProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*LimitOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}
	return &LimitProcedureSpec{
		N:      spec.N,
		Offset: spec.Offset,
	}, nil
}

func (s *LimitProcedureSpec) Kind() plan.ProcedureKind {
	return LimitKind
}
func (s *LimitProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(LimitProcedureSpec)
	*ns = *s
	return ns
}

// TriggerSpec implements plan.TriggerAwareProcedureSpec
func (s *LimitProcedureSpec) TriggerSpec() plan.TriggerSpec {
	return plan.NarrowTransformationTriggerSpec{}
}

func createLimitTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*LimitProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}
	t, d := NewLimitTransformation(s, id)
	return t, d, nil
}

type limitTransformation struct {
	execute.ExecutionNode
	d         *execute.PassthroughDataset
	n, offset int
}

func NewLimitTransformation(spec *LimitProcedureSpec, id execute.DatasetID) (execute.Transformation, execute.Dataset) {
	d := execute.NewPassthroughDataset(id)
	t := &limitTransformation{
		d:      d,
		n:      int(spec.N),
		offset: int(spec.Offset),
	}
	return t, d
}

func (t *limitTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *limitTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	tbl, err := table.Stream(tbl.Key(), tbl.Cols(), func(ctx context.Context, w *table.StreamWriter) error {
		return t.limitTable(ctx, w, tbl)
	})
	if err != nil {
		return err
	}
	return t.d.Process(tbl)
}

func (t *limitTransformation) limitTable(ctx context.Context, w *table.StreamWriter, tbl flux.Table) error {
	n, offset := t.n, t.offset
	return tbl.Do(func(cr flux.ColReader) error {
		if n <= 0 {
			return nil
		}
		l := cr.Len()
		if l <= offset {
			offset -= l
			// Skip entire batch
			return nil
		}
		start := offset
		stop := l
		count := stop - start
		if count > n {
			count = n
			stop = start + count
		}
		n -= count

		vs := make([]array.Interface, len(cr.Cols()))
		for j := range vs {
			arr := table.Values(cr, j)
			if arr.Len() == count {
				arr.Retain()
			} else {
				arr = arrow.Slice(arr, int64(start), int64(stop))
			}
			vs[j] = arr
		}
		return w.Write(vs)
	})
}

func appendSlicedCols(reader flux.ColReader, builder execute.TableBuilder, start, stop int) error {
	for j, c := range reader.Cols() {
		if j > len(builder.Cols()) {
			return errors.New(codes.Internal, "builder index out of bounds")
		}

		switch c.Type {
		case flux.TBool:
			s := arrow.BoolSlice(reader.Bools(j), start, stop)
			if err := builder.AppendBools(j, s); err != nil {
				s.Release()
				return err
			}
			s.Release()
		case flux.TInt:
			s := arrow.IntSlice(reader.Ints(j), start, stop)
			if err := builder.AppendInts(j, s); err != nil {
				s.Release()
				return err
			}
			s.Release()
		case flux.TUInt:
			s := arrow.UintSlice(reader.UInts(j), start, stop)
			if err := builder.AppendUInts(j, s); err != nil {
				s.Release()
				return err
			}
			s.Release()
		case flux.TFloat:
			s := arrow.FloatSlice(reader.Floats(j), start, stop)
			if err := builder.AppendFloats(j, s); err != nil {
				s.Release()
				return err
			}
			s.Release()
		case flux.TString:
			s := arrow.StringSlice(reader.Strings(j), start, stop)
			if err := builder.AppendStrings(j, s); err != nil {
				s.Release()
				return err
			}
			s.Release()
		case flux.TTime:
			s := arrow.IntSlice(reader.Times(j), start, stop)
			if err := builder.AppendTimes(j, s); err != nil {
				s.Release()
				return err
			}
			s.Release()
		default:
			execute.PanicUnknownType(c.Type)
		}
	}

	return nil
}

func (t *limitTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}
func (t *limitTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}
func (t *limitTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}
