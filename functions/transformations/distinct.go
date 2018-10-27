package transformations

import (
	"fmt"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/execute/tablebuilder"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
)

const DistinctKind = "distinct"

type DistinctOpSpec struct {
	Column string `json:"column"`
}

func init() {
	distinctSignature := flux.FunctionSignature(
		map[string]semantic.PolyType{
			"column": semantic.String,
		},
		nil,
	)

	flux.RegisterFunction(DistinctKind, createDistinctOpSpec, distinctSignature)
	flux.RegisterOpSpec(DistinctKind, newDistinctOp)
	plan.RegisterProcedureSpec(DistinctKind, newDistinctProcedure, DistinctKind)
	execute.RegisterTransformation(DistinctKind, createDistinctTransformation)
}

func createDistinctOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(DistinctOpSpec)

	if col, ok, err := args.GetString("column"); err != nil {
		return nil, err
	} else if ok {
		spec.Column = col
	} else {
		spec.Column = execute.DefaultValueColLabel
	}

	return spec, nil
}

func newDistinctOp() flux.OperationSpec {
	return new(DistinctOpSpec)
}

func (s *DistinctOpSpec) Kind() flux.OperationKind {
	return DistinctKind
}

type DistinctProcedureSpec struct {
	plan.DefaultCost
	Column string
}

func newDistinctProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*DistinctOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	return &DistinctProcedureSpec{
		Column: spec.Column,
	}, nil
}

func (s *DistinctProcedureSpec) Kind() plan.ProcedureKind {
	return DistinctKind
}
func (s *DistinctProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(DistinctProcedureSpec)

	*ns = *s

	return ns
}

func createDistinctTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*DistinctProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", spec)
	}
	cache := tablebuilder.NewCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)
	t := NewDistinctTransformation(d, cache, s)
	return t, d, nil
}

type distinctTransformation struct {
	d     execute.Dataset
	cache *tablebuilder.Cache

	column string
}

func NewDistinctTransformation(d execute.Dataset, cache *tablebuilder.Cache, spec *DistinctProcedureSpec) *distinctTransformation {
	return &distinctTransformation{
		d:      d,
		cache:  cache,
		column: spec.Column,
	}
}

func (t *distinctTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *distinctTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	builder := t.cache.Get(tbl.Key())

	colIdx := execute.ColIdx(t.column, tbl.Cols())
	if colIdx < 0 {
		if err := builder.AppendString(execute.DefaultValueColLabel, ""); err != nil {
			return err
		}
		// TODO: hack required to ensure data flows downstream
		return tbl.Do(func(flux.ColReader) error {
			return nil
		})
	}

	col := tbl.Cols()[colIdx]

	if tbl.Key().HasCol(t.column) {
		j := execute.ColIdx(t.column, tbl.Key().Cols())
		if err := builder.AppendValue(execute.DefaultValueColLabel, tbl.Key().Value(j)); err != nil {
			return err
		}

		// TODO: hack required to ensure data flows downstream
		return tbl.Do(func(flux.ColReader) error {
			return nil
		})
	}

	switch col.Type {
	case flux.TBool:
		distinct := make(map[bool]bool)
		return builder.Bools(execute.DefaultValueColLabel).Do(func(b array.BooleanBuilder) error {
			return tbl.Do(func(cr flux.ColReader) error {
				for _, v := range cr.Bools(colIdx) {
					if distinct[v] {
						continue
					}
					distinct[v] = true
					b.Append(v)
				}
				return nil
			})
		})
	case flux.TInt:
		distinct := make(map[int64]bool)
		return builder.Ints(execute.DefaultValueColLabel).Do(func(b array.IntBuilder) error {
			return tbl.Do(func(cr flux.ColReader) error {
				for _, v := range cr.Ints(colIdx) {
					if distinct[v] {
						continue
					}
					distinct[v] = true
					b.Append(v)
				}
				return nil
			})
		})
	case flux.TUInt:
		distinct := make(map[uint64]bool)
		return builder.UInts(execute.DefaultValueColLabel).Do(func(b array.UIntBuilder) error {
			return tbl.Do(func(cr flux.ColReader) error {
				for _, v := range cr.UInts(colIdx) {
					if distinct[v] {
						continue
					}
					distinct[v] = true
					b.Append(v)
				}
				return nil
			})
		})
	case flux.TFloat:
		distinct := make(map[float64]bool)
		return builder.Floats(execute.DefaultValueColLabel).Do(func(b array.FloatBuilder) error {
			return tbl.Do(func(cr flux.ColReader) error {
				for _, v := range cr.Floats(colIdx) {
					if distinct[v] {
						continue
					}
					distinct[v] = true
					b.Append(v)
				}
				return nil
			})
		})
	case flux.TString:
		distinct := make(map[string]bool)
		return builder.Strings(execute.DefaultValueColLabel).Do(func(b array.StringBuilder) error {
			return tbl.Do(func(cr flux.ColReader) error {
				for _, v := range cr.Strings(colIdx) {
					if distinct[v] {
						continue
					}
					distinct[v] = true
					b.Append(v)
				}
				return nil
			})
		})
	case flux.TTime:
		distinct := make(map[execute.Time]bool)
		return builder.Times(execute.DefaultValueColLabel).Do(func(b array.TimeBuilder) error {
			return tbl.Do(func(cr flux.ColReader) error {
				for _, v := range cr.Times(colIdx) {
					if distinct[v] {
						continue
					}
					distinct[v] = true
					b.Append(v)
				}
				return nil
			})
		})
	default:
		return nil
	}
}

func (t *distinctTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}
func (t *distinctTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}
func (t *distinctTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}
