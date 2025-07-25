package universe

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	arrowmemory "github.com/apache/arrow-go/v18/arrow/memory"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/table"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

//go:generate -command tmpl ../../gotool.sh github.com/benbjohnson/tmpl
//go:generate tmpl -data=@../../internal/types.tmpldata -o pivot.gen.go pivot.gen.go.tmpl

const (
	PivotKind       = "pivot"
	SortedPivotKind = "sortedPivot"
	nullValueLabel  = "null"
)

type PivotOpSpec struct {
	RowKey      []string `json:"rowKey"`
	ColumnKey   []string `json:"columnKey"`
	ValueColumn string   `json:"valueColumn"`
}

func init() {
	// default pivot
	pivotSignature := runtime.MustLookupBuiltinType("universe", "pivot")

	runtime.RegisterPackageValue("universe", PivotKind, flux.MustValue(flux.FunctionValue(PivotKind, createPivotOpSpec, pivotSignature)))

	plan.RegisterProcedureSpec(PivotKind, newPivotProcedure, PivotKind)
	execute.RegisterTransformation(PivotKind, createPivotTransformation)

	// optimized pivot
	execute.RegisterTransformation(SortedPivotKind, createSortedPivotTransformation)
}

func createPivotOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := &PivotOpSpec{}

	array, err := args.GetRequiredArray("rowKey", semantic.String)
	if err != nil {
		return nil, err
	}

	spec.RowKey, err = interpreter.ToStringArray(array)
	if err != nil {
		return nil, err
	}

	array, err = args.GetRequiredArray("columnKey", semantic.String)
	if err != nil {
		return nil, err
	}

	spec.ColumnKey, err = interpreter.ToStringArray(array)
	if err != nil {
		return nil, err
	}

	rowKeys := make(map[string]bool)
	for _, v := range spec.RowKey {
		rowKeys[v] = true
	}

	for _, v := range spec.ColumnKey {
		if _, ok := rowKeys[v]; ok {
			return nil, errors.Newf(codes.Invalid, "column name found in both rowKey and colKey: %s", v)
		}
	}

	valueCol, err := args.GetRequiredString("valueColumn")
	if err != nil {
		return nil, err
	}
	spec.ValueColumn = valueCol

	return spec, nil
}

func (s *PivotOpSpec) Kind() flux.OperationKind {
	return PivotKind
}

type PivotProcedureSpec struct {
	plan.DefaultCost
	RowKey      []string
	ColumnKey   []string
	ValueColumn string

	// IsSortedByFunc is a function that can be set by the planner
	// that can be used to determine if the parent is sorted by
	// the given columns.
	// TODO(jsternberg): See https://github.com/influxdata/flux/issues/2131 for details.
	IsSortedByFunc func(cols []string, desc bool) bool

	// IsKeyColumnFunc is a function that can be set by the planner
	// that can be used to determine if the given column would be
	// part of the group key if it were present.
	// TODO(jsternberg): See https://github.com/influxdata/flux/issues/2131 for details.
	IsKeyColumnFunc func(label string) bool
}

func newPivotProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*PivotOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}

	p := &PivotProcedureSpec{
		RowKey:      spec.RowKey,
		ColumnKey:   spec.ColumnKey,
		ValueColumn: spec.ValueColumn,
	}

	return p, nil
}

func (s *PivotProcedureSpec) Kind() plan.ProcedureKind {
	return PivotKind
}
func (s *PivotProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(PivotProcedureSpec)
	ns.RowKey = make([]string, len(s.RowKey))
	copy(ns.RowKey, s.RowKey)
	ns.ColumnKey = make([]string, len(s.ColumnKey))
	copy(ns.ColumnKey, s.ColumnKey)
	ns.ValueColumn = s.ValueColumn
	return ns
}

func createPivotTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*PivotProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}

	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)
	t := NewPivotTransformation(d, cache, s)
	return t, d, nil
}

type rowCol struct {
	nextCol int
	nextRow int
}

type pivotTransformation struct {
	execute.ExecutionNode
	d     execute.Dataset
	cache execute.TableBuilderCache
	spec  PivotProcedureSpec
	// for each table, we need to store a map to keep track of which rows/columns have already been created.
	colKeyMaps map[string]map[string]int
	rowKeyMaps map[string]map[string]int
	nextRowCol map[string]rowCol
}

func NewPivotTransformation(d execute.Dataset, cache execute.TableBuilderCache, spec *PivotProcedureSpec) *pivotTransformation {
	t := &pivotTransformation{
		d:          d,
		cache:      cache,
		spec:       *spec,
		colKeyMaps: make(map[string]map[string]int),
		rowKeyMaps: make(map[string]map[string]int),
		nextRowCol: make(map[string]rowCol),
	}
	return t
}

func (t *pivotTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *pivotTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	rowKeyIndex := make(map[string]int)
	for _, v := range t.spec.RowKey {
		idx := execute.ColIdx(v, tbl.Cols())
		if idx < 0 {
			return errors.Newf(codes.Invalid, "specified row key column does not exist in table: %v", v)
		}
		rowKeyIndex[v] = idx
	}

	// different from above because we'll get the column indices below when we
	// determine the initial column schema
	colKeyIndex := make(map[string]int)
	valueColIndex := -1
	var valueColType flux.ColType
	for _, v := range t.spec.ColumnKey {
		colKeyIndex[v] = -1
	}

	cols := make([]flux.ColMeta, 0, len(tbl.Cols()))
	keyCols := make([]flux.ColMeta, 0, len(tbl.Key().Cols()))
	keyValues := make([]values.Value, 0, len(tbl.Key().Cols()))
	newIDX := 0
	colMap := make([]int, len(tbl.Cols()))

	for colIDX, v := range tbl.Cols() {
		if _, ok := colKeyIndex[v.Label]; !ok && v.Label != t.spec.ValueColumn {
			// the columns we keep are: group key columns not in the column key and row key columns
			if tbl.Key().HasCol(v.Label) {
				colMap[newIDX] = colIDX
				newIDX++
				keyCols = append(keyCols, tbl.Cols()[colIDX])
				cols = append(cols, tbl.Cols()[colIDX])
				keyValues = append(keyValues, tbl.Key().LabelValue(v.Label))
			} else if _, ok := rowKeyIndex[v.Label]; ok {
				cols = append(cols, tbl.Cols()[colIDX])
				colMap[newIDX] = colIDX
				newIDX++
			}
		} else if v.Label == t.spec.ValueColumn {
			valueColIndex = colIDX
			valueColType = tbl.Cols()[colIDX].Type
		} else {
			// we need the location of the colKey columns in the original table
			colKeyIndex[v.Label] = colIDX
		}
	}

	if valueColIndex < 0 {
		return errors.Newf(codes.Invalid, "specified value column does not exist in table: %v", t.spec.ValueColumn)
	}

	for k, v := range colKeyIndex {
		if v < 0 {
			return errors.Newf(codes.Invalid, "specified column does not exist in table: %v", k)
		}
	}

	newGroupKey := execute.NewGroupKey(keyCols, keyValues)
	builder, created := t.cache.TableBuilder(newGroupKey)
	groupKeyString := newGroupKey.String()
	if created {
		for _, c := range cols {
			_, err := builder.AddCol(c)
			if err != nil {
				return err
			}

		}
		t.colKeyMaps[groupKeyString] = make(map[string]int)
		t.rowKeyMaps[groupKeyString] = make(map[string]int)
		t.nextRowCol[groupKeyString] = rowCol{nextCol: len(cols), nextRow: 0}
	}

	return tbl.Do(func(cr flux.ColReader) error {
		for row := 0; row < cr.Len(); row++ {
			var rkb strings.Builder
			for _, rk := range t.spec.RowKey {
				j := rowKeyIndex[rk]
				c := cr.Cols()[j]
				rkb.WriteString(valueToStr(cr, c, row, j))
			}
			rowKey := rkb.String()

			var ckb strings.Builder
			for _, ck := range t.spec.ColumnKey {
				j := colKeyIndex[ck]
				c := cr.Cols()[j]
				if ckb.Len() > 0 {
					ckb.WriteByte('_')
				}
				ckb.WriteString(valueToStr(cr, c, row, j))
			}
			colKey := ckb.String()

			// we have columns for the copy-over in place;
			// we know the row key;
			// we know the col key;
			//  0.  If we've not seen the colKey before, then we need to add a new column and backfill it.
			if _, ok := t.colKeyMaps[groupKeyString][colKey]; !ok {
				newCol := flux.ColMeta{
					Label: colKey,
					Type:  valueColType,
				}
				nextCol, err := builder.AddCol(newCol)
				if err != nil {
					// column already exists
					return errors.Newf(
						codes.Invalid,
						"value %q appears in a column key column, but a column named %q already exists; consider renaming %q to something else before pivoting",
						colKey, colKey, colKey,
					)
				}
				t.colKeyMaps[groupKeyString][colKey] = nextCol
			}
			//  1.  if we've not seen rowKey before, then we need to append a new row, with copied values for the
			//  existing columns, as well as zero values for the pivoted columns.
			if _, ok := t.rowKeyMaps[groupKeyString][rowKey]; !ok {
				// rowkey U groupKey cols
				for cidx := range cols {
					if err := builder.AppendValue(cidx, execute.ValueForRow(cr, row, colMap[cidx])); err != nil {
						return err
					}
				}

				// zero-out the known key columns we've already discovered.
				for _, v := range t.colKeyMaps[groupKeyString] {
					if err := growColumn(builder, v, 1); err != nil {
						return err
					}
				}
				nextRowCol := t.nextRowCol[groupKeyString]
				t.rowKeyMaps[groupKeyString][rowKey] = nextRowCol.nextRow
				nextRowCol.nextRow++
				t.nextRowCol[groupKeyString] = nextRowCol
			}

			// at this point, we've created, added and back-filled all the columns we know about
			// if we found a new row key, we added a new row with zeroes set for all the value columns
			// so in all cases we know the row exists, and the column exists.  we need to grab the
			// value from valueCol and assign it to its pivoted position.
			if err := builder.SetValue(t.rowKeyMaps[groupKeyString][rowKey], t.colKeyMaps[groupKeyString][colKey], execute.ValueForRow(cr, row, valueColIndex)); err != nil {
				return err
			}

		}
		return nil
	})
}

func growColumn(builder execute.TableBuilder, colIdx, nRows int) error {
	colType := builder.Cols()[colIdx].Type
	switch colType {
	case flux.TBool:
		return builder.GrowBools(colIdx, nRows)
	case flux.TInt:
		return builder.GrowInts(colIdx, nRows)
	case flux.TUInt:
		return builder.GrowUInts(colIdx, nRows)
	case flux.TFloat:
		return builder.GrowFloats(colIdx, nRows)
	case flux.TString:
		return builder.GrowStrings(colIdx, nRows)
	case flux.TTime:
		return builder.GrowTimes(colIdx, nRows)
	default:
		execute.PanicUnknownType(colType)
		return errors.Newf(codes.Internal, "invalid column type: %s", colType)
	}
}

func valueToStr(cr flux.ColReader, c flux.ColMeta, row, col int) string {
	result := nullValueLabel

	switch c.Type {
	case flux.TBool:
		if v := cr.Bools(col); v.IsValid(row) {
			result = strconv.FormatBool(v.Value(row))
		}
	case flux.TInt:
		if v := cr.Ints(col); v.IsValid(row) {
			result = strconv.FormatInt(v.Value(row), 10)
		}
	case flux.TUInt:
		if v := cr.UInts(col); v.IsValid(row) {
			result = strconv.FormatUint(v.Value(row), 10)
		}
	case flux.TFloat:
		if v := cr.Floats(col); v.IsValid(row) {
			result = strconv.FormatFloat(v.Value(row), 'E', -1, 64)
		}
	case flux.TString:
		if v := cr.Strings(col); v.IsValid(row) {
			result = v.Value(row)
		}
	case flux.TTime:
		if v := cr.Times(col); v.IsValid(row) {
			result = values.Time(v.Value(row)).String()
		}
	default:
		execute.PanicUnknownType(c.Type)
	}

	return result
}

func (t *pivotTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}

func (t *pivotTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}

func (t *pivotTransformation) Finish(id execute.DatasetID, err error) {

	t.d.Finish(err)
}

type SortedPivotProcedureSpec struct {
	plan.DefaultCost
	RowKey      []string
	ColumnKey   []string
	ValueColumn string
}

func (s *SortedPivotProcedureSpec) Kind() plan.ProcedureKind {
	return SortedPivotKind
}

func (s *SortedPivotProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(SortedPivotProcedureSpec)
	ns.RowKey = make([]string, len(s.RowKey))
	copy(ns.RowKey, s.RowKey)
	ns.ColumnKey = make([]string, len(s.ColumnKey))
	copy(ns.ColumnKey, s.ColumnKey)
	ns.ValueColumn = s.ValueColumn
	return ns
}

// OutputAttributes implements the OutputAttributer interface used by the planner
// to keep track of various data attributes at different points in the plan.
// For sorted pivot, we know that the input data will be sorted on the row key, and that the
// output of this operation will preserve that order.
func (s *SortedPivotProcedureSpec) OutputAttributes() plan.PhysicalAttributes {
	return plan.PhysicalAttributes{
		plan.CollationKey: &plan.CollationAttr{Columns: s.RowKey},
	}
}

func createSortedPivotTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*SortedPivotProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}
	t, dataset, err := newSortedPivotTransformation(a.Context(), *s, id, a.Allocator())
	if err != nil {
		return nil, nil, err
	}
	return t, dataset, nil
}

// sortedPivotTransformation is an optimized version of pivot.
// It can only be used when there is a single row and column key
// and it can only be used if the row key is sorted without
// null values.
type sortedPivotTransformation struct {
	execute.ExecutionNode
	d     *execute.PassthroughDataset
	ctx   context.Context
	alloc memory.Allocator
	spec  SortedPivotProcedureSpec

	watermark  execute.Time
	processing execute.Time

	prevOutputKey flux.GroupKey
	group         *pivotTableGroup
}

func newSortedPivotTransformation(ctx context.Context, spec SortedPivotProcedureSpec, id execute.DatasetID, alloc memory.Allocator) (execute.Transformation, execute.Dataset, error) {
	t := &sortedPivotTransformation{
		d:     execute.NewPassthroughDataset(id),
		ctx:   ctx,
		alloc: alloc,
		spec:  spec,
	}
	return t, t.d, nil
}

func (t *sortedPivotTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *sortedPivotTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	// Validate that the table has all of the requisite columns.
	if err := t.validateTable(tbl); err != nil {
		return err
	}

	// Compute the group key that the current table group belongs to.
	// This is calculated by taking the current group key and removing
	// the column keys and the value column.
	// This can be calculated in advance because pivot will never
	// add columns to the group key so it is safe to compute without
	// the other tables.
	currOutputKey := t.computeGroupKey(tbl.Key())

	// Determine the indices of everything.
	rowIndex := execute.ColIdx(t.spec.RowKey[0], tbl.Cols())
	rowType := tbl.Cols()[rowIndex].Type
	valueIndex := execute.ColIdx(t.spec.ValueColumn, tbl.Cols())
	valueType := tbl.Cols()[valueIndex].Type

	// If this is the first group, create a pivotTableGroup and
	// instantiate buffer map
	if t.prevOutputKey == nil {
		t.group = &pivotTableGroup{
			rowCol: flux.ColMeta{
				Label: t.spec.RowKey[0],
				Type:  rowType,
			},
			buffers: make(map[string]*pivotTableBuffer),
		}

		// Assign comparison key for subsequent passthroughs
		t.prevOutputKey = currOutputKey
	}

	// If previous key is not the same as current key, the table has changed
	// And it is time to pivot and flush the stored buffers
	if !t.prevOutputKey.Equal(currOutputKey) {
		// pivot the current table buffers
		tbl, err := t.group.doPivot(t.prevOutputKey, t.alloc)
		if err != nil {
			return &flux.Error{
				Code: codes.Internal,
				Msg:  "Unable to pivot table data",
			}
		}

		if err = t.d.Process(tbl); err != nil {
			return &flux.Error{
				Code: codes.Internal,
				Msg:  "Unable to process table data",
			}
		}
		// Update the comparison key
		t.prevOutputKey = currOutputKey

		// Flush data by making a new table group with empty buffers
		t.group = &pivotTableGroup{
			rowCol: flux.ColMeta{
				Label: t.spec.RowKey[0],
				Type:  rowType,
			},
			buffers: make(map[string]*pivotTableBuffer),
		}
	}

	// Read the table and insert each of the column readers
	// into the table group.
	return tbl.Do(func(cr flux.ColReader) error {
		colKey := t.spec.ColumnKey[0]
		// getting the value of the `_field` col
		colLabel := cr.Key().LabelValue(colKey)

		if colLabel == nil {
			// The column key is not part of the group key
			// so it does not have a consistent value.
			// This is possible for us to do, but it requires
			// regrouping the input and that is not implemented yet.
			return errors.New(codes.Unimplemented, "column keys that are not part of the group currOutputKey are not supported yet")
		}

		// The colLabel must be a string.
		if colLabel.Type().Nature() != semantic.String {
			return errors.New(codes.FailedPrecondition, "column key must be of type string")
		}
		label := colLabel.Str()

		// Retrieve the buffer associated with this value.
		buf, ok := t.group.buffers[label]
		if !ok {
			buf = &pivotTableBuffer{valueType: valueType}
			t.group.buffers[label] = buf
		}

		if buf.valueType != valueType {
			return errors.New(codes.FailedPrecondition, "value columns with the same column key have different types")
		}

		// Insert the array associated with the row
		// key and the value column into the buffer.
		k, v := t.getColumn(cr, rowIndex), t.getColumn(cr, valueIndex)
		buf.Insert(k, v)

		return nil
	})

}

func (t *sortedPivotTransformation) validateTable(tbl flux.Table) error {
	if missingColumn, ok := func() (string, bool) {
		for _, v := range t.spec.RowKey {
			if idx := execute.ColIdx(v, tbl.Cols()); idx < 0 {
				return v, false
			}
		}

		for _, v := range t.spec.ColumnKey {
			if idx := execute.ColIdx(v, tbl.Cols()); idx < 0 {
				return v, false
			}
		}

		if idx := execute.ColIdx(t.spec.ValueColumn, tbl.Cols()); idx < 0 {
			return t.spec.ValueColumn, false
		}
		return "", true
	}(); !ok {
		return errors.Newf(codes.FailedPrecondition, "specified column does not exist in table: %v", missingColumn)
	}
	return nil
}

// computeGroupKey will compute the group key for a table with a given key.
// This is constructed by removing any columns within the column key
// or the value column.
func (t *sortedPivotTransformation) computeGroupKey(key flux.GroupKey) flux.GroupKey {
	// TODO(jsternberg): This can be optimized further when we
	// refactor the group key implementation so it is more composable.
	// https://github.com/influxdata/flux/issues/1032
	// There's no requirement for us to copy the group key here
	// as this is a simple filter and we also don't even know if
	// we're going to even filter anything when we compute this.
	// But as this simplifies the current implementation, we'll revisit
	// this later.
	cols := make([]flux.ColMeta, 0, len(key.Cols()))
	vs := make([]values.Value, 0, len(key.Cols()))
	for i, col := range key.Cols() {
		if col.Label == t.spec.ValueColumn || t.contains(col.Label, t.spec.ColumnKey) {
			continue
		}
		cols = append(cols, col)
		vs = append(vs, key.Value(i))
	}
	return execute.NewGroupKey(cols, vs)
}

func (t *sortedPivotTransformation) contains(v string, ss []string) bool {
	for _, s := range ss {
		if s == v {
			return true
		}
	}
	return false
}

func (t *sortedPivotTransformation) getColumn(cr flux.ColReader, j int) array.Array {
	switch col := cr.Cols()[j]; col.Type {
	case flux.TInt:
		return cr.Ints(j)
	case flux.TUInt:
		return cr.UInts(j)
	case flux.TFloat:
		return cr.Floats(j)
	case flux.TString:
		return cr.Strings(j)
	case flux.TBool:
		return cr.Bools(j)
	case flux.TTime:
		return cr.Times(j)
	default:
		panic(fmt.Sprintf("unexpected column type: %s", col.Type))
	}
}

func (t *sortedPivotTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	t.watermark = mark
	return nil
}

func (t *sortedPivotTransformation) UpdateProcessingTime(id execute.DatasetID, mark execute.Time) error {
	t.processing = mark
	return nil
}

func (t *sortedPivotTransformation) Finish(id execute.DatasetID, err error) {
	// Inform the downstream dataset that we are finished.
	// Wrap this in a function so that we do not capture the err variable.
	// https://play.golang.org/p/QXns3c8s76f
	defer func() { t.d.Finish(err) }()

	if err != nil {

		return
	}

	if t.group != nil {
		var tbl flux.Table
		gr := t.group
		tbl, err = gr.doPivot(t.prevOutputKey, t.alloc)
		if err != nil {
			return
		}
		if err = t.d.Process(tbl); err != nil {
			return
		}
	}

	if err = t.d.UpdateWatermark(t.watermark); err != nil {
		return
	}
	if err = t.d.UpdateProcessingTime(t.processing); err != nil {
		return
	}
}

type pivotTableBuffer struct {
	keys      []array.Array
	valueType flux.ColType
	values    []array.Array
}

func (b *pivotTableBuffer) Insert(k, v array.Array) {
	k.Retain()
	b.keys = append(b.keys, k)
	v.Retain()
	b.values = append(b.values, v)
}

func (b *pivotTableBuffer) Release() {
	for _, k := range b.keys {
		k.Release()
	}
	for _, v := range b.values {
		v.Release()
	}
}

type pivotTableGroup struct {
	rowCol  flux.ColMeta
	buffers map[string]*pivotTableBuffer
}

func (gr *pivotTableGroup) doPivot(key flux.GroupKey, mem arrowmemory.Allocator) (flux.Table, error) {
	// Merge all of the keys from each buffer.
	keys := gr.mergeKeys(mem)

	// Create the table buffer that will be used for the final table.
	ncols := len(key.Cols()) + len(gr.buffers)
	tb := &arrow.TableBuffer{
		GroupKey: key,
		Columns:  make([]flux.ColMeta, 0, ncols),
		Values:   make([]array.Array, 0, ncols),
	}
	tb.Columns = append(tb.Columns, gr.rowCol)
	tb.Values = append(tb.Values, keys)

	// Add the group key columns to the table.
	for j, col := range key.Cols() {
		tb.Columns = append(tb.Columns, col)
		tb.Values = append(tb.Values, arrow.Repeat(key.Cols()[j].Type, key.Value(j), keys.Len(), mem))
	}

	// Build each column by appending a value when it matches
	// with one of the keys and null when it does not.
	labels := make([]string, 0, len(gr.buffers))
	for label := range gr.buffers {
		labels = append(labels, label)
	}
	sort.Strings(labels)

	for _, label := range labels {
		buf := gr.buffers[label]
		vs := gr.buildColumn(keys, buf, mem)
		tb.Columns = append(tb.Columns, flux.ColMeta{
			Label: label,
			Type:  buf.valueType,
		})
		tb.Values = append(tb.Values, vs)
		buf.Release()
	}

	if err := tb.Validate(); err != nil {
		return nil, err
	}
	return table.FromBuffer(tb), nil
}
