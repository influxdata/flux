package table

import (
	"github.com/apache/arrow/go/arrow/array"
	"github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
)

// BufferedBuilder is a table builder that constructs
// a BufferedTable with zero or more buffers.
type BufferedBuilder struct {
	GroupKey  flux.GroupKey
	Columns   []flux.ColMeta
	Buffers   []*arrow.TableBuffer
	Allocator memory.Allocator
}

// NewBufferedBuilder constructs a new BufferedBuilder.
func NewBufferedBuilder(key flux.GroupKey, mem memory.Allocator) *BufferedBuilder {
	return &BufferedBuilder{
		GroupKey:  key,
		Allocator: mem,
	}
}

// GetBufferedBuilder is a convenience method for retrieving a
// BufferedBuilder from the BuilderCache.
func GetBufferedBuilder(key flux.GroupKey, cache *BuilderCache) (builder *BufferedBuilder, created bool) {
	created = cache.Get(key, &builder)
	return builder, created
}

// AppendBuffer will append a new buffer to this table builder.
// It ensures the schemas are compatible and will backfill previous
// buffers with nil for new columns that didn't previously exist.
func (b *BufferedBuilder) AppendBuffer(cr flux.ColReader) error {
	if len(b.Buffers) == 0 {
		// If there are no buffers, then take the columns
		// from the column reader and append the buffer directly.
		b.Columns = cr.Cols()
		buffer := &arrow.TableBuffer{
			GroupKey: b.GroupKey,
			Columns:  b.Columns,
			Values:   make([]array.Interface, len(b.Columns)),
		}
		for j := range buffer.Values {
			buffer.Values[j] = Values(cr, j)
			buffer.Values[j].Retain()
		}
		b.Buffers = []*arrow.TableBuffer{buffer}
		return nil
	}

	// Normalize the columns by adding any missing ones
	// and ensuring the existing columns are the same.
	mem := b.getAllocator()
	if err := b.normalizeTableSchema(cr.Cols(), mem); err != nil {
		return err
	}

	// Construct a table buffer and put the arrays in the correct index.
	buffer := &arrow.TableBuffer{
		GroupKey: b.GroupKey,
		Columns:  b.Columns,
		Values:   make([]array.Interface, len(b.Columns)),
	}
	for j, c := range b.Columns {
		idx := execute.ColIdx(c.Label, cr.Cols())
		if idx < 0 {
			// This column existed in a previous table, but
			// doesn't exist in this one so we need to generate
			// a null buffer.
			buffer.Values[j] = b.newNullColumn(c.Type, cr.Len(), mem)
			continue
		}
		buffer.Values[j] = Values(cr, idx)
		buffer.Values[j].Retain()
	}
	b.Buffers = append(b.Buffers, buffer)
	return nil
}

// normalizeTableSchema will ensure the table schema for this builder
// contains all of the columns in the list and that the columns with
// the same name have the same type. This returns an error if there
// is a schema collision.
func (b *BufferedBuilder) normalizeTableSchema(cols []flux.ColMeta, mem memory.Allocator) error {
	for _, c := range cols {
		idx := execute.ColIdx(c.Label, b.Columns)
		if idx < 0 {
			// New column. Add the column and backfill the previous
			// buffers with null values.
			b.Columns = append(b.Columns, c)
			for _, buf := range b.Buffers {
				buf.Columns = append(buf.Columns, c)
				buf.Values = append(buf.Values, b.newNullColumn(c.Type, buf.Len(), mem))
			}
			continue
		}

		// Verify the column type is the same.
		if ec := b.Columns[idx]; ec.Type != c.Type {
			return errors.Newf(codes.FailedPrecondition, "schema collision detected: column \"%s\" is both of type %s and %s", c.Label, c.Type, ec.Type)
		}
	}
	return nil
}

// newNullColumn will construct a new column with only null values
// for the entire size. The resulting array will match the column
// type that is passed in.
func (b *BufferedBuilder) newNullColumn(typ flux.ColType, l int, mem memory.Allocator) array.Interface {
	builder := arrow.NewBuilder(typ, mem)
	builder.Resize(l)
	for i := 0; i < l; i++ {
		builder.AppendNull()
	}
	return builder.NewArray()
}

func (b *BufferedBuilder) getAllocator() memory.Allocator {
	mem := b.Allocator
	if mem == nil {
		mem = memory.DefaultAllocator
	}
	return mem
}

func (b *BufferedBuilder) Table() (flux.Table, error) {
	buffers := make([]flux.ColReader, 0, len(b.Buffers))
	for _, buf := range b.Buffers {
		buffers = append(buffers, buf)
	}
	b.Buffers = nil
	return &BufferedTable{
		GroupKey: b.GroupKey,
		Columns:  b.Columns,
		Buffers:  buffers,
	}, nil
}

func (b *BufferedBuilder) Release() {
	for _, buf := range b.Buffers {
		buf.Release()
	}
}
