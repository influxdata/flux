package table

import (
	"github.com/apache/arrow/go/arrow/array"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/arrow"
)

// View is a view of a Table.
// The view is divided into a set of rows with a common
// set of columns known as the group key.
// The view does not provide a full view of the entire group key
// and a Table is not guaranteed to have rows ordered by the group key.
type View struct {
	buf arrow.TableBuffer
}

// ViewFromBuffer will create a View from the TableBuffer.
func ViewFromBuffer(buf arrow.TableBuffer) View {
	return View{buf: buf}
}

// ViewFromReader will create a View from the ColReader.
func ViewFromReader(cr flux.ColReader) View {
	buf := arrow.TableBuffer{
		GroupKey: cr.Key(),
		Columns:  cr.Cols(),
		Values:   make([]array.Interface, len(cr.Cols())),
	}
	for j := range buf.Values {
		buf.Values[j] = Values(cr, j)
	}
	return ViewFromBuffer(buf)
}

// Key returns the columns which are common for each row this view.
func (v View) Key() flux.GroupKey {
	return v.buf.Key()
}

// Buffer returns the underlying TableBuffer used for this View.
// This is exposed for use by another package, but this method
// should never be invoked in normal code.
func (v View) Buffer() arrow.TableBuffer {
	return v.buf
}

// NCols returns the number of columns in this View.
func (v View) NCols() int {
	return len(v.buf.Columns)
}

// Len returns the number of rows.
func (v View) Len() int {
	return v.buf.Len()
}

// Cols returns the columns as a slice.
func (v View) Cols() []flux.ColMeta {
	return v.buf.Columns
}

// Col returns the metadata associated with the column.
func (v View) Col(j int) flux.ColMeta {
	return v.buf.Columns[j]
}

// Index returns the index of the column with the given name.
func (v View) Index(label string) int {
	for j, c := range v.buf.Columns {
		if c.Label == label {
			return j
		}
	}
	return -1
}

// HasCol returns whether a column with the given name exists.
func (v View) HasCol(label string) bool {
	return v.Index(label) >= 0
}

// Values returns the array of values in this View.
// This will retain a new reference to the array which
// must be released afterwards.
func (v View) Values(j int) array.Interface {
	values := v.buf.Values[j]
	values.Retain()
	return values
}

// Borrow returns the array of values in this View.
// It will not increase the reference count of the array.
func (v View) Borrow(j int) array.Interface {
	return v.buf.Values[j]
}

// Retain will retain a reference to this View.
func (v View) Retain() {
	v.buf.Retain()
}

// Release will release a reference to this buffer.
func (v View) Release() {
	v.buf.Release()
}

// Reserve will ensure that there is space to
// add n additional columns to the View.
func (v *View) Reserve(n int) {
	if sz := len(v.buf.Columns); cap(v.buf.Columns) < sz+n {
		meta := make([]flux.ColMeta, sz, sz+n)
		copy(meta, v.buf.Columns)
		v.buf.Columns = meta

		values := make([]array.Interface, sz, sz+n)
		copy(values, v.buf.Values)
		v.buf.Values = values
	}
}

func (v *View) AddColumn(label string, typ flux.ColType, values array.Interface) {
	v.buf.Columns = append(v.buf.Columns, flux.ColMeta{
		Label: label,
		Type:  typ,
	})

	// If we were given a nil set for values, create a builder
	// and a default column for this row.
	if values == nil {
		values = arrow.Empty(typ)
	}
	v.buf.Values = append(v.buf.Values, values)
}

// CopySchema will copy the schema of a View
// and allocate a new slice of values for that schema.
func (v View) CopySchema() arrow.TableBuffer {
	return arrow.TableBuffer{
		GroupKey: v.Key(),
		Columns:  v.Cols(),
		Values:   make([]array.Interface, v.NCols()),
	}
}
