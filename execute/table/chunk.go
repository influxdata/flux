package table

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/arrow"
)

// Chunk is a horizontal partition of a Table. It is a subset of rows
// and contains a set of columns known as the group key.
// It may not contain all columns that have been seen associated with
// that group key so transformations should verify the existence of columns
// for each chunk independently.
type Chunk struct {
	buf arrow.TableBuffer
}

// ChunkFromBuffer will create a Chunk from the TableBuffer.
//
// This function takes ownership of the arrow.TableBuffer
// and the Chunk goes out of scope at the same time
// as the arrow.TableBuffer unless Retain is called.
func ChunkFromBuffer(buf arrow.TableBuffer) Chunk {
	return Chunk{buf: buf}
}

// ChunkFromReader will create a Chunk from the ColReader.
//
// This function borrows a reference to the data in the ColReader
// and will go out of scope at the same time as the ColReader
// unless Retain is called.
func ChunkFromReader(cr flux.ColReader) Chunk {
	buf := arrow.TableBuffer{
		GroupKey: cr.Key(),
		Columns:  cr.Cols(),
		Values:   make([]array.Interface, len(cr.Cols())),
	}
	for j := range buf.Values {
		buf.Values[j] = Values(cr, j)
	}
	return ChunkFromBuffer(buf)
}

// Key returns the columns which are common for each row this view.
func (v Chunk) Key() flux.GroupKey {
	return v.buf.Key()
}

// Buffer returns the underlying TableBuffer used for this Chunk.
// This is exposed for use by another package, but this method
// should never be invoked in normal code.
func (v Chunk) Buffer() arrow.TableBuffer {
	return v.buf
}

// NCols returns the number of columns in this Chunk.
func (v Chunk) NCols() int {
	return len(v.buf.Columns)
}

// Len returns the number of rows.
func (v Chunk) Len() int {
	return v.buf.Len()
}

// Cols returns the columns as a slice.
func (v Chunk) Cols() []flux.ColMeta {
	return v.buf.Columns
}

// Col returns the metadata associated with the column.
func (v Chunk) Col(j int) flux.ColMeta {
	return v.buf.Columns[j]
}

// Index returns the index of the column with the given name.
func (v Chunk) Index(label string) int {
	for j, c := range v.buf.Columns {
		if c.Label == label {
			return j
		}
	}
	return -1
}

// HasCol returns whether a column with the given name exists.
func (v Chunk) HasCol(label string) bool {
	return v.Index(label) >= 0
}

// Values returns a reference to the array of values in this Chunk.
// The returned array is a borrowed reference and the caller can
// call Retain on the returned array to retain its own reference
// to the array.
func (v Chunk) Values(j int) array.Interface {
	return v.buf.Values[j]
}

// Bools is a convenience function for retrieving an array
// as a boolean array.
func (v Chunk) Bools(j int) *array.Boolean {
	return v.Values(j).(*array.Boolean)
}

// Ints is a convenience function for retrieving an array
// as an int array.
func (v Chunk) Ints(j int) *array.Int {
	return v.Values(j).(*array.Int)
}

// Uints is a convenience function for retrieving an array
// as a uint array.
func (v Chunk) Uints(j int) *array.Uint {
	return v.Values(j).(*array.Uint)
}

// Floats is a convenience function for retrieving an array
// as a float array.
func (v Chunk) Floats(j int) *array.Float {
	return v.Values(j).(*array.Float)
}

// Strings is a convenience function for retrieving an array
// as a string array.
func (v Chunk) Strings(j int) *array.String {
	return v.Values(j).(*array.String)
}

// Retain will retain a reference to this Chunk.
func (v Chunk) Retain() {
	v.buf.Retain()
}

// Release will release a reference to this buffer.
func (v Chunk) Release() {
	v.buf.Release()
}
