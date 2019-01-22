package universe

import (
	"fmt"
	"time"

	"github.com/apache/arrow/go/arrow/array"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
)

const DerivativeKind = "derivative"

type DerivativeOpSpec struct {
	Unit        flux.Duration `json:"unit"`
	NonNegative bool          `json:"nonNegative"`
	Columns     []string      `json:"columns"`
	TimeColumn  string        `json:"timeColumn"`
}

func init() {
	derivativeSignature := flux.FunctionSignature(
		map[string]semantic.PolyType{
			"unit":        semantic.Duration,
			"nonNegative": semantic.Bool,
			"columns":     semantic.NewArrayPolyType(semantic.String),
			"timeColumn":  semantic.String,
		},
		nil,
	)

	flux.RegisterPackageValue("universe", DerivativeKind, flux.FunctionValue(DerivativeKind, createDerivativeOpSpec, derivativeSignature))
	flux.RegisterOpSpec(DerivativeKind, newDerivativeOp)
	plan.RegisterProcedureSpec(DerivativeKind, newDerivativeProcedure, DerivativeKind)
	execute.RegisterTransformation(DerivativeKind, createDerivativeTransformation)
}

func createDerivativeOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(DerivativeOpSpec)

	if unit, ok, err := args.GetDuration("unit"); err != nil {
		return nil, err
	} else if ok {
		spec.Unit = unit
	} else {
		//Default is 1s
		spec.Unit = flux.Duration(time.Second)
	}

	if nn, ok, err := args.GetBool("nonNegative"); err != nil {
		return nil, err
	} else if ok {
		spec.NonNegative = nn
	}
	if timeCol, ok, err := args.GetString("timeColumn"); err != nil {
		return nil, err
	} else if ok {
		spec.TimeColumn = timeCol
	} else {
		spec.TimeColumn = execute.DefaultTimeColLabel
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

func newDerivativeOp() flux.OperationSpec {
	return new(DerivativeOpSpec)
}

func (s *DerivativeOpSpec) Kind() flux.OperationKind {
	return DerivativeKind
}

type DerivativeProcedureSpec struct {
	plan.DefaultCost
	Unit        flux.Duration `json:"unit"`
	NonNegative bool          `json:"non_negative"`
	Columns     []string      `json:"columns"`
	TimeColumn  string        `json:"timeColumn"`
}

func newDerivativeProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*DerivativeOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	return &DerivativeProcedureSpec{
		Unit:        spec.Unit,
		NonNegative: spec.NonNegative,
		Columns:     spec.Columns,
		TimeColumn:  spec.TimeColumn,
	}, nil
}

func (s *DerivativeProcedureSpec) Kind() plan.ProcedureKind {
	return DerivativeKind
}
func (s *DerivativeProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(DerivativeProcedureSpec)
	*ns = *s
	if s.Columns != nil {
		ns.Columns = make([]string, len(s.Columns))
		copy(ns.Columns, s.Columns)
	}
	return ns
}

func createDerivativeTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*DerivativeProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", spec)
	}
	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)
	t := NewDerivativeTransformation(d, cache, s)
	return t, d, nil
}

type derivativeTransformation struct {
	d     execute.Dataset
	cache execute.TableBuilderCache

	unit        float64
	nonNegative bool
	columns     []string
	timeCol     string
}

func NewDerivativeTransformation(d execute.Dataset, cache execute.TableBuilderCache, spec *DerivativeProcedureSpec) *derivativeTransformation {
	return &derivativeTransformation{
		d:           d,
		cache:       cache,
		unit:        float64(spec.Unit),
		nonNegative: spec.NonNegative,
		columns:     spec.Columns,
		timeCol:     spec.TimeColumn,
	}
}

func (t *derivativeTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *derivativeTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	builder, created := t.cache.TableBuilder(tbl.Key())
	if !created {
		return fmt.Errorf("derivative found duplicate table with key: %v", tbl.Key())
	}
	cols := tbl.Cols()
	doDerivative := make([]bool, len(cols))
	timeIdx := -1
	for j, c := range cols {
		found := false
		for _, label := range t.columns {
			if c.Label == label {
				found = true
				break
			}
		}
		if c.Label == t.timeCol {
			timeIdx = j
		}

		if found {
			dc := c
			// Derivative always results in a float
			dc.Type = flux.TFloat
			_, err := builder.AddCol(dc)
			if err != nil {
				return err
			}
			doDerivative[j] = true
		} else {
			_, err := builder.AddCol(c)
			if err != nil {
				return err
			}
		}
	}
	if timeIdx < 0 {
		return fmt.Errorf("no column %q exists", t.timeCol)
	}

	return tbl.Do(func(cr flux.ColReader) error {
		if cr.Len() == 0 {
			return nil
		}

		for j, c := range cr.Cols() {
			var err error
			switch c.Type {
			case flux.TBool:
				err = t.passThroughBool(cr.Times(timeIdx), cr.Bools(j), builder, j)
			case flux.TInt:
				err = t.doInt(cr.Times(timeIdx), cr.Ints(j), builder, j, doDerivative[j])
			case flux.TUInt:
				err = t.doUInt(cr.Times(timeIdx), cr.UInts(j), builder, j, doDerivative[j])
			case flux.TFloat:
				err = t.doFloat(cr.Times(timeIdx), cr.Floats(j), builder, j, doDerivative[j])
			case flux.TString:
				err = t.passThroughString(cr.Times(timeIdx), cr.Strings(j), builder, j)
			case flux.TTime:
				err = t.passThroughTime(cr.Times(timeIdx), cr.Times(j), builder, j)
			}

			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (t *derivativeTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}
func (t *derivativeTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}
func (t *derivativeTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}

func (t *derivativeTransformation) passThroughBool(ts *array.Int64, vs *array.Boolean, b execute.TableBuilder, bj int) error {
	i := 0

	// Consume the first input value, which doesn't produce an output value
	if ts.IsNull(i) {
		return fmt.Errorf("derivative found null time in time column")
	}
	pTime := execute.Time(ts.Value(i))
	i++

	// Process the rest of the rows
	l := vs.Len()
	for ; i < l; i++ {
		if ts.IsNull(i) {
			return fmt.Errorf("derivative found null time in time column")
		}

		cTime := execute.Time(ts.Value(i))
		if cTime < pTime {
			return fmt.Errorf("derivative found out-of-order times in time column")
		}

		if cTime == pTime {
			// Only use the first value found if a time value is the same as
			// the previous row.
			continue
		}

		// We have a valid time for this row.  Code below should not exit
		// the loop early so pTime can be set for the next iteration.

		if vs.IsValid(i) {
			if err := b.AppendBool(bj, vs.Value(i)); err != nil {
				return err
			}
		} else {
			if err := b.AppendNil(bj); err != nil {
				return err
			}
		}

		pTime = cTime
	}

	return nil
}

func (t *derivativeTransformation) doInt(ts, vs *array.Int64, b execute.TableBuilder, bj int, doDerivative bool) error {
	i := 0

	if ts.IsNull(i) {
		return fmt.Errorf("derivative found null time in time column")
	}

	var pValue int64
	var pValueTime execute.Time
	validPValue := false

	// Now consume the first input value, which doesn't produce an output value
	pTime := execute.Time(ts.Value(i))
	if vs.IsValid(i) {
		pValue = vs.Value(i)
		pValueTime = pTime
		validPValue = true
	}
	i++

	// Process the rest of the rows
	l := vs.Len()
	for ; i < l; i++ {
		if ts.IsNull(i) {
			return fmt.Errorf("derivative found null time in time column")
		}

		cTime := execute.Time(ts.Value(i))
		if cTime < pTime {
			return fmt.Errorf("derivative found out-of-order times in time column")
		}

		if cTime == pTime {
			// if time did not increase with this row, ignore it.
			continue
		}

		// We have a valid time for this row.  Code below should not exit
		// the loop early so pTime can be set for the next iteration.

		if !doDerivative {
			// Just write the value to the builder
			if vs.IsValid(i) {
				if err := b.AppendInt(bj, vs.Value(i)); err != nil {
					return err
				}
			} else {
				if err := b.AppendNil(bj); err != nil {
					return err
				}
			}
		} else if !validPValue {
			// We have not yet seen a valid value.
			if err := b.AppendNil(bj); err != nil {
				return err
			}

			if vs.IsValid(i) {
				pValue = vs.Value(i)
				pValueTime = cTime
				validPValue = true
			}
		} else if vs.IsNull(i) {
			// If current value is null, then produce null value
			if err := b.AppendNil(bj); err != nil {
				return err
			}
		} else {
			// We have a valid previous value and current value.
			cValue := vs.Value(i)
			if t.nonNegative && pValue > cValue {
				if err := b.AppendNil(bj); err != nil {
					return err
				}
			} else {
				// Finally, do the derivative.
				elapsed := float64(cTime-pValueTime) / t.unit
				diff := float64(cValue - pValue)
				if err := b.AppendFloat(bj, diff/elapsed); err != nil {
					return err
				}
			}

			pValue = cValue
			pValueTime = cTime
		}

		pTime = cTime
	}

	return nil
}

func (t *derivativeTransformation) doUInt(ts *array.Int64, vs *array.Uint64, b execute.TableBuilder, bj int, doDerivative bool) error {
	i := 0

	if ts.IsNull(i) {
		return fmt.Errorf("derivative found null time in time column")
	}

	var pValue uint64
	var pValueTime execute.Time
	validPValue := false

	// Now consume the first input value, which doesn't produce an output value
	pTime := execute.Time(ts.Value(i))
	if vs.IsValid(i) {
		pValue = vs.Value(i)
		pValueTime = pTime
		validPValue = true
	}
	i++

	// Process the rest of the rows
	l := vs.Len()
	for ; i < l; i++ {
		if ts.IsNull(i) {
			return fmt.Errorf("derivative found null time in time column")
		}

		cTime := execute.Time(ts.Value(i))
		if cTime < pTime {
			return fmt.Errorf("derivative found out-of-order times in time column")
		}

		if cTime == pTime {
			// if time did not increase with this row, ignore it.
			continue
		}

		// We have a valid time for this row.  Code below should not exit
		// the loop early so pTime can be set for the next iteration.

		if !doDerivative {
			// Just write the value to the builder
			if vs.IsValid(i) {
				if err := b.AppendUInt(bj, vs.Value(i)); err != nil {
					return err
				}
			} else {
				if err := b.AppendNil(bj); err != nil {
					return err
				}
			}
		} else if !validPValue {
			// We have not yet seen a valid value.
			if err := b.AppendNil(bj); err != nil {
				return err
			}

			if vs.IsValid(i) {
				pValue = vs.Value(i)
				pValueTime = cTime
				validPValue = true
			}
		} else if vs.IsNull(i) {
			// If current value is null, then produce null value
			if err := b.AppendNil(bj); err != nil {
				return err
			}
		} else {
			// We have a valid previous value and current value.
			cValue := vs.Value(i)
			isNeg := pValue > cValue
			if t.nonNegative && isNeg {
				if err := b.AppendNil(bj); err != nil {
					return err
				}
			} else {
				// Finally, do the derivative.
				elapsed := float64(cTime-pValueTime) / t.unit

				var diff float64
				if isNeg {
					// Avoid wrapping on unsigned subtraction
					diff = -float64(pValue - cValue)
				} else {
					diff = float64(cValue - pValue)
				}

				if err := b.AppendFloat(bj, diff/elapsed); err != nil {
					return err
				}
			}

			pValue = cValue
			pValueTime = cTime
		}

		pTime = cTime
	}

	return nil
}

func (t *derivativeTransformation) doFloat(ts *array.Int64, vs *array.Float64, b execute.TableBuilder, bj int, doDerivative bool) error {
	i := 0

	if ts.IsNull(i) {
		return fmt.Errorf("derivative found null time in time column")
	}

	var pValue float64
	var pValueTime execute.Time
	validPValue := false

	// Now consume the first input value, which doesn't produce an output value
	pTime := execute.Time(ts.Value(i))
	if vs.IsValid(i) {
		pValue = vs.Value(i)
		pValueTime = pTime
		validPValue = true
	}
	i++

	// Process the rest of the rows
	l := vs.Len()
	for ; i < l; i++ {
		if ts.IsNull(i) {
			return fmt.Errorf("derivative found null time in time column")
		}

		cTime := execute.Time(ts.Value(i))
		if cTime < pTime {
			return fmt.Errorf("derivative found out-of-order times in time column")
		}

		if cTime == pTime {
			// if time did not increase with this row, ignore it.
			continue
		}

		// We have a valid time for this row.  Code below should not exit
		// the loop early so pTime can be set for the next iteration.

		if !doDerivative {
			// Just write the value to the builder
			if vs.IsValid(i) {
				if err := b.AppendFloat(bj, vs.Value(i)); err != nil {
					return err
				}
			} else {
				if err := b.AppendNil(bj); err != nil {
					return err
				}
			}
		} else if !validPValue {
			// We have not yet seen a valid value.
			if err := b.AppendNil(bj); err != nil {
				return err
			}

			if vs.IsValid(i) {
				pValue = vs.Value(i)
				pValueTime = cTime
				validPValue = true
			}
		} else if vs.IsNull(i) {
			// If current value is null, then produce null value
			if err := b.AppendNil(bj); err != nil {
				return err
			}
		} else {
			// We have a valid previous value and current value.
			cValue := vs.Value(i)
			if t.nonNegative && pValue > cValue {
				if err := b.AppendNil(bj); err != nil {
					return err
				}
			} else {
				// Finally, do the derivative.
				elapsed := float64(cTime-pValueTime) / t.unit
				diff := float64(cValue - pValue)
				if err := b.AppendFloat(bj, diff/elapsed); err != nil {
					return err
				}
			}

			pValue = cValue
			pValueTime = cTime
		}

		pTime = cTime
	}

	return nil
}

func (t *derivativeTransformation) passThroughString(ts *array.Int64, vs *array.Binary, b execute.TableBuilder, bj int) error {
	i := 0

	// Consume the first input value, which doesn't produce an output value
	if ts.IsNull(i) {
		return fmt.Errorf("derivative found null time in time column")
	}
	pTime := execute.Time(ts.Value(i))
	i++

	// Process the rest of the rows
	l := vs.Len()
	for ; i < l; i++ {
		if ts.IsNull(i) {
			return fmt.Errorf("derivative found null time in time column")
		}

		cTime := execute.Time(ts.Value(i))
		if cTime < pTime {
			return fmt.Errorf("derivative found out-of-order times in time column")
		}

		if cTime == pTime {
			// Only use the first value found if a time value is the same as
			// the previous row.
			continue
		}

		// We have a valid time for this row.  Code below should not exit
		// the loop early so pTime can be set for the next iteration.

		if vs.IsValid(i) {
			if err := b.AppendString(bj, string(vs.Value(i))); err != nil {
				return err
			}
		} else {
			if err := b.AppendNil(bj); err != nil {
				return err
			}
		}

		pTime = cTime
	}

	return nil
}

func (t *derivativeTransformation) passThroughTime(ts *array.Int64, vs *array.Int64, b execute.TableBuilder, bj int) error {
	i := 0

	// Consume the first input value, which doesn't produce an output value
	if ts.IsNull(i) {
		return fmt.Errorf("derivative found null time in time column")
	}
	pTime := execute.Time(ts.Value(i))
	i++

	// Process the rest of the rows
	l := vs.Len()
	for ; i < l; i++ {
		if ts.IsNull(i) {
			return fmt.Errorf("derivative found null time in time column")
		}

		cTime := execute.Time(ts.Value(i))
		if cTime < pTime {
			return fmt.Errorf("derivative found out-of-order times in time column")
		}

		if cTime == pTime {
			// Only use the first value found if a time value is the same as
			// the previous row.
			continue
		}

		// We have a valid time for this row.  Code below should not exit
		// the loop early so pTime can be set for the next iteration.

		if vs.IsValid(i) {
			if err := b.AppendTime(bj, execute.Time(vs.Value(i))); err != nil {
				return err
			}
		} else {
			if err := b.AppendNil(bj); err != nil {
				return err
			}
		}

		pTime = cTime
	}

	return nil
}
