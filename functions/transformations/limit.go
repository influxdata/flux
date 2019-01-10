package transformations

import (
	"fmt"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
	"github.com/pkg/errors"
)

const LimitKind = "limit"

// LimitOpSpec limits the number of rows returned per table.
type LimitOpSpec struct {
	N      int64 `json:"n"`
	Offset int64 `json:"offset"`
}

func init() {
	limitSignature := flux.FunctionSignature(
		map[string]semantic.PolyType{
			"n":      semantic.Int,
			"offset": semantic.Int,
		},
		[]string{"n"},
	)

	flux.RegisterFunction(LimitKind, createLimitOpSpec, limitSignature)
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
		return nil, fmt.Errorf("invalid spec type %T", qs)
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

func createLimitTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*LimitProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", spec)
	}
	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)
	t := NewLimitTransformation(d, cache, s)
	return t, d, nil
}

type limitTransformation struct {
	d     execute.Dataset
	cache execute.TableBuilderCache

	n, offset int
}

func NewLimitTransformation(d execute.Dataset, cache execute.TableBuilderCache, spec *LimitProcedureSpec) *limitTransformation {
	return &limitTransformation{
		d:      d,
		cache:  cache,
		n:      int(spec.N),
		offset: int(spec.Offset),
	}
}

func (t *limitTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *limitTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	builder, created := t.cache.TableBuilder(tbl.Key())
	if !created {
		return fmt.Errorf("limit found duplicate table with key: %v", tbl.Key())
	}
	if err := execute.AddTableCols(tbl, builder); err != nil {
		return err
	}
	// AppendTable with limit
	n := t.n
	offset := t.offset
	var finishedErr error
	err := tbl.Do(func(cr flux.ColReader) error {
		if n <= 0 {
			// Returning an error terminates iteration
			finishedErr = errors.New("finished")
			return finishedErr
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

		err := appendSlicedCols(cr, builder, start, stop)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil && finishedErr == nil {
		return err
	}
	return nil
}

func appendSlicedCols(reader flux.ColReader, builder execute.TableBuilder, start, stop int) error {
	for j, c := range reader.Cols() {
		if j > len(builder.Cols()) {
			return errors.New("builder index out of bounds")
		}

		switch c.Type {
		case flux.TBool:
			s := arrow.BoolSlice(reader.Bools(j), start, stop)
			if err := arrow.AppendBools(builder, j, s); err != nil {
				s.Release()
				return err
			}
			s.Release()
		case flux.TInt:
			s := arrow.IntSlice(reader.Ints(j), start, stop)
			if err := arrow.AppendInts(builder, j, s); err != nil {
				s.Release()
				return err
			}
			s.Release()
		case flux.TUInt:
			s := arrow.UintSlice(reader.UInts(j), start, stop)
			if err := arrow.AppendUInts(builder, j, s); err != nil {
				s.Release()
				return err
			}
			s.Release()
		case flux.TFloat:
			s := arrow.FloatSlice(reader.Floats(j), start, stop)
			if err := arrow.AppendFloats(builder, j, s); err != nil {
				s.Release()
				return err
			}
			s.Release()
		case flux.TString:
			s := arrow.StringSlice(reader.Strings(j), start, stop)
			if err := arrow.AppendStrings(builder, j, s); err != nil {
				s.Release()
				return err
			}
			s.Release()
		case flux.TTime:
			s := arrow.IntSlice(reader.Times(j), start, stop)
			if err := arrow.AppendTimes(builder, j, s); err != nil {
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
