package table

import (
	"github.com/apache/arrow/go/arrow/array"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/values"
)

// Mask will return a no-copy Table that masks the given
// columns. If the columns are part of the group key,
// they will be removed from the key.
//
// This function will not attempt any regrouping with
// other tables. This function should only be used when
// it is known that the group key will not conflict with
// others and the Table needs to have certain columns
// filtered either for display or other purposes.
func Mask(tbl flux.Table, columns []string) flux.Table {
	key := tbl.Key()
	keyCols := make([]flux.ColMeta, 0, len(key.Cols()))
	keyValues := make([]values.Value, 0, cap(keyCols))
	for j, c := range key.Cols() {
		if execute.ContainsStr(columns, c.Label) {
			continue
		}
		keyCols = append(keyCols, c)
		keyValues = append(keyValues, key.Value(j))
	}

	offsets := make([]int, 0, len(tbl.Cols()))
	cols := make([]flux.ColMeta, 0, cap(offsets))
	for j, c := range tbl.Cols() {
		if execute.ContainsStr(columns, c.Label) {
			continue
		}
		cols = append(cols, c)
		offsets = append(offsets, j-len(offsets))
	}
	return &maskTable{
		key:     execute.NewGroupKey(keyCols, keyValues),
		cols:    cols,
		table:   tbl,
		offsets: offsets,
	}
}

type maskTable struct {
	key     flux.GroupKey
	cols    []flux.ColMeta
	table   flux.Table
	offsets []int
}

func (m *maskTable) Key() flux.GroupKey {
	return m.key
}

func (m *maskTable) Cols() []flux.ColMeta {
	return m.cols
}

func (m *maskTable) Do(f func(flux.ColReader) error) error {
	return m.table.Do(func(cr flux.ColReader) error {
		view := maskTableView{
			key:     m.key,
			cols:    m.cols,
			reader:  cr,
			offsets: m.offsets,
		}
		return f(&view)
	})
}

func (m *maskTable) Done() {
	m.table.Done()
}

func (m *maskTable) Empty() bool {
	return m.table.Empty()
}

type maskTableView struct {
	key     flux.GroupKey
	cols    []flux.ColMeta
	reader  flux.ColReader
	offsets []int
}

func (m *maskTableView) Key() flux.GroupKey {
	return m.key
}

func (m *maskTableView) Cols() []flux.ColMeta {
	return m.cols
}

func (m *maskTableView) Len() int                    { return m.reader.Len() }
func (m *maskTableView) Bools(j int) *array.Boolean  { return m.reader.Bools(j + m.offsets[j]) }
func (m *maskTableView) Ints(j int) *array.Int64     { return m.reader.Ints(j + m.offsets[j]) }
func (m *maskTableView) UInts(j int) *array.Uint64   { return m.reader.UInts(j + m.offsets[j]) }
func (m *maskTableView) Floats(j int) *array.Float64 { return m.reader.Floats(j + m.offsets[j]) }
func (m *maskTableView) Strings(j int) *array.Binary { return m.reader.Strings(j + m.offsets[j]) }
func (m *maskTableView) Times(j int) *array.Int64    { return m.reader.Times(j + m.offsets[j]) }
func (m *maskTableView) Retain()                     { m.reader.Retain() }
func (m *maskTableView) Release()                    { m.reader.Release() }
