package executetest

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/apache/arrow/go/arrow/array"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
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
	// Alloc is the allocator used to create the column readers.
	// Memory is not tracked unless this is set.
	Alloc *memory.Allocator
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
				if v.Type().Nature() == semantic.Invalid {
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
	} else if t.IsDone {
		return errors.New(codes.Internal, "table already read")
	}
	t.IsDone = true

	cols := make([]array.Interface, len(t.ColMeta))
	for j, col := range t.ColMeta {
		switch col.Type {
		case flux.TBool:
			b := arrow.NewBoolBuilder(t.Alloc)
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
			b := arrow.NewFloatBuilder(t.Alloc)
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
			b := arrow.NewIntBuilder(t.Alloc)
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
			b := arrow.NewStringBuilder(t.Alloc)
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
			b := arrow.NewIntBuilder(t.Alloc)
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
			b := arrow.NewUintBuilder(t.Alloc)
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
	defer cr.Release()
	return f(cr)
}

func (t *Table) Done() {
	t.IsDone = true
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

func (cr *ColReader) Retain() {
	for _, col := range cr.cols {
		col.Retain()
	}
}

func (cr *ColReader) Release() {
	for _, col := range cr.cols {
		col.Release()
	}
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
	return tables, err
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

// NormalizeTables ensures that each table is normalized and that tables and columns are sorted in
// alphabetical order for consistent testing
func NormalizeTables(bs []*Table) {
	for _, b := range bs {
		b.Key()
	}
	sortByGroupKey(bs)
}

func sortByGroupKey(tables []*Table) {
	sort.Sort(SortedTables(tables))
	for _, table := range tables {
		sortColumns(table)
	}
}

func sortColumns(table *Table) {
	// index reference to make sure that sort is consistent for metadata and data
	indexes := createReferenceArray(table)

	// the copy function doesn't work well with flux.ColMeta. Doing a manual copy here instead
	columnsCopy := make([]flux.ColMeta, len(table.ColMeta))
	for i, value := range indexes {
		columnsCopy[i] = table.ColMeta[value]
	}
	table.ColMeta = columnsCopy

	// don't create a new table.Data if there is no Data
	if len(table.Data) == 0 {
		return
	}
	valuesCopy := make([][]interface{}, len(table.Data))
	for j, row := range table.Data {
		if len(row) > 0 {
			rowCopy := make([]interface{}, len(row))
			for i, value := range indexes {
				rowCopy[i] = row[value]
			}
			valuesCopy[j] = rowCopy
		}
	}
	table.Data = valuesCopy
}

func createReferenceArray(table *Table) []int {
	indexes := make([]int, len(table.ColMeta))
	for i := range indexes {
		indexes[i] = i
	}
	sort.Slice(indexes, func(i, j int) bool {
		return table.ColMeta[indexes[i]].Label < table.ColMeta[indexes[j]].Label
	})
	return indexes
}

func MustCopyTable(tbl flux.Table) flux.Table {
	cpy, err := execute.CopyTable(tbl)
	if err != nil {
		panic(err)
	}
	return cpy
}

type TableTest struct {
	// NewFn returns a new TableIterator that can be processed.
	// The table iterator that is produced should have multiple
	// tables of different shapes and sizes to get coverage of
	// as much of the code as possible. The TableIterator will
	// be created once for each subtest.
	NewFn func(ctx context.Context, alloc *memory.Allocator) flux.TableIterator

	// IsDone will report if the table is considered done for reading.
	// The call to Done should force this to be true, but it is possible
	// for this to return true before the table has been processed.
	IsDone func(flux.Table) bool
}

func (tt TableTest) run(t *testing.T, name string, f func(tt *tableTest)) {
	t.Run(name, func(t *testing.T) {
		defer func() {
			if err := recover(); err != nil {
				t.Errorf("panic occurred while running the test: %v", err)
			}
		}()
		f(&tableTest{
			TableTest: tt,
			t:         t,
			logger:    zaptest.NewLogger(t),
			alloc:     &memory.Allocator{},
		})
	})
}

type tableTest struct {
	TableTest
	t      *testing.T
	logger *zap.Logger
	alloc  *memory.Allocator
}

func (tt *tableTest) do(f func(tbl flux.Table) error) {
	tt.t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tables := tt.NewFn(ctx, tt.alloc)
	if err := tables.Do(func(tbl flux.Table) error {
		tt.logger.Debug("processing table", zap.Stringer("key", tbl.Key()))
		if err := f(tbl); err != nil {
			return err
		}
		if !tt.IsDone(tbl) {
			tt.t.Error("table is not done after the test has finished")
		}
		return nil
	}); err != nil {
		tt.t.Errorf("unexpected error when processing tables: %s", err)
	}
}

func (tt *tableTest) finish(allocatorUsed bool) {
	tt.t.Helper()

	if allocatorUsed {
		// This ensures the allocator was actually used at some point
		// to ensure that the below check is actually valid.
		if got := tt.alloc.MaxAllocated(); got == 0 {
			tt.t.Error("memory allocator was not used")
		}
	}

	// Verify that all memory is correctly released if we use the table properly.
	if got := tt.alloc.Allocated(); got != 0 {
		tt.t.Errorf("caught memory leak: %d bytes were not released", got)
	}
}

// RunTableTests will run the common table tests over each table
// in the returned TableIterator. The function will be called for
// each test.
func RunTableTests(t *testing.T, tt TableTest) {
	tt.run(t, "Normal", func(tt *tableTest) {
		// Ensure that calling Do works correctly.
		// The Done call should not be required.
		tt.do(func(tbl flux.Table) error {
			return tbl.Do(func(flux.ColReader) error {
				return nil
			})
		})
		tt.finish(true)
	})
	tt.run(t, "MultipleDoCalls", func(tt *tableTest) {
		// When Do is called multiple times, the second use should
		// fail with an error.
		tt.do(func(tbl flux.Table) error {
			if err := tbl.Do(func(flux.ColReader) error {
				return nil
			}); err != nil {
				return err
			}

			// The second call should return an error.
			if err := tbl.Do(func(flux.ColReader) error {
				tt.t.Error("unexpected column reader on the second call to Do")
				return nil
			}); err == nil {
				tt.t.Error("expected error when calling Do twice")
			}
			return nil
		})
		tt.finish(true)
	})
	tt.run(t, "MultipleDoneCalls", func(tt *tableTest) {
		// Calling Done multiple times should be safe and not panic.
		tt.do(func(tbl flux.Table) error {
			tbl.Done()
			if !tt.IsDone(tbl) {
				t.Error("table is not done after calling Done")
			}
			tbl.Done()
			return nil
		})
		tt.finish(false)
	})
	tt.run(t, "DoneOnly", func(tt *tableTest) {
		// If the only thing called is Done, the table should work properly.
		tt.do(func(tbl flux.Table) error {
			tbl.Done()
			if !tt.IsDone(tbl) {
				t.Error("table is not done after calling Done")
			}
			return nil
		})
		tt.finish(false)
	})
	tt.run(t, "Empty", func(tt *tableTest) {
		// If Do returns no rows, then Empty should have returned true.
		tt.do(func(tbl flux.Table) error {
			got, want := tbl.Empty(), true
			if err := tbl.Do(func(cr flux.ColReader) error {
				if cr.Len() > 0 {
					want = false
				}
				return nil
			}); err != nil {
				return err
			}

			if got != want {
				tt.t.Errorf("unexpected value for empty -got/+want\n\t- %v\n\t+ %v", got, want)
			}
			return nil
		})
		tt.finish(false)
	})
	tt.run(t, "EmptyAfterDo", func(tt *tableTest) {
		// It should be ok to call empty after do and get the same result.
		tt.do(func(tbl flux.Table) error {
			want := tbl.Empty()
			if err := tbl.Do(func(cr flux.ColReader) error {
				if cr.Len() > 0 {
					want = false
				}
				return nil
			}); err != nil {
				return err
			}
			got := tbl.Empty()

			if got != want {
				tt.t.Errorf("unexpected value for empty -got/+want\n\t- %v\n\t+ %v", got, want)
			}
			return nil
		})
		tt.finish(true)
	})
	tt.run(t, "EmptyAfterDone", func(tt *tableTest) {
		// It should be ok to call empty after done and get the same result.
		tt.do(func(tbl flux.Table) error {
			want := tbl.Empty()
			tbl.Done()
			got := tbl.Empty()

			if got != want {
				tt.t.Errorf("unexpected value for empty -got/+want\n\t- %v\n\t+ %v", got, want)
			}
			return nil
		})
		tt.finish(false)
	})
	tt.run(t, "Retain", func(tt *tableTest) {
		// Retain should allow the column reader to be used outside of Do.
		tt.do(func(tbl flux.Table) error {
			if tbl.Empty() {
				return nil
			}

			var crs []flux.ColReader
			if err := tbl.Do(func(cr flux.ColReader) error {
				cr.Retain()
				crs = append(crs, cr)
				return nil
			}); err != nil {
				return err
			}

			// We are outside of the Do call. The column reader should
			// still be valid so we should be able to read a value from it.
			checkColReaders := func() {
				for _, cr := range crs {
					v := execute.ValueForRow(cr, 0, 0)
					tt.t.Logf("first value for first column is %v", v)
				}

				// If there were multiple column readers returned for a single table,
				// we need to check that they did not share a buffer.
				if len(crs) > 1 {
					for i := 0; i < len(crs)-1; i++ {
						for j := i + 1; j < len(crs); j++ {
							if colReadersEqual(crs[i], crs[j]) {
								tt.t.Errorf("retained column reader is the same as another column reader (%d, %d) for table %v", i, j, tbl.Key())
							}
						}
					}
				}
			}
			checkColReaders()

			// Call Done (which should have already happened by the call to Do)
			// and ensure that the above check still succeeds.
			tbl.Done()
			checkColReaders()

			for _, cr := range crs {
				cr.Release()
			}
			return nil
		})
		tt.finish(true)
	})
	tt.run(t, "Len", func(tt *tableTest) {
		tt.do(func(tbl flux.Table) error {
			return tbl.Do(func(cr flux.ColReader) error {
				want := cr.Len()
				for i, n := 0, len(cr.Cols()); i < n; i++ {
					got := func(cr flux.ColReader, i int) int {
						switch cr.Cols()[i].Type {
						case flux.TFloat:
							return cr.Floats(i).Len()
						case flux.TInt:
							return cr.Ints(i).Len()
						case flux.TUInt:
							return cr.UInts(i).Len()
						case flux.TString:
							return cr.Strings(i).Len()
						case flux.TBool:
							return cr.Bools(i).Len()
						case flux.TTime:
							return cr.Times(i).Len()
						default:
							panic(fmt.Errorf("unexpected column type: %v", cr.Cols()[i].Type))
						}
					}(cr, i)
					if got != want {
						tt.t.Errorf("column %d does not have the wanted length (%v) -want/+got:\n\t- %d\n\t+ %d", i, tbl.Key(), want, got)
					}
				}
				return nil
			})
		})
		tt.finish(true)
	})
}

func colReadersEqual(a, b flux.ColReader) bool {
	if a.Len() != b.Len() {
		return false
	}

	if len(a.Cols()) != len(b.Cols()) {
		return false
	}

	for i, n := 0, len(a.Cols()); i < n; i++ {
		if a.Cols()[i] != b.Cols()[i] {
			return false
		}

		// Compare the memory address for the arrow buffer.
		// The same buffer contents with different memory addresses
		// are considered different.
		switch a.Cols()[i].Type {
		case flux.TFloat:
			if a.Floats(i) != b.Floats(i) {
				return false
			}
		case flux.TInt:
			if a.Ints(i) != b.Ints(i) {
				return false
			}
		case flux.TUInt:
			if a.UInts(i) != b.UInts(i) {
				return false
			}
		case flux.TString:
			if a.Strings(i) != b.Strings(i) {
				return false
			}
		case flux.TBool:
			if a.Bools(i) != b.Bools(i) {
				return false
			}
		case flux.TTime:
			if a.Times(i) != b.Times(i) {
				return false
			}
		}
	}
	return true
}
