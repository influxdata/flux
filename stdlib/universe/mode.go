package universe

import (
	"sort"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/values"
)

const ModeKind = "mode"

type ModeOpSpec struct {
	Column string `json:"column"`
}

func init() {
	modeSignature := semantic.LookupBuiltInType("universe", "mode")

	flux.RegisterPackageValue("universe", ModeKind, flux.MustValue(flux.FunctionValue(ModeKind, createModeOpSpec, modeSignature)))
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
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
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
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
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
		return errors.Newf(codes.FailedPrecondition, "mode found duplicate table with key: %v", tbl.Key())
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

	srcColIdx := colIdx

	destColIdx, err := builder.AddCol(flux.ColMeta{
		Label: execute.DefaultValueColLabel,
		Type:  col.Type,
	})
	if err != nil {
		return err
	}

	if tbl.Key().HasCol(t.column) {
		j := execute.ColIdx(t.column, tbl.Key().Cols())

		if err := builder.AppendValue(destColIdx, tbl.Key().Value(j)); err != nil {
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

	switch col.Type {
	case flux.TBool:
		return tbl.Do(func(cr flux.ColReader) error {
			return t.doBool(cr, tbl, builder, srcColIdx, destColIdx)
		})
	case flux.TInt:
		return tbl.Do(func(cr flux.ColReader) error {
			return t.doInt(cr, tbl, builder, srcColIdx, destColIdx)
		})
	case flux.TUInt:
		return tbl.Do(func(cr flux.ColReader) error {
			return t.doUInt(cr, tbl, builder, srcColIdx, destColIdx)
		})
	case flux.TFloat:
		return tbl.Do(func(cr flux.ColReader) error {
			return t.doFloat(cr, tbl, builder, srcColIdx, destColIdx)
		})
	case flux.TString:
		return tbl.Do(func(cr flux.ColReader) error {
			return t.doString(cr, tbl, builder, srcColIdx, destColIdx)
		})
	case flux.TTime:
		return tbl.Do(func(cr flux.ColReader) error {
			return t.doTime(cr, tbl, builder, srcColIdx, destColIdx)
		})
	}

	if err := execute.AppendKeyValues(tbl.Key(), builder); err != nil {
		return err
	}

	return nil
}

func (t *modeTransformation) doString(cr flux.ColReader, tbl flux.Table, builder execute.TableBuilder, srcColIdx, destColIdx int) error {
	stringMode := make(map[string]int64)
	l := cr.Len()
	j := srcColIdx // execute.ColIdx(t.column, tbl.Cols())
	numEntries := 0

	// log all values in the map with the number of occurrences
	for i := 0; i < l; i++ {
		// if the value is null we skip it
		if cr.Strings(j).IsNull(i) {
			continue
		}
		v := cr.Strings(j).ValueString(i)
		stringMode[v]++
	}

	// find the mode by finding the value(s) with the most occurrences
	max, total := int64(0), int64(0)
	for val := range stringMode {
		if stringMode[val] > max {
			max, total = stringMode[val], 1
		} else if stringMode[val] == max {
			total++
		}
	}

	// if every value occurs the same number of times or all values are null, there is no mode
	// if len(stringMode) == 0, there are only nulls, so total == 0 also
	// if len(stringMode) == total, then every value occurs the same number of times
	if int64(len(stringMode)) == total {
		if err := builder.AppendNil(destColIdx); err != nil {
			return err
		}
		if err := execute.AppendKeyValues(tbl.Key(), builder); err != nil {
			return err
		}
		return nil
	}

	// slice to store the modes
	storedVals := make([]string, 0, total)
	for val := range stringMode {
		if stringMode[val] == max {
			storedVals = append(storedVals, val)
		}
	}

	// added the modes to the builder in sorted order
	sort.Strings(storedVals)
	for j := range storedVals {
		if err := builder.AppendString(destColIdx, storedVals[j]); err != nil {
			return err
		}
	}
	numEntries = len(storedVals)

	// append the values in the builder to the output
	for i := 0; i < numEntries; i++ {
		if err := execute.AppendKeyValues(tbl.Key(), builder); err != nil {
			return err
		}
	}

	return nil
}

func (t *modeTransformation) doBool(cr flux.ColReader, tbl flux.Table, builder execute.TableBuilder, srcColIdx, destColIdx int) error {
	boolMode := make(map[bool]int64)
	l := cr.Len()
	j := srcColIdx
	numEntries := 0
	for i := 0; i < l; i++ {
		if cr.Bools(j).IsNull(i) {
			continue
		}
		v := cr.Bools(j).Value(i)
		boolMode[v]++
	}

	max, total := int64(0), int64(0)
	for val := range boolMode {
		if boolMode[val] > max {
			max, total = boolMode[val], 1
		} else if boolMode[val] == max {
			total++
		}
	}

	if int64(len(boolMode)) == total {
		if err := builder.AppendNil(destColIdx); err != nil {
			return err
		}
		if err := execute.AppendKeyValues(tbl.Key(), builder); err != nil {
			return err
		}
		return nil
	}

	storedVals := make([]bool, 0, total)
	for val := range boolMode {
		if boolMode[val] == max {
			storedVals = append(storedVals, val)
		}
	}
	for j := range storedVals {
		if err := builder.AppendBool(destColIdx, storedVals[j]); err != nil {
			return err
		}
	}
	numEntries = 1

	for i := 0; i < numEntries; i++ {
		if err := execute.AppendKeyValues(tbl.Key(), builder); err != nil {
			return err
		}
	}
	return nil
}

func (t *modeTransformation) doInt(cr flux.ColReader, tbl flux.Table, builder execute.TableBuilder, srcColIdx, destColIdx int) error {
	intMode := make(map[int64]int64)
	l := cr.Len()
	j := srcColIdx
	numEntries := 0
	for i := 0; i < l; i++ {
		if cr.Ints(j).IsNull(i) {
			continue
		}
		v := cr.Ints(j).Value(i)
		intMode[v]++
	}

	max, total := int64(0), int64(0)
	for val := range intMode {
		if intMode[val] > max {
			max, total = intMode[val], 1
		} else if intMode[val] == max {
			total++
		}
	}

	if int64(len(intMode)) == total {
		if err := builder.AppendNil(destColIdx); err != nil {
			return err
		}
		if err := execute.AppendKeyValues(tbl.Key(), builder); err != nil {
			return err
		}
		return nil
	}

	storedVals := make([]int64, 0, total)
	for val := range intMode {
		if intMode[val] == max {
			storedVals = append(storedVals, val)
		}
	}
	sort.Slice(storedVals, func(i, j int) bool { return storedVals[i] < storedVals[j] })
	for j := range storedVals {
		if err := builder.AppendInt(destColIdx, storedVals[j]); err != nil {
			return err
		}
	}
	numEntries = len(storedVals)

	for i := 0; i < numEntries; i++ {
		if err := execute.AppendKeyValues(tbl.Key(), builder); err != nil {
			return err
		}
	}

	return nil
}

func (t *modeTransformation) doUInt(cr flux.ColReader, tbl flux.Table, builder execute.TableBuilder, srcColIdx, destColIdx int) error {
	uintMode := make(map[uint64]int64)
	l := cr.Len()
	j := srcColIdx
	numEntries := 0
	for i := 0; i < l; i++ {
		if cr.UInts(j).IsNull(i) {
			continue
		}
		v := cr.UInts(j).Value(i)
		uintMode[v]++
	}

	max, total := int64(0), int64(0)
	for val := range uintMode {
		if uintMode[val] > max {
			max, total = uintMode[val], 1
		} else if uintMode[val] == max {
			total++
		}
	}

	if int64(len(uintMode)) == total {
		if err := builder.AppendNil(destColIdx); err != nil {
			return err
		}
		if err := execute.AppendKeyValues(tbl.Key(), builder); err != nil {
			return err
		}
		return nil
	}

	storedVals := make([]uint64, 0, total)
	for val := range uintMode {
		if uintMode[val] == max {
			storedVals = append(storedVals, val)
		}
	}
	sort.Slice(storedVals, func(i, j int) bool { return storedVals[i] < storedVals[j] })
	for j := range storedVals {
		if err := builder.AppendUInt(destColIdx, storedVals[j]); err != nil {
			return err
		}
	}
	numEntries = len(storedVals)

	for i := 0; i < numEntries; i++ {
		if err := execute.AppendKeyValues(tbl.Key(), builder); err != nil {
			return err
		}
	}
	return nil
}

func (t *modeTransformation) doFloat(cr flux.ColReader, tbl flux.Table, builder execute.TableBuilder, srcColIdx, destColIdx int) error {
	floatMode := make(map[float64]int64)
	l := cr.Len()
	j := srcColIdx
	numEntries := 0
	for i := 0; i < l; i++ {
		if cr.Floats(j).IsNull(i) {
			continue
		}
		v := cr.Floats(j).Value(i)
		floatMode[v]++
	}

	max, total := int64(0), int64(0)
	for val := range floatMode {
		if floatMode[val] > max {
			max, total = floatMode[val], 1
		} else if floatMode[val] == max {
			total++
		}
	}

	if int64(len(floatMode)) == total {
		if err := builder.AppendNil(destColIdx); err != nil {
			return err
		}
		if err := execute.AppendKeyValues(tbl.Key(), builder); err != nil {
			return err
		}
		return nil
	}

	storedVals := make([]float64, 0, total)
	for val := range floatMode {
		if floatMode[val] == max {
			storedVals = append(storedVals, val)
		}
	}
	sort.Float64s(storedVals)
	for j := range storedVals {
		if err := builder.AppendFloat(destColIdx, storedVals[j]); err != nil {
			return err
		}
	}
	numEntries = len(storedVals)

	for i := 0; i < numEntries; i++ {
		if err := execute.AppendKeyValues(tbl.Key(), builder); err != nil {
			return err
		}
	}

	return nil
}

func (t *modeTransformation) doTime(cr flux.ColReader, tbl flux.Table, builder execute.TableBuilder, srcColIdx, destColIdx int) error {
	timeMode := make(map[execute.Time]int64)
	l := cr.Len()
	j := srcColIdx
	numEntries := 0
	for i := 0; i < l; i++ {
		if cr.Times(j).IsNull(i) {
			continue
		}
		v := values.Time(cr.Times(j).Value(i))
		timeMode[v]++
	}

	max, total := int64(0), int64(0)
	for val := range timeMode {
		if timeMode[val] > max {
			max, total = timeMode[val], 1
		} else if timeMode[val] == max {
			total++
		}
	}

	if int64(len(timeMode)) == total {
		if err := builder.AppendNil(destColIdx); err != nil {
			return err
		}
		if err := execute.AppendKeyValues(tbl.Key(), builder); err != nil {
			return err
		}
		return nil
	}

	storedVals := make([]execute.Time, 0, total)
	for val := range timeMode {
		if timeMode[val] == max {
			storedVals = append(storedVals, val)
		}
	}
	sort.Slice(storedVals, func(i, j int) bool { return storedVals[i] < storedVals[j] })
	for j := range storedVals {
		if err := builder.AppendTime(destColIdx, storedVals[j]); err != nil {
			return err
		}
	}
	numEntries = len(storedVals)

	for i := 0; i < numEntries; i++ {
		if err := execute.AppendKeyValues(tbl.Key(), builder); err != nil {
			return err
		}
	}

	return nil
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
