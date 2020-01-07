package table

import (
	"reflect"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
)

// Builder is the minimum interface for constructing a Table.
type Builder interface {
	// Table will construct a Table from the existing contents.
	// Invoking this method should reset the builder and all allocated
	// memory will be owned by the returned flux.Table.
	Table() (flux.Table, error)

	// Release will release the buffered contents from the builder.
	// This method is unnecessary if Table is called.
	Release()
}

// BuilderCache hold a mapping of group keys to Builder.
// When a Builder is requested for a specific group key,
// the BuilderCache will return a Builder that is unique
// for that GroupKey.
type BuilderCache struct {
	// New will be called to construct a new Builder
	// when a GroupKey that hasn't been seen before is
	// requested. The returned Builder should be empty.
	New func(key flux.GroupKey) Builder

	tables *execute.GroupLookup
}

// Get retrieves the Builder for this group key.
// If one doesn't exist, it will invoke the New function and
// store it within the Builder.
// If the builder was newly created, this method returns true
// for the second parameter.
// The interface must be a pointer to the type that is created
// from the New method. This method will use reflection to set
// the value of the pointer.
func (d *BuilderCache) Get(key flux.GroupKey, b interface{}) bool {
	builder, ok := d.lookupState(key)
	if !ok {
		if d.tables == nil {
			d.tables = execute.NewGroupLookup()
		}
		builder = d.New(key)
		d.tables.Set(key, builder)
	}
	r := reflect.ValueOf(b)
	r.Elem().Set(reflect.ValueOf(builder))
	return !ok
}

// Table will remove a builder from the cache and construct a flux.Table
// from the buffered contents.
func (d *BuilderCache) Table(key flux.GroupKey) (flux.Table, error) {
	builder, ok := d.lookupState(key)
	if !ok {
		return nil, errors.Newf(codes.Internal, "table not found with key %v", key)
	}
	return builder.Table()
}

func (d *BuilderCache) ForEach(f func(key flux.GroupKey, builder Builder) error) error {
	if d.tables == nil {
		return nil
	}
	var err error
	d.tables.Range(func(key flux.GroupKey, value interface{}) {
		if err != nil {
			return
		}
		builder := value.(Builder)
		err = f(key, builder)
	})
	return err
}

func (d *BuilderCache) lookupState(key flux.GroupKey) (Builder, bool) {
	if d.tables == nil {
		return nil, false
	}
	v, ok := d.tables.Lookup(key)
	if !ok {
		return nil, false
	}
	return v.(Builder), true
}

func (d *BuilderCache) DiscardTable(key flux.GroupKey) {
	if d.tables == nil {
		return
	}

	if b, ok := d.lookupState(key); ok {
		// If the builder supports Clear, then call that method.
		if builder, ok := b.(interface {
			Clear()
		}); ok {
			builder.Clear()
		} else {
			// Release the table and construct a new one.
			b.Release()
			d.tables.Set(key, d.New(key))
		}
	}
}

func (d *BuilderCache) ExpireTable(key flux.GroupKey) {
	if d.tables == nil {
		return
	}
	ts, ok := d.tables.Delete(key)
	if ok {
		ts.(Builder).Release()
	}
}
