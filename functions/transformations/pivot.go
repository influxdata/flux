package transformations

import (
	"fmt"
	"strconv"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

const PivotKind = "pivot"

type PivotOpSpec struct {
	RowKey   []string `json:"rowKey"`
	ColKey   []string `json:"colKey"`
	ValueCol string   `json:"valueCol"`
}

var pivotSignature = flux.DefaultFunctionSignature()

var fromRowsBuiltin = `
// fromRows will access a database and retrieve data aligned into time-aligned tuples, grouped by measurement.  
fromRows = (bucket="",bucketID="") => from(bucket:bucket,bucketID:bucketID) |> pivot(rowKey:["_time"], colKey: ["_field"], valueCol: "_value")
`

func init() {
	pivotSignature.Params["rowKey"] = semantic.Array
	pivotSignature.Params["colKey"] = semantic.Array
	pivotSignature.Params["valueCol"] = semantic.String

	flux.RegisterFunction(PivotKind, createPivotOpSpec, pivotSignature)
	flux.RegisterBuiltIn("fromRows", fromRowsBuiltin)
	flux.RegisterOpSpec(PivotKind, newPivotOp)

	plan.RegisterProcedureSpec(PivotKind, newPivotProcedure, PivotKind)
	execute.RegisterTransformation(PivotKind, createPivotTransformation)
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

	array, err = args.GetRequiredArray("colKey", semantic.String)
	if err != nil {
		return nil, err
	}

	spec.ColKey, err = interpreter.ToStringArray(array)
	if err != nil {
		return nil, err
	}

	rowKeys := make(map[string]bool)
	for _, v := range spec.RowKey {
		rowKeys[v] = true
	}

	for _, v := range spec.ColKey {
		if _, ok := rowKeys[v]; ok {
			return nil, fmt.Errorf("column name found in both rowKey and colKey: %s", v)
		}
	}

	valueCol, err := args.GetRequiredString("valueCol")
	if err != nil {
		return nil, err
	}
	spec.ValueCol = valueCol

	return spec, nil
}

func newPivotOp() flux.OperationSpec {
	return new(PivotOpSpec)
}

func (s *PivotOpSpec) Kind() flux.OperationKind {
	return PivotKind
}

type PivotProcedureSpec struct {
	RowKey   []string
	ColKey   []string
	ValueCol string
}

func newPivotProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*PivotOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	p := &PivotProcedureSpec{
		RowKey:   spec.RowKey,
		ColKey:   spec.ColKey,
		ValueCol: spec.ValueCol,
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
	ns.ColKey = make([]string, len(s.ColKey))
	copy(ns.ColKey, s.ColKey)
	ns.ValueCol = s.ValueCol
	return ns
}

func createPivotTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*PivotProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", spec)
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
			return fmt.Errorf("specified column does not exist in table: %v", v)
		}
		rowKeyIndex[v] = idx
	}

	// different from above because we'll get the column indices below when we
	// determine the initial column schema
	colKeyIndex := make(map[string]int)
	valueColIndex := -1
	var valueColType flux.DataType
	for _, v := range t.spec.ColKey {
		colKeyIndex[v] = -1
	}

	cols := make([]flux.ColMeta, 0, len(tbl.Cols()))
	keyCols := make([]flux.ColMeta, 0, len(tbl.Key().Cols()))
	keyValues := make([]values.Value, 0, len(tbl.Key().Cols()))
	newIDX := 0
	colMap := make([]int, len(tbl.Cols()))

	for colIDX, v := range tbl.Cols() {
		if _, ok := colKeyIndex[v.Label]; !ok && v.Label != t.spec.ValueCol {
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
		} else if v.Label == t.spec.ValueCol {
			valueColIndex = colIDX
			valueColType = tbl.Cols()[colIDX].Type
		} else {
			// we need the location of the colKey columns in the original table
			colKeyIndex[v.Label] = colIDX
		}
	}

	for k, v := range colKeyIndex {
		if v < 0 {
			return fmt.Errorf("specified column does not exist in table: %v", k)
		}
	}

	newGroupKey := execute.NewGroupKey(keyCols, keyValues)
	builder, created := t.cache.TableBuilder(newGroupKey)
	groupKeyString := newGroupKey.String()
	if created {
		for _, c := range cols {
			builder.AddCol(c)
		}
		t.colKeyMaps[groupKeyString] = make(map[string]int)
		t.rowKeyMaps[groupKeyString] = make(map[string]int)
		t.nextRowCol[groupKeyString] = rowCol{nextCol: len(cols), nextRow: 0}
	}

	err := tbl.Do(func(cr flux.ColReader) error {
		for row := 0; row < cr.Len(); row++ {
			rowKey := ""
			colKey := ""
			for j, c := range cr.Cols() {
				if _, ok := rowKeyIndex[c.Label]; ok {
					rowKey += valueToStr(cr, c, row, j)
				} else if _, ok := colKeyIndex[c.Label]; ok {
					if colKey == "" {
						colKey = valueToStr(cr, c, row, j)
					} else {
						colKey = colKey + "_" + valueToStr(cr, c, row, j)
					}
				}
			}

			// we have columns for the copy-over in place;
			// we know the row key;
			// we know the col key;
			//  0.  If we've not seen the colKey before, then we need to add a new column and backfill it.
			if _, ok := t.colKeyMaps[groupKeyString][colKey]; !ok {
				newCol := flux.ColMeta{
					Label: colKey,
					Type:  valueColType,
				}
				builder.AddCol(newCol)
				nextRowCol := t.nextRowCol[groupKeyString]
				growColumn(builder, newCol.Type, nextRowCol.nextCol, builder.NRows())
				t.colKeyMaps[groupKeyString][colKey] = nextRowCol.nextCol
				nextRowCol.nextCol++
				t.nextRowCol[groupKeyString] = nextRowCol
			}
			//  1.  if we've not seen rowKey before, then we need to append a new row, with copied values for the
			//  existing columns, as well as zero values for the pivoted columns.
			if _, ok := t.rowKeyMaps[groupKeyString][rowKey]; !ok {
				// rowkey U groupKey cols
				for cidx, c := range cols {
					appendBuilderValue(cr, builder, c.Type, row, colMap[cidx], cidx)
				}

				// zero-out the known key columns we've already discovered.
				for _, v := range t.colKeyMaps[groupKeyString] {
					growColumn(builder, valueColType, v, 1)
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
			setBuilderValue(cr, builder, valueColType, row, valueColIndex, t.rowKeyMaps[groupKeyString][rowKey],
				t.colKeyMaps[groupKeyString][colKey])

		}
		return nil
	})

	return err
}

func growColumn(builder execute.TableBuilder, colType flux.DataType, colIdx, nRows int) {
	switch colType {
	case flux.TBool:
		builder.GrowBools(colIdx, nRows)
	case flux.TInt:
		builder.GrowInts(colIdx, nRows)
	case flux.TUInt:
		builder.GrowUInts(colIdx, nRows)
	case flux.TFloat:
		builder.GrowFloats(colIdx, nRows)
	case flux.TString:
		builder.GrowStrings(colIdx, nRows)
	case flux.TTime:
		builder.GrowTimes(colIdx, nRows)
	default:
		execute.PanicUnknownType(colType)
	}
}

func setBuilderValue(cr flux.ColReader, builder execute.TableBuilder, readerColType flux.DataType, readerRowIndex, readerColIndex, builderRow, builderCol int) {
	switch readerColType {
	case flux.TBool:
		builder.SetBool(builderRow, builderCol, cr.Bools(readerColIndex)[readerRowIndex])
	case flux.TInt:
		builder.SetInt(builderRow, builderCol, cr.Ints(readerColIndex)[readerRowIndex])
	case flux.TUInt:
		builder.SetUInt(builderRow, builderCol, cr.UInts(readerColIndex)[readerRowIndex])
	case flux.TFloat:
		builder.SetFloat(builderRow, builderCol, cr.Floats(readerColIndex)[readerRowIndex])
	case flux.TString:
		builder.SetString(builderRow, builderCol, cr.Strings(readerColIndex)[readerRowIndex])
	case flux.TTime:
		builder.SetTime(builderRow, builderCol, cr.Times(readerColIndex)[readerRowIndex])
	default:
		execute.PanicUnknownType(readerColType)
	}
}

func appendBuilderValue(cr flux.ColReader, builder execute.TableBuilder, readerColType flux.DataType, readerRowIndex, readerColIndex, builderColIndex int) {
	switch readerColType {
	case flux.TBool:
		builder.AppendBool(builderColIndex, cr.Bools(readerColIndex)[readerRowIndex])
	case flux.TInt:
		builder.AppendInt(builderColIndex, cr.Ints(readerColIndex)[readerRowIndex])
	case flux.TUInt:
		builder.AppendUInt(builderColIndex, cr.UInts(readerColIndex)[readerRowIndex])
	case flux.TFloat:
		builder.AppendFloat(builderColIndex, cr.Floats(readerColIndex)[readerRowIndex])
	case flux.TString:
		builder.AppendString(builderColIndex, cr.Strings(readerColIndex)[readerRowIndex])
	case flux.TTime:
		builder.AppendTime(builderColIndex, cr.Times(readerColIndex)[readerRowIndex])
	default:
		execute.PanicUnknownType(readerColType)
	}
}

func valueToStr(cr flux.ColReader, c flux.ColMeta, row, col int) string {
	switch c.Type {
	case flux.TBool:
		return strconv.FormatBool(cr.Bools(col)[row])
	case flux.TInt:
		return strconv.FormatInt(cr.Ints(col)[row], 10)
	case flux.TUInt:
		return strconv.FormatUint(cr.UInts(col)[row], 10)
	case flux.TFloat:
		return strconv.FormatFloat(cr.Floats(col)[row], 'E', -1, 64)
	case flux.TString:
		return cr.Strings(col)[row]
	case flux.TTime:
		return cr.Times(col)[row].String()
	default:
		execute.PanicUnknownType(c.Type)
	}
	return ""
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
