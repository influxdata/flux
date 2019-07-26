package universe

import (
	"fmt"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
)

const ChandeMomentumOscillatorKind = "chandeMomentumOscillator"

type ChandeMomentumOscillatorOpSpec struct {
	N       int64    `json:"n"`
	Columns []string `json:"columns"`
}

func init() {
	chandeMomentumOscillatorSignature := flux.FunctionSignature(
		map[string]semantic.PolyType{
			"n":       semantic.Int,
			"columns": semantic.NewArrayPolyType(semantic.String),
		},
		[]string{"n"},
	)

	flux.RegisterPackageValue("universe", ChandeMomentumOscillatorKind, flux.FunctionValue(ChandeMomentumOscillatorKind, createChandeMomentumOscillatorOpSpec, chandeMomentumOscillatorSignature))
	flux.RegisterOpSpec(ChandeMomentumOscillatorKind, newChandeMomentumOscillatorOp)
	plan.RegisterProcedureSpec(ChandeMomentumOscillatorKind, newChandeMomentumOscillatorProcedure, ChandeMomentumOscillatorKind)
	execute.RegisterTransformation(ChandeMomentumOscillatorKind, createChandeMomentumOscillatorTransformation)
}

func createChandeMomentumOscillatorOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(ChandeMomentumOscillatorOpSpec)

	if n, err := args.GetRequiredInt("n"); err != nil {
		return nil, err
	} else {
		spec.N = n
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

func newChandeMomentumOscillatorOp() flux.OperationSpec {
	return new(ChandeMomentumOscillatorOpSpec)
}

func (s *ChandeMomentumOscillatorOpSpec) Kind() flux.OperationKind {
	return ChandeMomentumOscillatorKind
}

type ChandeMomentumOscillatorProcedureSpec struct {
	plan.DefaultCost
	N       int64    `json:"n"`
	Columns []string `json:"columns"`
}

func newChandeMomentumOscillatorProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*ChandeMomentumOscillatorOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	return &ChandeMomentumOscillatorProcedureSpec{
		N:       spec.N,
		Columns: spec.Columns,
	}, nil
}

func (s *ChandeMomentumOscillatorProcedureSpec) Kind() plan.ProcedureKind {
	return ChandeMomentumOscillatorKind
}

func (s *ChandeMomentumOscillatorProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(ChandeMomentumOscillatorProcedureSpec)
	*ns = *s
	if s.Columns != nil {
		ns.Columns = make([]string, len(s.Columns))
		copy(ns.Columns, s.Columns)
	}
	return ns
}

// TriggerSpec implements plan.TriggerAwareProcedureSpec
func (s *ChandeMomentumOscillatorProcedureSpec) TriggerSpec() plan.TriggerSpec {
	return plan.NarrowTransformationTriggerSpec{}
}

func createChandeMomentumOscillatorTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*ChandeMomentumOscillatorProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", spec)
	}
	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)
	t := NewChandeMomentumOscillatorTransformation(d, cache, s)
	return t, d, nil
}

type chandeMomentumOscillatorTransformation struct {
	d     execute.Dataset
	cache execute.TableBuilderCache

	n       int64
	columns []string
}

func NewChandeMomentumOscillatorTransformation(d execute.Dataset, cache execute.TableBuilderCache, spec *ChandeMomentumOscillatorProcedureSpec) *chandeMomentumOscillatorTransformation {
	return &chandeMomentumOscillatorTransformation{
		d:       d,
		cache:   cache,
		n:       spec.N,
		columns: spec.Columns,
	}
}

func (t *chandeMomentumOscillatorTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *chandeMomentumOscillatorTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	builder, created := t.cache.TableBuilder(tbl.Key())
	if !created {
		return fmt.Errorf("chande momentum oscillator found duplicate table with key: %v", tbl.Key())
	}
	cols := tbl.Cols()
	doChandeMomentumOscillator := make([]bool, len(cols))
	for j, c := range cols {
		found := false
		for _, label := range t.columns {
			if c.Label == label {
				if c.Type != flux.TInt && c.Type != flux.TUInt && c.Type != flux.TFloat {
					return fmt.Errorf("cannot take chande momentum oscillator of column %s (type %s)", c.Label, c.Type.String())
				}
				found = true
				break
			}
		}

		if found {
			mac := c
			mac.Type = flux.TFloat
			_, err := builder.AddCol(mac)
			if err != nil {
				return err
			}
			doChandeMomentumOscillator[j] = true
		} else {
			_, err := builder.AddCol(c)
			if err != nil {
				return err
			}
		}
	}

	return tbl.Do(func(cr flux.ColReader) error {
		if cr.Len() == 0 || len(cr.Cols()) == 0 {
			return nil
		}

		for j, c := range cr.Cols() {
			if !doChandeMomentumOscillator[j] {
				for i := int(t.n); i < cr.Len(); i++ {
					if err := builder.AppendValue(j, execute.ValueForRow(cr, i, j)); err != nil {
						return err
					}
				}
				continue
			}
			if err := t.do(int(t.n), cr, c, builder, j); err != nil {
				return err
			}
		}
		return nil
	})
}

func (t *chandeMomentumOscillatorTransformation) do(n int, cr flux.ColReader, c flux.ColMeta, builder execute.TableBuilder, j int) error {
	var sumUp float64
	var sumDown float64
	sumUp = 0
	sumDown = 0

	switch c.Type {
	case flux.TInt:
		arrValues := cr.Ints(j)
		c.Type = flux.TFloat
		prev := arrValues.Value(0)
		curr := int64(0)
		for i := 0; i < cr.Len(); i++ {
			curr = arrValues.Value(i)
			diff := float64(curr - prev)
			if diff > 0 {
				sumUp += diff
			} else if diff < 0 {
				sumDown -= diff
			}

			if i >= n {
				val := 100 * (sumUp - sumDown) / (sumUp + sumDown)
				if err := builder.AppendFloat(j, val); err != nil {
					return err
				}

				diffNAgo := float64(arrValues.Value(i-n+1) - arrValues.Value(i-n))
				if diffNAgo > 0 {
					sumUp -= diffNAgo
				} else if diffNAgo < 0 {
					sumDown += diffNAgo
				}
			}

			prev = curr
		}
	case flux.TUInt:
		arrValues := cr.UInts(j)
		c.Type = flux.TFloat
		prev := arrValues.Value(0)
		curr := uint64(0)
		for i := 0; i < cr.Len(); i++ {
			curr = arrValues.Value(i)
			diff := float64(curr - prev)
			if diff > 0 {
				sumUp += diff
			} else if diff < 0 {
				sumDown -= diff
			}

			if i >= n {
				val := 100 * (sumUp - sumDown) / (sumUp + sumDown)
				if err := builder.AppendFloat(j, val); err != nil {
					return err
				}

				diffNAgo := float64(arrValues.Value(i-n+1) - arrValues.Value(i-n))
				if diffNAgo > 0 {
					sumUp -= diffNAgo
				} else if diffNAgo < 0 {
					sumDown += diffNAgo
				}
			}

			prev = curr
		}
	case flux.TFloat:
		arrValues := cr.Floats(j)
		prev := arrValues.Value(0)
		curr := 0.0
		for i := 0; i < cr.Len(); i++ {
			curr = arrValues.Value(i)
			diff := float64(curr - prev)
			if diff > 0 {
				sumUp += diff
			} else if diff < 0 {
				sumDown -= diff
			}

			if i >= n {
				val := 100 * (sumUp - sumDown) / (sumUp + sumDown)
				if err := builder.AppendFloat(j, val); err != nil {
					return err
				}

				diffNAgo := float64(arrValues.Value(i-n+1) - arrValues.Value(i-n))
				if diffNAgo > 0 {
					sumUp -= diffNAgo
				} else if diffNAgo < 0 {
					sumDown += diffNAgo
				}
			}

			prev = curr
		}
	}

	return nil
}

func (t *chandeMomentumOscillatorTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}

func (t *chandeMomentumOscillatorTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}

func (t *chandeMomentumOscillatorTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}
