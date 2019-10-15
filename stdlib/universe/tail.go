package universe

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
)

const TailKind = "tail"

// TailOpSpec tails the number of rows returned per table.
type TailOpSpec struct {
	N      int64 `json:"n"`
	Offset int64 `json:"offset"`
}

func init() {
	tailSignature := flux.FunctionSignature(
		map[string]semantic.PolyType{
			"n":      semantic.Int,
			"offset": semantic.Int,
		},
		[]string{"n"},
	)

	flux.RegisterPackageValue("universe", TailKind, flux.FunctionValue(TailKind, createTailOpSpec, tailSignature))
	flux.RegisterOpSpec(TailKind, newTailOp)
	plan.RegisterProcedureSpec(TailKind, newTailProcedure, TailKind)
	execute.RegisterTransformation(TailKind, createTailTransformation)
}

func createTailOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(TailOpSpec)

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

func newTailOp() flux.OperationSpec {
	return new(TailOpSpec)
}

func (s *TailOpSpec) Kind() flux.OperationKind {
	return TailKind
}

type TailProcedureSpec struct {
	plan.DefaultCost
	N      int64 `json:"n"`
	Offset int64 `json:"offset"`
}

func newTailProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*TailOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}
	return &TailProcedureSpec{
		N:      spec.N,
		Offset: spec.Offset,
	}, nil
}

func (s *TailProcedureSpec) Kind() plan.ProcedureKind {
	return TailKind
}
func (s *TailProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(TailProcedureSpec)
	*ns = *s
	return ns
}

// TriggerSpec implements plan.TriggerAwareProcedureSpec
func (s *TailProcedureSpec) TriggerSpec() plan.TriggerSpec {
	return plan.NarrowTransformationTriggerSpec{}
}

func createTailTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*TailProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}
	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)
	t := NewTailTransformation(d, cache, s)
	return t, d, nil
}

type tailTransformation struct {
	d     execute.Dataset
	cache execute.TableBuilderCache

	n, offset int
}

func NewTailTransformation(d execute.Dataset, cache execute.TableBuilderCache, spec *TailProcedureSpec) *tailTransformation {
	return &tailTransformation{
		d:      d,
		cache:  cache,
		n:      int(spec.N),
		offset: int(spec.Offset),
	}
}

func (t *tailTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *tailTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	builder, created := t.cache.TableBuilder(tbl.Key())
	if !created {
		return errors.Newf(codes.FailedPrecondition, "tail found duplicate table with key: %v", tbl.Key())
	}
	if err := execute.AddTableCols(tbl, builder); err != nil {
		return err
	}

	n := t.n
	offset := t.offset
	readers := make([]flux.ColReader, 0)
	numRecords := 0

	var finished bool
	if err := tbl.Do(func(cr flux.ColReader) error {
		if n <= 0 {
			// Returning an error terminates iteration
			finished = true
			return errors.New(codes.Canceled)
		}

		cr.Retain()
		readers = append(readers, cr)
		numRecords += cr.Len()

		for numRecords-readers[0].Len() >= n+offset {
			numRecords -= readers[0].Len()
			readers[0].Release()
			readers = readers[1:]
		}

		return nil
	}); err != nil && !finished {
		return err
	}

	endIndex := numRecords
	offsetIndex := endIndex - offset
	startIndex := offsetIndex - n

	curr := 0
	for _, cr := range readers {
		var start, end int

		if startIndex > curr && startIndex < cr.Len() {
			start = startIndex
		} else {
			start = 0
		}

		if offsetIndex > curr && offsetIndex < curr+cr.Len() {
			end = offsetIndex - curr
		} else if offsetIndex <= curr {
			break
		} else {
			end = cr.Len()
		}

		if err := appendSlicedCols(cr, builder, start, end); err != nil {
			return err
		}

		curr += cr.Len()

		cr.Release()
	}

	return nil
}

func (t *tailTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}
func (t *tailTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}
func (t *tailTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}
