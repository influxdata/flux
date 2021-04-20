package table

import "github.com/influxdata/flux"

// Copy returns a buffered copy of the table and consumes the
// input table. If the input table is already buffered, it "consumes"
// the input and returns the same table.
//
// The buffered table can then be copied additional times using the
// BufferedTable.Copy method.
//
// This method should be used sparingly if at all. It will retain
// each of the buffers of data coming out of a table so the entire
// table is materialized in memory. For large datasets, this could
// potentially cause a problem. The allocator is meant to catch when
// this happens and prevent it.
func Copy(t flux.Table) (flux.BufferedTable, error) {
	if tbl, ok := t.(flux.BufferedTable); ok {
		return tbl, nil
	}

	tbl := tableBuffer{
		key:     t.Key(),
		colMeta: t.Cols(),
	}
	if t.Empty() {
		return &tbl, nil
	}

	if err := t.Do(func(cr flux.ColReader) error {
		cr.Retain()
		tbl.buffers = append(tbl.buffers, cr)
		return nil
	}); err != nil {
		tbl.Done()
		return nil, err
	}
	return &tbl, nil
}

// tableBuffer maintains a buffer of the data within a table.
// It is created by reading a table and using Retain to retain
// a reference to each ColReader that is returned.
//
// This implements the flux.BufferedTable interface.
type tableBuffer struct {
	key     flux.GroupKey
	colMeta []flux.ColMeta
	i       int
	buffers []flux.ColReader
}

func (tb *tableBuffer) Key() flux.GroupKey {
	return tb.key
}

func (tb *tableBuffer) Cols() []flux.ColMeta {
	return tb.colMeta
}

func (tb *tableBuffer) Do(f func(flux.ColReader) error) error {
	defer tb.Done()
	for ; tb.i < len(tb.buffers); tb.i++ {
		b := tb.buffers[tb.i]
		if err := f(b); err != nil {
			return err
		}
		b.Release()
	}
	return nil
}

func (tb *tableBuffer) Done() {
	for ; tb.i < len(tb.buffers); tb.i++ {
		tb.buffers[tb.i].Release()
	}
}

func (tb *tableBuffer) Empty() bool {
	return len(tb.buffers) == 0
}

func (tb *tableBuffer) Buffer(i int) flux.ColReader {
	return tb.buffers[i]
}

func (tb *tableBuffer) BufferN() int {
	return len(tb.buffers)
}

func (tb *tableBuffer) Copy() flux.BufferedTable {
	for i := tb.i; i < len(tb.buffers); i++ {
		tb.buffers[i].Retain()
	}
	return &tableBuffer{
		key:     tb.key,
		colMeta: tb.colMeta,
		i:       tb.i,
		buffers: tb.buffers,
	}
}
