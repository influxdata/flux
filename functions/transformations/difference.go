package transformations

import (
	"fmt"
	"math"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
)

const DifferenceKind = "difference"

type DifferenceOpSpec struct {
	NonNegative bool     `json:"nonNegative"`
	Columns     []string `json:"columns"`
}

func init() {
	differenceSignature := flux.FunctionSignature(
		map[string]semantic.PolyType{
			"nonNegative": semantic.Bool,
			"columns":     semantic.NewArrayPolyType(semantic.String),
		},
		nil,
	)

	flux.RegisterFunction(DifferenceKind, createDifferenceOpSpec, differenceSignature)
	flux.RegisterOpSpec(DifferenceKind, newDifferenceOp)
	plan.RegisterProcedureSpec(DifferenceKind, newDifferenceProcedure, DifferenceKind)
	execute.RegisterTransformation(DifferenceKind, createDifferenceTransformation)
}

func createDifferenceOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	err := a.AddParentFromArgs(args)
	if err != nil {
		return nil, err
	}

	spec := new(DifferenceOpSpec)

	if nn, ok, err := args.GetBool("nonNegative"); err != nil {
		return nil, err
	} else if ok {
		spec.NonNegative = nn
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
		spec.Columns = []string{execute.DefaultValueColLabel}
	}

	return spec, nil
}

func newDifferenceOp() flux.OperationSpec {
	return new(DifferenceOpSpec)
}

func (s *DifferenceOpSpec) Kind() flux.OperationKind {
	return DifferenceKind
}

type DifferenceProcedureSpec struct {
	plan.DefaultCost
	NonNegative bool     `json:"non_negative"`
	Columns     []string `json:"columns"`
}

func newDifferenceProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*DifferenceOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	return &DifferenceProcedureSpec{
		NonNegative: spec.NonNegative,
		Columns:     spec.Columns,
	}, nil
}

func (s *DifferenceProcedureSpec) Kind() plan.ProcedureKind {
	return DifferenceKind
}
func (s *DifferenceProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(DifferenceProcedureSpec)
	*ns = *s
	if s.Columns != nil {
		ns.Columns = make([]string, len(s.Columns))
		copy(ns.Columns, s.Columns)
	}
	return ns
}

func createDifferenceTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*DifferenceProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", spec)
	}
	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)
	t := NewDifferenceTransformation(d, cache, s)
	return t, d, nil
}

type differenceTransformation struct {
	d     execute.Dataset
	cache execute.TableBuilderCache

	nonNegative bool
	columns     []string
}

func NewDifferenceTransformation(d execute.Dataset, cache execute.TableBuilderCache, spec *DifferenceProcedureSpec) *differenceTransformation {
	return &differenceTransformation{
		d:           d,
		cache:       cache,
		nonNegative: spec.NonNegative,
		columns:     spec.Columns,
	}
}

func (t *differenceTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *differenceTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	builder, created := t.cache.TableBuilder(tbl.Key())
	if !created {
		return fmt.Errorf("difference found duplicate table with key: %v", tbl.Key())
	}
	cols := tbl.Cols()
	differences := make([]*difference, len(cols))
	for j, c := range cols {
		found := false
		for _, label := range t.columns {
			if c.Label == label {
				found = true
				break
			}
		}

		if found {
			var typ flux.ColType
			switch c.Type {
			case flux.TInt, flux.TUInt:
				typ = flux.TInt
			case flux.TFloat:
				typ = flux.TFloat
			}
			if _, err := builder.AddCol(flux.ColMeta{
				Label: c.Label,
				Type:  typ,
			}); err != nil {
				return err
			}
			differences[j] = newDifference(j, t.nonNegative)
		} else {
			_, err := builder.AddCol(c)
			if err != nil {
				return err
			}
		}
	}

	// We need to drop the first row since its difference is undefined
	firstIdx := 1
	return tbl.Do(func(cr flux.ColReader) error {
		l := cr.Len()

		if l != 0 {
			for j, c := range cols {
				d := differences[j]
				switch c.Type {
				case flux.TBool:
					s := arrow.BoolSlice(cr.Bools(j), firstIdx, l)
					if err := builder.AppendBools(j, s); err != nil {
						s.Release()
						return err
					}
					s.Release()
				case flux.TInt:
					if d != nil {
						for i := 0; i < l; i++ {
							if vs := cr.Ints(j); vs.IsValid(i) {
								if v, first := d.updateInt(vs.Value(i)); !first {
									if d.nonNegative && v < 0 {
										if err := builder.AppendNil(j); err != nil {
											return err
										}
									} else {
										if err := builder.AppendInt(j, v); err != nil {
											return err
										}
									}
								}
							} else if err := builder.AppendNil(j); err != nil {
								return err
							}
						}
					} else {
						s := arrow.IntSlice(cr.Ints(j), firstIdx, l)
						if err := builder.AppendInts(j, s); err != nil {
							s.Release()
							return err
						}
						s.Release()
					}
				case flux.TUInt:
					if d != nil {
						for i := 0; i < l; i++ {
							if vs := cr.UInts(j); vs.IsValid(i) {
								if v, first := d.updateUInt(vs.Value(i)); !first {
									if d.nonNegative && v < 0 {
										if err := builder.AppendNil(j); err != nil {
											return err
										}
									} else {
										if err := builder.AppendInt(j, v); err != nil {
											return err
										}
									}
								}
							} else if err := builder.AppendNil(j); err != nil {
								return err
							}
						}
					} else {
						s := arrow.UintSlice(cr.UInts(j), firstIdx, l)
						if err := builder.AppendUInts(j, s); err != nil {
							s.Release()
							return err
						}
						s.Release()
					}
				case flux.TFloat:
					if d != nil {
						for i := 0; i < l; i++ {
							if vs := cr.Floats(j); vs.IsValid(i) {
								if v, first := d.updateFloat(vs.Value(i)); !first {
									if d.nonNegative && v < 0 {
										if err := builder.AppendNil(j); err != nil {
											return err
										}
									} else {
										if err := builder.AppendFloat(j, v); err != nil {
											return err
										}
									}
								}
							} else if err := builder.AppendNil(j); err != nil {
								return err
							}
						}
					} else {
						s := arrow.FloatSlice(cr.Floats(j), firstIdx, l)
						if err := builder.AppendFloats(j, s); err != nil {
							s.Release()
							return err
						}
						s.Release()
					}
				case flux.TString:
					s := arrow.StringSlice(cr.Strings(j), firstIdx, l)
					if err := builder.AppendStrings(j, s); err != nil {
						s.Release()
						return err
					}
					s.Release()
				case flux.TTime:
					s := arrow.IntSlice(cr.Times(j), firstIdx, l)
					if err := builder.AppendTimes(j, s); err != nil {
						s.Release()
						return err
					}
					s.Release()
				}
			}
		}

		// Now that we skipped the first row, start at 0 for the rest of the batches
		firstIdx = 0
		return nil
	})
}

func (t *differenceTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}
func (t *differenceTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}
func (t *differenceTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}

func newDifference(col int, nonNegative bool) *difference {
	return &difference{
		col:         col,
		first:       true,
		nonNegative: nonNegative,
	}
}

type difference struct {
	col         int
	first       bool
	nonNegative bool

	pIntValue   int64
	pUIntValue  uint64
	pFloatValue float64
}

func (d *difference) updateInt(v int64) (int64, bool) {
	if d.first {
		d.pIntValue = v
		d.first = false
		return 0, true
	}

	diff := v - d.pIntValue
	d.pIntValue = v

	return diff, false
}
func (d *difference) updateUInt(v uint64) (int64, bool) {
	if d.first {
		d.pUIntValue = v
		d.first = false
		return 0, true
	}

	var diff int64
	if d.pUIntValue > v {
		// Prevent uint64 overflow by applying the negative sign after the conversion to an int64.
		diff = int64(d.pUIntValue-v) * -1
	} else {
		diff = int64(v - d.pUIntValue)
	}

	d.pUIntValue = v

	return diff, false
}
func (d *difference) updateFloat(v float64) (float64, bool) {
	if d.first {
		d.pFloatValue = v
		d.first = false
		return math.NaN(), true
	}

	diff := v - d.pFloatValue
	d.pFloatValue = v

	return diff, false
}
