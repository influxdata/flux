package universe

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

const (
	PivotKind      = "pivot"
	nullValueLabel = "null"
)

type PivotOpSpec struct {
	RowKey      []string `json:"rowKey"`
	ColumnKey   []string `json:"columnKey"`
	ValueColumn string   `json:"valueColumn"`
}

func init() {
	pivotSignature := flux.FunctionSignature(
		map[string]semantic.PolyType{
			"rowKey":      semantic.NewArrayPolyType(semantic.String),
			"columnKey":   semantic.NewArrayPolyType(semantic.String),
			"valueColumn": semantic.String,
		},
		[]string{"rowKey", "columnKey", "valueColumn"},
	)

	flux.RegisterPackageValue("universe", PivotKind, flux.FunctionValue(PivotKind, createPivotOpSpec, pivotSignature))
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
			return nil, fmt.Errorf("column name found in both rowKey and colKey: %s", v)
		}
	}

	valueCol, err := args.GetRequiredString("valueColumn")
	if err != nil {
		return nil, err
	}
	spec.ValueColumn = valueCol

	return spec, nil
}

func newPivotOp() flux.OperationSpec {
	return new(PivotOpSpec)
}

func (s *PivotOpSpec) Kind() flux.OperationKind {
	return PivotKind
}

type PivotProcedureSpec struct {
	plan.DefaultCost
	RowKey      []string
	ColumnKey   []string
	ValueColumn string
}

func newPivotProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*PivotOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
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
			rowKey := ""
			colKey := ""
			for _, rk := range t.spec.RowKey {
				j := rowKeyIndex[rk]
				c := cr.Cols()[j]
				rowKey += valueToStr(cr, c, row, j)
			}

			for _, ck := range t.spec.ColumnKey {
				j := colKeyIndex[ck]
				c := cr.Cols()[j]
				if colKey == "" {
					colKey = valueToStr(cr, c, row, j)
				} else {
					colKey = colKey + "_" + valueToStr(cr, c, row, j)
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
				nextCol, err := builder.AddCol(newCol)
				if err != nil {
					return err
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
		return fmt.Errorf("invalid column type: %s", colType)
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
			result = v.ValueString(row)
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
