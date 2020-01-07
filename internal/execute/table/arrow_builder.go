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

// ArrowBuilder is a Builder that uses arrow array builders
// as the underlying builder mechanism.
type ArrowBuilder struct {
	GroupKey  flux.GroupKey
	Columns   []flux.ColMeta
	Builders  []array.Builder
	Allocator memory.Allocator
}

// NewArrowBuilder constructs a new ArrowBuilder.
func NewArrowBuilder(key flux.GroupKey, mem memory.Allocator) *ArrowBuilder {
	return &ArrowBuilder{
		GroupKey:  key,
		Allocator: mem,
	}
}

// GetArrowBuilder is a convenience method for retrieving an
// ArrowBuilder from the BuilderCache.
func GetArrowBuilder(key flux.GroupKey, cache *BuilderCache) (builder *ArrowBuilder, created bool) {
	created = cache.Get(key, &builder)
	return builder, created
}

func (a *ArrowBuilder) Key() flux.GroupKey {
	return a.GroupKey
}

func (a *ArrowBuilder) Cols() []flux.ColMeta {
	return a.Columns
}

// AddCol will add a column with the given metadata.
// If the column exists, an error is returned.
func (a *ArrowBuilder) AddCol(c flux.ColMeta) (int, error) {
	if execute.ColIdx(c.Label, a.Columns) >= 0 {
		return -1, errors.Newf(codes.Invalid, "table builder already has a column with label %s", c.Label)
	}

	// Retrieve the memory allocator or use the default.
	mem := a.Allocator
	if mem == nil {
		mem = memory.DefaultAllocator
	}

	// Determine the current size of all of the builders.
	n := 0
	if len(a.Builders) > 0 {
		n = a.Builders[0].Len()
	}
	for i := 1; i < len(a.Builders); i++ {
		if a.Builders[i].Len() != n {
			return -1, errors.Newf(codes.Internal, "column %d (len: %d) has a different size than the first column (len: %d)", i, a.Builders[i].Len(), n)
		}
	}

	// Create a builder and append null values to match the default size.
	b := arrow.NewBuilder(c.Type, mem)
	if n > 0 {
		b.Reserve(n)
		for i := 0; i < n; i++ {
			b.AppendNull()
		}
	}
	a.Columns = append(a.Columns, c)
	a.Builders = append(a.Builders, b)
	return len(a.Columns) - 1, nil
}

// CheckCol will check if a column exists with the label
// and the same type. This will return an error if the column
// does not exist or has an incompatible type.
func (a *ArrowBuilder) CheckCol(c flux.ColMeta) (int, error) {
	idx := execute.ColIdx(c.Label, a.Columns)
	if idx < 0 {
		return -1, errors.Newf(codes.NotFound, "table builder is missing a column with label %s", c.Label)
	} else if ec := a.Columns[idx]; ec.Type != c.Type {
		return -1, errors.Newf(codes.FailedPrecondition, "schema collision detected: column \"%s\" is both of type %s and %s", c.Label, c.Type, ec.Type)
	}
	return idx, nil
}

// Table constructs a flux.Table from the current builders.
func (a *ArrowBuilder) Table() (flux.Table, error) {
	values := make([]array.Interface, len(a.Builders))
	for j, b := range a.Builders {
		values[j] = b.NewArray()
	}
	buffer := &arrow.TableBuffer{
		GroupKey: a.GroupKey,
		Columns:  a.Columns,
		Values:   values,
	}
	if err := buffer.Validate(); err != nil {
		return nil, err
	}
	return FromBuffer(buffer), nil
}

func (a *ArrowBuilder) Release() {
	for _, b := range a.Builders {
		b.Release()
	}
}
