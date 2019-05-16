package executetest

import (
	"fmt"
	"testing"

	"github.com/apache/arrow/go/arrow/array"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

// Table is an implementation of execute.Table
// It is designed to make it easy to statically declare the data within the table.
// Not all fields need to be set. See comments on each field.
// Use Normalize to ensure that all fields are set before equality comparisons.
type Table struct {
	// GroupKey of the table. Does not need to be set explicitly.
	GroupKey flux.GroupKey
	// KeyCols is a list of column that are part of the group key.
	// The column type is deduced from the ColMeta slice.
	KeyCols []string
	// KeyValues is a list of values for the group key columns.
	// Only needs to be set when no data is present on the table.
	KeyValues []interface{}
	// ColMeta is a list of columns of the table.
	ColMeta []flux.ColMeta
	// Data is a list of rows, i.e. Data[row][col]
	// Each row must be a list with length equal to len(ColMeta)
	Data [][]interface{}
	// Err contains the error that should be returned
	// by this table when calling Do.
	Err error
	// IsDone indicates if this table has been used.
	IsDone bool
}

// Normalize ensures all fields of the table are set correctly.
func (t *Table) Normalize() {
	if t.GroupKey == nil {
		cols := make([]flux.ColMeta, len(t.KeyCols))
		vs := make([]values.Value, len(t.KeyCols))
		if len(t.KeyValues) != len(t.KeyCols) {
			t.KeyValues = make([]interface{}, len(t.KeyCols))
		}
		for j, label := range t.KeyCols {
			idx := execute.ColIdx(label, t.ColMeta)
			if idx < 0 {
				panic(fmt.Errorf("table invalid: missing group column %q", label))
			}
			cols[j] = t.ColMeta[idx]
			if len(t.Data) > 0 {
				t.KeyValues[j] = t.Data[0][idx]
			}
			var v values.Value
			if t.KeyValues[j] == nil {
				v = values.NewNull(flux.SemanticType(t.ColMeta[idx].Type))
			} else {
				v = values.New(t.KeyValues[j])
				if v.Type() == semantic.Invalid {
					panic(fmt.Errorf("invalid value: %s", t.KeyValues[j]))
				}
			}
			vs[j] = v
		}
		t.GroupKey = execute.NewGroupKey(cols, vs)
	}
}

func (t *Table) Empty() bool {
	return len(t.Data) == 0
}

func (t *Table) RefCount(n int) {}
func (t *Table) Done() {
	t.IsDone = true
}

func (t *Table) Cols() []flux.ColMeta {
	return t.ColMeta
}

func (t *Table) Key() flux.GroupKey {
	t.Normalize()
	return t.GroupKey
}

func (t *Table) Do(f func(flux.ColReader) error) error {
	if t.Err != nil {
		return t.Err
	}

	cols := make([]array.Interface, len(t.ColMeta))
	for j, col := range t.ColMeta {
		switch col.Type {
		case flux.TBool:
			b := arrow.NewBoolBuilder(nil)
			for i := range t.Data {
				if v := t.Data[i][j]; v != nil {
					b.Append(v.(bool))
				} else {
					b.AppendNull()
				}
			}
			cols[j] = b.NewBooleanArray()
			b.Release()
		case flux.TFloat:
			b := arrow.NewFloatBuilder(nil)
			for i := range t.Data {
				if v := t.Data[i][j]; v != nil {
					b.Append(v.(float64))
				} else {
					b.AppendNull()
				}
			}
			cols[j] = b.NewFloat64Array()
			b.Release()
		case flux.TInt:
			b := arrow.NewIntBuilder(nil)
			for i := range t.Data {
				if v := t.Data[i][j]; v != nil {
					b.Append(v.(int64))
				} else {
					b.AppendNull()
				}
			}
			cols[j] = b.NewInt64Array()
			b.Release()
		case flux.TString:
			b := arrow.NewStringBuilder(nil)
			for i := range t.Data {
				if v := t.Data[i][j]; v != nil {
					b.AppendString(v.(string))
				} else {
					b.AppendNull()
				}
			}
			cols[j] = b.NewBinaryArray()
			b.Release()
		case flux.TTime:
			b := arrow.NewIntBuilder(nil)
			for i := range t.Data {
				if v := t.Data[i][j]; v != nil {
					b.Append(int64(v.(values.Time)))
				} else {
					b.AppendNull()
				}
			}
			cols[j] = b.NewInt64Array()
			b.Release()
		case flux.TUInt:
			b := arrow.NewUintBuilder(nil)
			for i := range t.Data {
				if v := t.Data[i][j]; v != nil {
					b.Append(v.(uint64))
				} else {
					b.AppendNull()
				}
			}
			cols[j] = b.NewUint64Array()
			b.Release()
		}
	}

	cr := &ColReader{
		key:  t.Key(),
		meta: t.ColMeta,
		cols: cols,
	}
	return f(cr)
}

type ColReader struct {
	key  flux.GroupKey
	meta []flux.ColMeta
	cols []array.Interface
}

func (cr *ColReader) Key() flux.GroupKey {
	return cr.key
}

func (cr *ColReader) Cols() []flux.ColMeta {
	return cr.meta
}

func (cr *ColReader) Len() int {
	if len(cr.cols) == 0 {
		return 0
	}
	return cr.cols[0].Len()
}

func (cr *ColReader) Bools(j int) *array.Boolean {
	return cr.cols[j].(*array.Boolean)
}

func (cr *ColReader) Ints(j int) *array.Int64 {
	return cr.cols[j].(*array.Int64)
}

func (cr *ColReader) UInts(j int) *array.Uint64 {
	return cr.cols[j].(*array.Uint64)
}

func (cr *ColReader) Floats(j int) *array.Float64 {
	return cr.cols[j].(*array.Float64)
}

func (cr *ColReader) Strings(j int) *array.Binary {
	return cr.cols[j].(*array.Binary)
}

func (cr *ColReader) Times(j int) *array.Int64 {
	return cr.cols[j].(*array.Int64)
}

// RowWiseTable is a flux Table implementation that
// calls f once for each row in its Do method.
type RowWiseTable struct {
	*Table
}

// Do calls f once for each row in the table
func (t *RowWiseTable) Do(f func(flux.ColReader) error) error {
	cols := make([]array.Interface, len(t.ColMeta))
	for j, col := range t.ColMeta {
		switch col.Type {
		case flux.TBool:
			b := arrow.NewBoolBuilder(nil)
			for i := range t.Data {
				if v := t.Data[i][j]; v != nil {
					b.Append(v.(bool))
				} else {
					b.AppendNull()
				}
			}
			cols[j] = b.NewBooleanArray()
			b.Release()
		case flux.TFloat:
			b := arrow.NewFloatBuilder(nil)
			for i := range t.Data {
				if v := t.Data[i][j]; v != nil {
					b.Append(v.(float64))
				} else {
					b.AppendNull()
				}
			}
			cols[j] = b.NewFloat64Array()
			b.Release()
		case flux.TInt:
			b := arrow.NewIntBuilder(nil)
			for i := range t.Data {
				if v := t.Data[i][j]; v != nil {
					b.Append(v.(int64))
				} else {
					b.AppendNull()
				}
			}
			cols[j] = b.NewInt64Array()
			b.Release()
		case flux.TString:
			b := arrow.NewStringBuilder(nil)
			for i := range t.Data {
				if v := t.Data[i][j]; v != nil {
					b.AppendString(v.(string))
				} else {
					b.AppendNull()
				}
			}
			cols[j] = b.NewBinaryArray()
			b.Release()
		case flux.TTime:
			b := arrow.NewIntBuilder(nil)
			for i := range t.Data {
				if v := t.Data[i][j]; v != nil {
					b.Append(int64(v.(values.Time)))
				} else {
					b.AppendNull()
				}
			}
			cols[j] = b.NewInt64Array()
			b.Release()
		case flux.TUInt:
			b := arrow.NewUintBuilder(nil)
			for i := range t.Data {
				if v := t.Data[i][j]; v != nil {
					b.Append(v.(uint64))
				} else {
					b.AppendNull()
				}
			}
			cols[j] = b.NewUint64Array()
			b.Release()
		}
	}

	release := func(cols []array.Interface) {
		for _, arr := range cols {
			arr.Release()
		}
	}
	defer release(cols)

	l := cols[0].Len()
	for i := 0; i < l; i++ {
		row := make([]array.Interface, len(t.ColMeta))
		for j, col := range t.ColMeta {
			switch col.Type {
			case flux.TBool:
				row[j] = arrow.BoolSlice(cols[j].(*array.Boolean), i, i+1)
			case flux.TFloat:
				row[j] = arrow.FloatSlice(cols[j].(*array.Float64), i, i+1)
			case flux.TInt:
				row[j] = arrow.IntSlice(cols[j].(*array.Int64), i, i+1)
			case flux.TString:
				row[j] = arrow.StringSlice(cols[j].(*array.Binary), i, i+1)
			case flux.TTime:
				row[j] = arrow.IntSlice(cols[j].(*array.Int64), i, i+1)
			case flux.TUInt:
				row[j] = arrow.UintSlice(cols[j].(*array.Uint64), i, i+1)
			}
		}
		if err := f(&ColReader{
			key:  t.Key(),
			meta: t.ColMeta,
			cols: row,
		}); err != nil {
			return err
		}
		release(row)
	}
	return nil
}

func TablesFromCache(c execute.DataCache) (tables []*Table, err error) {
	c.ForEach(func(key flux.GroupKey) {
		if err != nil {
			return
		}
		var tbl flux.Table
		tbl, err = c.Table(key)
		if err != nil {
			return
		}
		var cb *Table
		cb, err = ConvertTable(tbl)
		if err != nil {
			return
		}
		tables = append(tables, cb)
		c.ExpireTable(key)
	})
	return tables, nil
}

func ConvertTable(tbl flux.Table) (*Table, error) {
	key := tbl.Key()
	blk := &Table{
		GroupKey: key,
		ColMeta:  tbl.Cols(),
	}

	keyCols := key.Cols()
	if len(keyCols) > 0 {
		blk.KeyCols = make([]string, len(keyCols))
		blk.KeyValues = make([]interface{}, len(keyCols))
		for j, c := range keyCols {
			blk.KeyCols[j] = c.Label
			var v interface{}
			if !key.IsNull(j) {
				switch c.Type {
				case flux.TBool:
					v = key.ValueBool(j)
				case flux.TUInt:
					v = key.ValueUInt(j)
				case flux.TInt:
					v = key.ValueInt(j)
				case flux.TFloat:
					v = key.ValueFloat(j)
				case flux.TString:
					v = key.ValueString(j)
				case flux.TTime:
					v = key.ValueTime(j)
				default:
					return nil, fmt.Errorf("unsupported column type %v", c.Type)
				}
			}
			blk.KeyValues[j] = v
		}
	}

	err := tbl.Do(func(cr flux.ColReader) error {
		l := cr.Len()
		for i := 0; i < l; i++ {
			row := make([]interface{}, len(blk.ColMeta))
			for j, c := range blk.ColMeta {
				switch c.Type {
				case flux.TBool:
					if col := cr.Bools(j); col.IsValid(i) {
						row[j] = col.Value(i)
					}
				case flux.TInt:
					if col := cr.Ints(j); col.IsValid(i) {
						row[j] = col.Value(i)
					}
				case flux.TUInt:
					if col := cr.UInts(j); col.IsValid(i) {
						row[j] = col.Value(i)
					}
				case flux.TFloat:
					if col := cr.Floats(j); col.IsValid(i) {
						row[j] = col.Value(i)
					}
				case flux.TString:
					if col := cr.Strings(j); col.IsValid(i) {
						row[j] = col.ValueString(i)
					}
				case flux.TTime:
					if col := cr.Times(j); col.IsValid(i) {
						row[j] = values.Time(col.Value(i))
					}
				default:
					panic(fmt.Errorf("unknown column type %s", c.Type))
				}
			}
			blk.Data = append(blk.Data, row)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return blk, nil
}

type SortedTables []*Table

func (b SortedTables) Len() int {
	return len(b)
}

func (b SortedTables) Less(i int, j int) bool {
	return b[i].Key().Less(b[j].Key())
}

func (b SortedTables) Swap(i int, j int) {
	b[i], b[j] = b[j], b[i]
}

// NormalizeTables ensures that each table is normalized
func NormalizeTables(bs []*Table) {
	for _, b := range bs {
		b.Key()
	}
}

func MustCopyTable(tbl flux.Table) flux.Table {
	cpy, _ := execute.CopyTable(tbl, UnlimitedAllocator)
	return cpy
}

type TableTest struct {
	CreateTableFn      func() flux.Table
	CreateEmptyTableFn func() flux.Table
	IsDone             func(flux.Table) bool
}

func (tt *TableTest) CreateTable(t *testing.T) flux.Table {
	t.Helper()

	tbl := tt.CreateTableFn()
	if tt.IsDone(tbl) {
		t.Fatal("table is done before the test has started")
	} else if tbl.Empty() {
		t.Fatal("table is empty")
	}
	return tbl
}

func (tt *TableTest) CreateEmptyTable(t *testing.T) flux.Table {
	t.Helper()

	tbl := tt.CreateEmptyTableFn()
	if !tbl.Empty() {
		t.Fatal("table is not empty")
	}
	return tbl
}

// RunTableTests will run the common table tests over the table
// implementation. The function will be called for each test.
func RunTableTests(t *testing.T, tt TableTest) {
	t.Run("Normal", func(t *testing.T) {
		tbl := tt.CreateTable(t)
		if err := tbl.Do(func(flux.ColReader) error {
			return nil
		}); err != nil {
			t.Errorf("unexpected error when reading table: %s", err)
		}

		if !tt.IsDone(tbl) {
			t.Error("table is not done after calling Do")
		}
	})
	t.Run("MultipleDoCalls", func(t *testing.T) {
		tbl := tt.CreateTable(t)
		if err := tbl.Do(func(flux.ColReader) error {
			return nil
		}); err != nil {
			t.Errorf("unexpected error when reading table: %s", err)
		}

		if err := tbl.Do(func(flux.ColReader) error {
			return nil
		}); err == nil {
			t.Error("expected error when calling Do twice")
		}
	})
	t.Run("MultipleDoneCalls", func(t *testing.T) {
		tbl := tt.CreateTable(t)
		tbl.Done()
		if !tt.IsDone(tbl) {
			t.Error("table is not done after calling Done")
		}
		tbl.Done()
	})
	t.Run("DoneOnly", func(t *testing.T) {
		tbl := tt.CreateTable(t)
		tbl.Done()
		if !tt.IsDone(tbl) {
			t.Error("table is not done after calling Done")
		}
	})
	t.Run("DoneWhileEmpty", func(t *testing.T) {
		tbl := tt.CreateEmptyTable(t)
		// Table should already be done.
		if !tt.IsDone(tbl) {
			t.Error("empty table should be immediately done")
		}
		// Just ensure this doesn't panic or anything.
		tbl.Done()
	})
}
