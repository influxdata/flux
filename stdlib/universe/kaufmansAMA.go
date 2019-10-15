package universe

import (
	"math"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
)

const kamaKind = "kaufmansAMA"

type KamaOpSpec struct {
	N      int64  `json:"n"`
	Column string `json:"column"`
}

func init() {
	kamaSignature := flux.FunctionSignature(
		map[string]semantic.PolyType{
			"n":      semantic.Int,
			"column": semantic.String,
		},
		[]string{"n"},
	)

	flux.RegisterPackageValue("universe", kamaKind, flux.FunctionValue(kamaKind, createkamaOpSpec, kamaSignature))
	flux.RegisterOpSpec(kamaKind, newkamaOp)
	plan.RegisterProcedureSpec(kamaKind, newkamaProcedure, kamaKind)
	execute.RegisterTransformation(kamaKind, createkamaTransformation)
}

func createkamaOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(KamaOpSpec)

	if n, err := args.GetRequiredInt("n"); err != nil {
		return nil, err
	} else {
		spec.N = n
	}

	if col, ok, err := args.GetString("column"); err != nil {
		return nil, err
	} else if !ok {
		spec.Column = execute.DefaultValueColLabel
	} else {
		spec.Column = col
	}

	return spec, nil
}

func newkamaOp() flux.OperationSpec {
	return new(KamaOpSpec)
}

func (s *KamaOpSpec) Kind() flux.OperationKind {
	return kamaKind
}

type KamaProcedureSpec struct {
	plan.DefaultCost
	N      int64  `json:"n"`
	Column string `json:"column"`
}

func newkamaProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*KamaOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}

	return &KamaProcedureSpec{
		N:      spec.N,
		Column: spec.Column,
	}, nil
}

func (s *KamaProcedureSpec) Kind() plan.ProcedureKind {
	return kamaKind
}

func (s *KamaProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(KamaProcedureSpec)
	*ns = *s
	return ns
}

// TriggerSpec implements plan.TriggerAwareProcedureSpec
func (s *KamaProcedureSpec) TriggerSpec() plan.TriggerSpec {
	return plan.NarrowTransformationTriggerSpec{}
}

func createkamaTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*KamaProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}
	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)
	t := NewkamaTransformation(d, cache, s)
	return t, d, nil
}

type kamaTransformation struct {
	d     execute.Dataset
	cache execute.TableBuilderCache

	n      int64
	column string
}

func NewkamaTransformation(d execute.Dataset, cache execute.TableBuilderCache, spec *KamaProcedureSpec) *kamaTransformation {
	return &kamaTransformation{
		d:     d,
		cache: cache,

		n:      spec.N,
		column: spec.Column,
	}
}

func (t *kamaTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *kamaTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	builder, created := t.cache.TableBuilder(tbl.Key())
	if !created {
		return errors.Newf(codes.FailedPrecondition, "KAMA found duplicate table with key: %v", tbl.Key())
	}
	if t.n <= 0 {
		return errors.Newf(codes.Invalid, "cannot take KaufmansAMA with a period of %v (must be greater than 0)", t.n)
	}
	cols := tbl.Cols()
	doKAMA := make([]bool, len(cols))
	for j, c := range cols {
		found := false
		if c.Label == t.column {
			if c.Type != flux.TInt && c.Type != flux.TUInt && c.Type != flux.TFloat {
				return errors.Newf(codes.FailedPrecondition, "cannot take KAMA of column %s (type %s)", c.Label, c.Type.String())
			}
			found = true
		}

		if found {
			mac := c
			mac.Type = flux.TFloat
			_, err := builder.AddCol(mac)
			if err != nil {
				return err
			}
			doKAMA[j] = true
		} else {
			_, err := builder.AddCol(c)
			if err != nil {
				return err
			}
		}
	}

	var prevValue float64
	var currValue float64
	var prevKAMA float64
	var sumUp float64
	var sumDown float64
	var diffNAgo []float64 // keeps track of the last n diff values

	var rowCount []int64

	prevValue = 0
	currValue = 0
	prevKAMA = 0
	sumUp = 0
	sumDown = 0
	diffNAgo = make([]float64, t.n)
	rowCount = make([]int64, len(cols))

	err := tbl.Do(func(cr flux.ColReader) error {
		if cr.Len() == 0 || len(cr.Cols()) == 0 {
			return nil
		}

		for j, c := range cr.Cols() {
			if !doKAMA[j] {
				for i := 0; i < cr.Len(); i++ {
					if rowCount[j] >= t.n {
						if err := builder.AppendValue(j, execute.ValueForRow(cr, i, j)); err != nil {
							return err
						}
					}
					rowCount[j]++
				}
			} else {
				var values []float64
				switch c.Type {
				case flux.TInt:
					arr := cr.Ints(j)
					for i := 0; i < arr.Len(); i++ {
						values = append(values, float64(arr.Value(i)))
					}
				case flux.TUInt:
					arr := cr.UInts(j)
					for i := 0; i < arr.Len(); i++ {
						values = append(values, float64(arr.Value(i)))
					}
				case flux.TFloat:
					arr := cr.Floats(j)
					values = arr.Float64Values()
				}

				if rowCount[j] == 0 {
					prevValue = values[0]
					for i := 0; i < cr.Len(); i++ {
						currValue = values[i]
						kers, su, sd := t.nextKER(prevValue, currValue, sumUp, sumDown, diffNAgo, rowCount[j], t.n)
						sumUp = su
						sumDown = sd

						if rowCount[j] >= t.n {
							if rowCount[j] == t.n {
								prevKAMA = prevValue
							}
							var kama float64
							sc := 0.0
							kama = prevKAMA
							sc = math.Pow(kers*(2.0/(2.0+1.0)-2.0/(30.0+1.0))+2.0/(30.0+1.0), 2)
							kama = kama + sc*(currValue-kama)
							if err := builder.AppendFloat(j, kama); err != nil {
								return err
							}
							prevKAMA = kama
						}

						diffNAgo[rowCount[j]%t.n] = currValue - prevValue

						rowCount[j]++
						prevValue = currValue
					}
				} else {
					for i := 0; i < cr.Len(); i++ {
						currValue = values[i]
						kers, su, sd := t.nextKER(prevValue, currValue, sumUp, sumDown, diffNAgo, rowCount[j], t.n)
						sumUp = su
						sumDown = sd

						if rowCount[j] >= t.n {
							if rowCount[j] == t.n {
								prevKAMA = prevValue
							}
							var kama float64
							sc := 0.0
							kama = prevKAMA
							sc = math.Pow(kers*(2.0/(2.0+1.0)-2.0/(30.0+1.0))+2.0/(30.0+1.0), 2)
							kama = kama + sc*(currValue-kama)
							if err := builder.AppendFloat(j, kama); err != nil {
								return err
							}
							prevKAMA = kama
						}

						diffNAgo[rowCount[j]%t.n] = currValue - prevValue

						rowCount[j]++
						prevValue = currValue
					}
				}
			}
		}

		return nil
	})

	return err
}

// gives the current KER value, after considering the current value
func (t *kamaTransformation) nextKER(prevValue, currValue, sumUp, sumDown float64, diffNAgo []float64, count, n int64) (float64, float64, float64) {
	diff := currValue - prevValue
	if count >= n {
		val, su, sd := nextCMO(sumUp, sumDown, diff, diffNAgo[(count+1)%n])
		return math.Abs(val) / 100.0, su, sd
	}
	_, su, sd := nextCMO(sumUp, sumDown, diff, 0)
	return -999, su, sd
}

func (t *kamaTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}

func (t *kamaTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}

func (t *kamaTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}
