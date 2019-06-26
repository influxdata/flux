package universe

import (
	"fmt"

	"github.com/influxdata/flux/values"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
)

const ModeKind = "mode"

type ModeOpSpec struct {
	Column string `json:"column"`
}

func init() {
	modeSignature := flux.FunctionSignature(
		map[string]semantic.PolyType{
			"column": semantic.String,
		},
		nil,
	)

	flux.RegisterPackageValue("universe", ModeKind, flux.FunctionValue(ModeKind, createModeOpSpec, modeSignature))
	flux.RegisterOpSpec(ModeKind, newModeOp)
	plan.RegisterProcedureSpec(ModeKind, newModeProcedure, ModeKind)
	execute.RegisterTransformation(ModeKind, createModeTransformation)
}

func createModeOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(ModeOpSpec)

	if col, ok, err := args.GetString("column"); err != nil {
		return nil, err
	} else if ok {
		spec.Column = col
	} else {
		spec.Column = execute.DefaultValueColLabel
	}

	return spec, nil
}

func newModeOp() flux.OperationSpec {
	return new(ModeOpSpec)
}

func (s *ModeOpSpec) Kind() flux.OperationKind {
	return ModeKind
}

type ModeProcedureSpec struct {
	plan.DefaultCost
	Column string
}

func newModeProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*ModeOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	return &ModeProcedureSpec{
		Column: spec.Column,
	}, nil
}

func (s *ModeProcedureSpec) Kind() plan.ProcedureKind {
	return ModeKind
}
func (s *ModeProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(ModeProcedureSpec)

	*ns = *s

	return ns
}

// TriggerSpec implements plan.TriggerAwareProcedureSpec
func (s *ModeProcedureSpec) TriggerSpec() plan.TriggerSpec {
	return plan.NarrowTransformationTriggerSpec{}
}

func createModeTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*ModeProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", spec)
	}
	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)
	t := NewModeTransformation(d, cache, s)
	return t, d, nil
}

type modeTransformation struct {
	d     execute.Dataset
	cache execute.TableBuilderCache

	column string
}

func NewModeTransformation(d execute.Dataset, cache execute.TableBuilderCache, spec *ModeProcedureSpec) *modeTransformation {
	return &modeTransformation{
		d:      d,
		cache:  cache,
		column: spec.Column,
	}
}

func (t *modeTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *modeTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	builder, created := t.cache.TableBuilder(tbl.Key())
	if !created {
		return fmt.Errorf("mode found duplicate table with key: %v", tbl.Key())
	}

	colIdx := execute.ColIdx(t.column, tbl.Cols())
	if colIdx < 0 {
		// doesn't exist in this table, so add an empty value
		if err := execute.AddTableKeyCols(tbl.Key(), builder); err != nil {
			return err
		}
		colIdx, err := builder.AddCol(flux.ColMeta{
			Label: execute.DefaultValueColLabel,
			Type:  flux.TString,
		})
		if err != nil {
			return err
		}

		if err := builder.AppendString(colIdx, ""); err != nil {
			return err
		}
		if err := execute.AppendKeyValues(tbl.Key(), builder); err != nil {
			return err
		}
		// TODO: hack required to ensure data flows downstream
		return tbl.Do(func(flux.ColReader) error {
			return nil
		})
	}

	col := tbl.Cols()[colIdx]

	if err := execute.AddTableKeyCols(tbl.Key(), builder); err != nil {
		return err
	}
	colIdx, err := builder.AddCol(flux.ColMeta{
		Label: execute.DefaultValueColLabel,
		Type:  col.Type,
	})
	if err != nil {
		return err
	}

	if tbl.Key().HasCol(t.column) {
		j := execute.ColIdx(t.column, tbl.Key().Cols())
		switch col.Type {
		case flux.TBool:
			if err := builder.AppendBool(colIdx, tbl.Key().ValueBool(j)); err != nil {
				return err
			}
		case flux.TInt:
			if err := builder.AppendInt(colIdx, tbl.Key().ValueInt(j)); err != nil {
				return err
			}
		case flux.TUInt:
			if err := builder.AppendUInt(colIdx, tbl.Key().ValueUInt(j)); err != nil {
				return err
			}
		case flux.TFloat:
			if err := builder.AppendFloat(colIdx, tbl.Key().ValueFloat(j)); err != nil {
				return err
			}
		case flux.TString:
			if err := builder.AppendString(colIdx, tbl.Key().ValueString(j)); err != nil {
				return err
			}
		case flux.TTime:
			if err := builder.AppendTime(colIdx, tbl.Key().ValueTime(j)); err != nil {
				return err
			}
		}

		if err := execute.AppendKeyValues(tbl.Key(), builder); err != nil {
			return err
		}
		// TODO: hack required to ensure data flows downstream
		return tbl.Do(func(flux.ColReader) error {
			return nil
		})
	}

	var (
		//nullMode   bool
		boolMode   map[bool]int64
		intMode    map[int64]int64
		uintMode   map[uint64]int64
		floatMode  map[float64]int64
		stringMode map[string]int64
		timeMode   map[execute.Time]int64
	)
	switch col.Type {
	case flux.TBool:
		boolMode = make(map[bool]int64)
	case flux.TInt:
		intMode = make(map[int64]int64)
	case flux.TUInt:
		uintMode = make(map[uint64]int64)
	case flux.TFloat:
		floatMode = make(map[float64]int64)
	case flux.TString:
		stringMode = make(map[string]int64)
	case flux.TTime:
		timeMode = make(map[execute.Time]int64)
	}

	j := execute.ColIdx(t.column, tbl.Cols())
	return tbl.Do(func(cr flux.ColReader) error {
		l := cr.Len()

		for i := 0; i < l; i++ {
			// Check mode
			switch col.Type {
			case flux.TBool:
				/*
				if cr.Bools(j).IsNull(i) {
					if !nullMode {
						if err := builder.AppendNil(colIdx); err != nil {
							return err
						}
						nullMode = true
					}
					continue
				}
				*/
				v := cr.Bools(j).Value(i)
				boolMode[v]++
			case flux.TInt:
				/*
				if cr.Ints(j).IsNull(i) {
					if !nullMode {
						if err := builder.AppendNil(colIdx); err != nil {
							return err
						}
						nullMode = true
					}
					continue
				}
				*/
				v := cr.Ints(j).Value(i)
				intMode[v]++
			case flux.TUInt:
				/*
				if cr.UInts(j).IsNull(i) {
					if !nullMode {
						if err := builder.AppendNil(colIdx); err != nil {
							return err
						}
						nullMode = true
					}
					continue
				}
				*/
				v := cr.UInts(j).Value(i)
				uintMode[v]++
			case flux.TFloat:
				/*
				if cr.Floats(j).IsNull(i) {
					if !nullMode {
						if err := builder.AppendNil(colIdx); err != nil {
							return err
						}
						nullMode = true
					}
					continue
				}
				*/
				v := cr.Floats(j).Value(i)
				floatMode[v]++
			case flux.TString:
				/*
				if cr.Strings(j).IsNull(i) {
					if !nullMode {
						if err := builder.AppendNil(colIdx); err != nil {
							return err
						}
						nullMode = true
					}
					continue
				}
				*/
				v := cr.Strings(j).ValueString(i)
				stringMode[v]++
			case flux.TTime:
				/*
				if cr.Times(j).IsNull(i) {
					if !nullMode {
						if err := builder.AppendNil(colIdx); err != nil {
							return err
						}
						nullMode = true
					}
					continue
				}
				*/
				v := values.Time(cr.Times(j).Value(i))
				timeMode[v]++
			}
		}
		// Find mode
		switch col.Type {
		case flux.TBool:
			v := false
			n := int64(0)
			for k := range boolMode {
				if boolMode[k] > n {
					v = k
					n = boolMode[k]
				}
			}
			if err := builder.AppendBool(colIdx, v); err != nil {
				return err
			}
		case flux.TInt:
			v := int64(0)
			n := int64(0)
			for k := range intMode {
				if intMode[k] > n {
					v = k
					n = intMode[k]
				}
			}
			if err := builder.AppendInt(colIdx, v); err != nil {
				return err
			}
		case flux.TUInt:
			v := uint64(0)
			n := int64(0)
			for k := range uintMode {
				if uintMode[k] > n {
					v = k
					n = uintMode[k]
				}
			}
			if err := builder.AppendUInt(colIdx, v); err != nil {
				return err
			}
		case flux.TFloat:
			v := float64(0)
			n := int64(0)
			for k := range floatMode {
				if floatMode[k] > n {
					v = k
					n = floatMode[k]
				}
			}
			if err := builder.AppendFloat(colIdx, v); err != nil {
				return err
			}
		case flux.TString:
			/*
			* In the test file mode_test.go, the last test should result in the mode being nil.
			* Here, I can't initialize v to be nil since I'll need to eventually AppendString(coldIdx, v),
			* which requires a valid string to append (it cannot take nil as an argument). 
			*/
			v := ""
			n := int64(0)
			for k := range stringMode {
				if stringMode[k] > n {
					v = k
					n = stringMode[k]
				}
			}
			if err := builder.AppendString(colIdx, v); err != nil {
				return err
			}
		case flux.TTime:
			v := execute.Time(0)
			n := int64(0)
			timeMode[v]++
			for k := range timeMode {
				if timeMode[k] > n {
					v = k
					n = timeMode[k]
				}
			}
			if err := builder.AppendTime(colIdx, v); err != nil {
				return err
			}
		}

		if err := execute.AppendKeyValues(tbl.Key(), builder); err != nil {
			return err
		}

		return nil
	})
}

func (t *modeTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}
func (t *modeTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}
func (t *modeTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}
