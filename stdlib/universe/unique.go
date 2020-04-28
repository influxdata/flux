package universe

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
)

const UniqueKind = "unique"

type UniqueOpSpec struct {
	Column string `json:"column"`
}

func init() {
	uniqueSignature := runtime.MustLookupBuiltinType("universe", "unique")

	runtime.RegisterPackageValue("universe", UniqueKind, flux.MustValue(flux.FunctionValue(UniqueKind, createUniqueOpSpec, uniqueSignature)))
	flux.RegisterOpSpec(UniqueKind, newUniqueOp)
	plan.RegisterProcedureSpec(UniqueKind, newUniqueProcedure, UniqueKind)
	execute.RegisterTransformation(UniqueKind, createUniqueTransformation)
}

func createUniqueOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(UniqueOpSpec)

	if col, ok, err := args.GetString("column"); err != nil {
		return nil, err
	} else if ok {
		spec.Column = col
	} else {
		spec.Column = execute.DefaultValueColLabel
	}

	return spec, nil
}

func newUniqueOp() flux.OperationSpec {
	return new(UniqueOpSpec)
}

func (s *UniqueOpSpec) Kind() flux.OperationKind {
	return UniqueKind
}

type UniqueProcedureSpec struct {
	plan.DefaultCost
	Column string
}

func newUniqueProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*UniqueOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}

	return &UniqueProcedureSpec{
		Column: spec.Column,
	}, nil
}

func (s *UniqueProcedureSpec) Kind() plan.ProcedureKind {
	return UniqueKind
}
func (s *UniqueProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(UniqueProcedureSpec)

	*ns = *s

	return ns
}

// TriggerSpec implements plan.TriggerAwareProcedureSpec
func (s *UniqueProcedureSpec) TriggerSpec() plan.TriggerSpec {
	return plan.NarrowTransformationTriggerSpec{}
}

func createUniqueTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*UniqueProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}
	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)
	t := NewUniqueTransformation(d, cache, s)
	return t, d, nil
}

type uniqueTransformation struct {
	d     execute.Dataset
	cache execute.TableBuilderCache

	column string
}

func NewUniqueTransformation(d execute.Dataset, cache execute.TableBuilderCache, spec *UniqueProcedureSpec) *uniqueTransformation {
	return &uniqueTransformation{
		d:      d,
		cache:  cache,
		column: spec.Column,
	}
}

func (t *uniqueTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *uniqueTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	builder, created := t.cache.TableBuilder(tbl.Key())
	if !created {
		return errors.Newf(codes.FailedPrecondition, "unique found duplicate table with key: %v", tbl.Key())
	}
	if err := execute.AddTableCols(tbl, builder); err != nil {
		return err
	}

	colIdx := execute.ColIdx(t.column, builder.Cols())
	if colIdx < 0 {
		return errors.Newf(codes.FailedPrecondition, "no column %q exists", t.column)
	}
	col := builder.Cols()[colIdx]

	var (
		boolUnique   map[bool]bool
		intUnique    map[int64]bool
		uintUnique   map[uint64]bool
		floatUnique  map[float64]bool
		stringUnique map[string]bool
		timeUnique   map[execute.Time]bool
		nullUnique   bool
	)
	switch col.Type {
	case flux.TBool:
		boolUnique = make(map[bool]bool)
	case flux.TInt:
		intUnique = make(map[int64]bool)
	case flux.TUInt:
		uintUnique = make(map[uint64]bool)
	case flux.TFloat:
		floatUnique = make(map[float64]bool)
	case flux.TString:
		stringUnique = make(map[string]bool)
	case flux.TTime:
		timeUnique = make(map[execute.Time]bool)
	}

	return tbl.Do(func(cr flux.ColReader) error {
		l := cr.Len()
		for i := 0; i < l; i++ {
			// Check unique
			switch col.Type {
			case flux.TBool:
				if vs := cr.Bools(colIdx); vs.IsNull(i) {
					if nullUnique {
						continue
					}
					nullUnique = true
				} else {
					v := vs.Value(i)
					if boolUnique[v] {
						continue
					}
					boolUnique[v] = true
				}
			case flux.TInt:
				if vs := cr.Ints(colIdx); vs.IsNull(i) {
					if nullUnique {
						continue
					}
					nullUnique = true
				} else {
					v := vs.Value(i)
					if intUnique[v] {
						continue
					}
					intUnique[v] = true
				}
			case flux.TUInt:
				if vs := cr.UInts(colIdx); vs.IsNull(i) {
					if nullUnique {
						continue
					}
					nullUnique = true
				} else {
					v := vs.Value(i)
					if uintUnique[v] {
						continue
					}
					uintUnique[v] = true
				}
			case flux.TFloat:
				if vs := cr.Floats(colIdx); vs.IsNull(i) {
					if nullUnique {
						continue
					}
					nullUnique = true
				} else {
					v := vs.Value(i)
					if floatUnique[v] {
						continue
					}
					floatUnique[v] = true
				}
			case flux.TString:
				if vs := cr.Strings(colIdx); vs.IsNull(i) {
					if nullUnique {
						continue
					}
					nullUnique = true
				} else {
					v := vs.ValueString(i)
					if stringUnique[v] {
						continue
					}
					stringUnique[v] = true
				}
			case flux.TTime:
				if vs := cr.Times(colIdx); vs.IsNull(i) {
					if nullUnique {
						continue
					}
					nullUnique = true
				} else {
					v := execute.Time(vs.Value(i))
					if timeUnique[v] {
						continue
					}
					timeUnique[v] = true
				}
			}

			if err := execute.AppendRecord(i, cr, builder); err != nil {
				return err
			}
		}
		return nil
	})
}

func (t *uniqueTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}
func (t *uniqueTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}
func (t *uniqueTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}
