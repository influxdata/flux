package tablebuilder

import (
	"errors"
	"fmt"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/staticarray"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

// Instance builds tables using an underlying column constructor.
type Instance struct {
	// alloc is the allocator used for the array builders.
	alloc *memory.Allocator

	// key contains the group key for this table. This group key is specified
	// by WithGroupKey and will only be set if AddKeyValue is not called.
	// If AddKeyValue is called, the keyBuilder will be used instead.
	key flux.GroupKey

	// keyBuilder contains an in-progress key that is being built.
	keyBuilder execute.GroupKeyBuilder

	// keyBuilderKeys keeps track of the actual keys in the key builder since
	// the key builder does not do that.
	keyBuilderKeys map[string]struct{}

	// columns contains the various builders for different columns.
	columns []array.BaseBuilder

	// indexes contains an index of existing column names to the actual column.
	indexes map[string]int

	// size is the size of the table in rows. This is used as a hint
	// for automatically reserving the necessary space in new and existing
	// columns with a call to Resize. If there are no value columns,
	// then this is the size of the table. If this number mismatches
	// the actual length of the table columns, the table columns will
	// determine the length of the table.
	size int
}

// New will construct a new table builder with the given allocator.
func New(a *memory.Allocator) *Instance {
	return &Instance{alloc: a}
}

// WithGroupKey will use the given group key for this table. If group key entries have
// already been added, this will append or replace the values within the constructed group key.
// See AddKeyValue for details of how the group key is constructed.
func (b *Instance) WithGroupKey(key flux.GroupKey) (*Instance, error) {
	if b.key == nil && b.keyBuilder.Len() == 0 {
		// If we have indexes already, we have to validate that we aren't causing a conflict.
		if len(b.indexes) > 0 {
			for _, c := range key.Cols() {
				if _, ok := b.indexes[c.Label]; ok {
					// This error message should be correct since we don't have a group key yet.
					// So the column in the index can't be part of the group key.
					return nil, fmt.Errorf("key %q is already present as a column", c.Label)
				}
			}
		}

		// Set the key and exit. Nothing else to do since we aren't making a copy.
		b.key = key

		// Add indexes for each of the keys with a nil builder.
		if b.indexes == nil {
			b.indexes = make(map[string]int)
		}
		for _, c := range key.Cols() {
			b.indexes[c.Label] = len(b.columns)
			b.columns = append(b.columns, nil)
		}
		return b, nil
	}

	// We need to make a copy so this method is the equivalent of calling AddKeyValue on each
	// entry in the group key.
	for i, c := range key.Cols() {
		if err := b.AddKeyValue(c.Label, key.Value(i)); err != nil {
			return nil, err
		}
	}
	return b, nil
}

// NRows lists the number of rows that would be created if the table
// was constructed at that moment assuming that there are no errors.
// If there would be an error in table construction, this function is
// not guaranteed to return the correct length.
func (b *Instance) NRows() int {
	// Look for the first valid builder column and return that size.
	// If the columns are unequal, then building a table would return
	// an error and we don't care.
	for _, c := range b.columns {
		if c != nil {
			return c.Len()
		}
	}
	// The table is composed of only group keys so the size is whatever
	// Resize set it to.
	return b.size
}

// AddKeyValue will add an additional column to the table and mark it as part of the group key.
// The column will not be modifiable as group keys remain consistent within the table.
// The column type is automatically inferred from the value.
func (b *Instance) AddKeyValue(key string, value values.Value) error {
	// Verify that the key is not already present in a column and it is not already in
	// the key.
	if b.isInGroupKey(key) {
		return fmt.Errorf("key %q is already in the group key", key)
	} else if _, ok := b.indexes[key]; ok {
		return fmt.Errorf("key %q is already present as a column", key)
	}

	// Initialize the mapping of current keys in the key builder if not previously done.
	if b.keyBuilderKeys == nil {
		b.keyBuilderKeys = make(map[string]struct{})
	}

	if b.indexes == nil {
		b.indexes = make(map[string]int)
	}

	// We can add the key.
	if b.key != nil {
		// We now need to use the group key builder so let's add the existing keys to that.
		for i, c := range b.key.Cols() {
			b.keyBuilder.AddKeyValue(c.Label, b.key.Value(i))
			b.keyBuilderKeys[c.Label] = struct{}{}
			b.indexes[c.Label] = len(b.columns)
			b.columns = append(b.columns, nil)
		}
		b.key = nil
	}

	b.keyBuilder.AddKeyValue(key, value)
	b.keyBuilderKeys[key] = struct{}{}
	b.indexes[key] = len(b.columns)
	b.columns = append(b.columns, nil)
	return nil
}

// isInGroupKey checks if the given name is within the group key.
func (b *Instance) isInGroupKey(name string) bool {
	if b.key != nil {
		return b.key.HasCol(name)
	}
	_, ok := b.keyBuilderKeys[name]
	return ok
}

// Resize will call resize on each column and ensure that the table has n columns.
// This is important to call before constructing the table to ensure that it has
// the appropriate number of columns. If a table is only comprised of group keys,
// the size specified here determines how many rows there are.
func (b *Instance) Resize(n int) {
	for _, c := range b.columns {
		if c == nil {
			continue
		}
		c.Reserve(n)
	}
	b.size = n
}

// TODO(jsternberg): Consider a good way to deal with a building method that handles access
// to schema changes that are compatible. This is, for example, when we are constructing a table from
// multiple inputs and one of the inputs does not have a certain column in its output schema. The
// number of columns in the output will be wrong in this circumstance.

// column retrieves a column of the given name. If the column doesn't exist, it uses the builderFn
// to construct the builder.
func (b *Instance) column(name string, builderFn func() array.BaseBuilder) (int, bool, array.BaseBuilder) {
	idx, ok := b.indexes[name]
	if !ok {
		idx = len(b.columns)
		b.columns = append(b.columns, builderFn())
		if b.indexes == nil {
			b.indexes = make(map[string]int)
		}
		b.indexes[name] = idx

		if b.size > 0 {
			b.columns[idx].Reserve(b.size)
		}
	}
	return idx, !ok, b.columns[idx]
}

// FloatColumn is a wrapper around an array.FloatColumn that contains additional information
// for the builder about the current state of the column. The additional attributes are read
// only and modifying them will not affect anything.
type FloatColumn struct {
	// Name contains the name of the column.
	Name string

	// Index contains the current index of the column.
	Index int

	// Created is set to true when the column has been newly created.
	// This is useful if an algorithm needs to access columns multiple times and expects
	// the schema to change during the algorithm.
	Created bool

	// BaseBuilder holds the underlying builder. If this column is part of the group key,
	// then the builder will be nil.
	array.BaseBuilder
}

// Floats will find a column with the given name and invoke the function on the builder for that
// column. If a column with that name does not exist, it is created and appended to the Instance.
// If a column of that name exists with a different type or it is already added as part of the group key,
// the Do function will return an error.
func (b *Instance) Floats(name string) FloatColumn {
	idx, created, builder := b.column(name, b.newFloatBuilder)
	return FloatColumn{
		Name:        name,
		Index:       idx,
		Created:     created,
		BaseBuilder: builder,
	}
}

func (b *Instance) newFloatBuilder() array.BaseBuilder {
	return staticarray.FloatBuilder(b.alloc)
}

func (b FloatColumn) Append(v float64) error {
	return b.Do(func(b array.FloatBuilder) error {
		b.Append(v)
		return nil
	})
}

func (b FloatColumn) AppendValues(v []float64, valid ...[]bool) error {
	return b.Do(func(b array.FloatBuilder) error {
		b.AppendValues(v, valid...)
		return nil
	})
}

func (b FloatColumn) Do(fn func(b array.FloatBuilder) error) error {
	if b.BaseBuilder == nil {
		return fmt.Errorf("%s is part of the group key and cannot be modified", b.Name)
	}

	// Try typecasting to the correct type.
	builder, ok := b.BaseBuilder.(array.FloatBuilder)
	if !ok {
		return fmt.Errorf("incompatible column type: %s", b.BaseBuilder.Type())
	}
	return fn(builder)
}

type IntColumn struct {
	// Name contains the name of the column.
	Name string

	// Index contains the current index of the column.
	Index int

	// Created is set to true when the column has been newly created.
	// This is useful if an algorithm needs to access columns multiple times and expects
	// the schema to change during the algorithm.
	Created bool

	// BaseBuilder holds the underlying builder. If this column is part of the group key,
	// then the builder will be nil.
	array.BaseBuilder
}

func (b *Instance) Ints(name string) IntColumn {
	idx, created, builder := b.column(name, b.newIntBuilder)
	return IntColumn{
		Name:        name,
		Index:       idx,
		Created:     created,
		BaseBuilder: builder,
	}
}

func (b *Instance) newIntBuilder() array.BaseBuilder {
	return staticarray.IntBuilder(b.alloc)
}

func (b IntColumn) Append(v int64) error {
	return b.Do(func(b array.IntBuilder) error {
		b.Append(v)
		return nil
	})
}

func (b IntColumn) AppendValues(v []int64, valid ...[]bool) error {
	return b.Do(func(b array.IntBuilder) error {
		b.AppendValues(v, valid...)
		return nil
	})
}

func (b IntColumn) Do(fn func(b array.IntBuilder) error) error {
	if b.BaseBuilder == nil {
		return fmt.Errorf("%s is part of the group key and cannot be modified", b.Name)
	}

	// Try typecasting to the correct type.
	builder, ok := b.BaseBuilder.(array.IntBuilder)
	if !ok {
		return fmt.Errorf("incompatible column type: %s", b.BaseBuilder.Type())
	}
	return fn(builder)
}

type UIntColumn struct {
	// Name contains the name of the column.
	Name string

	// Index contains the current index of the column.
	Index int

	// Created is set to true when the column has been newly created.
	// This is useful if an algorithm needs to access columns multiple times and expects
	// the schema to change during the algorithm.
	Created bool

	// BaseBuilder holds the underlying builder. If this column is part of the group key,
	// then the builder will be nil.
	array.BaseBuilder
}

func (b *Instance) UInts(name string) UIntColumn {
	idx, created, builder := b.column(name, b.newUIntBuilder)
	return UIntColumn{
		Name:        name,
		Index:       idx,
		Created:     created,
		BaseBuilder: builder,
	}
}

func (b *Instance) newUIntBuilder() array.BaseBuilder {
	return staticarray.UIntBuilder(b.alloc)
}

func (b UIntColumn) Append(v uint64) error {
	return b.Do(func(b array.UIntBuilder) error {
		b.Append(v)
		return nil
	})
}

func (b UIntColumn) AppendValues(v []uint64, valid ...[]bool) error {
	return b.Do(func(b array.UIntBuilder) error {
		b.AppendValues(v, valid...)
		return nil
	})
}

func (b UIntColumn) Do(fn func(b array.UIntBuilder) error) error {
	if b.BaseBuilder == nil {
		return fmt.Errorf("%s is part of the group key and cannot be modified", b.Name)
	}

	// Try typecasting to the correct type.
	builder, ok := b.BaseBuilder.(array.UIntBuilder)
	if !ok {
		return fmt.Errorf("incompatible column type: %s", b.BaseBuilder.Type())
	}
	return fn(builder)
}

type StringColumn struct {
	// Name contains the name of the column.
	Name string

	// Index contains the current index of the column.
	Index int

	// Created is set to true when the column has been newly created.
	// This is useful if an algorithm needs to access columns multiple times and expects
	// the schema to change during the algorithm.
	Created bool

	// BaseBuilder holds the underlying builder. If this column is part of the group key,
	// then the builder will be nil.
	array.BaseBuilder
}

func (b *Instance) Strings(name string) StringColumn {
	idx, created, builder := b.column(name, b.newStringBuilder)
	return StringColumn{
		Name:        name,
		Index:       idx,
		Created:     created,
		BaseBuilder: builder,
	}
}

func (b *Instance) newStringBuilder() array.BaseBuilder {
	return staticarray.StringBuilder(b.alloc)
}

func (b StringColumn) Append(v string) error {
	return b.Do(func(b array.StringBuilder) error {
		b.Append(v)
		return nil
	})
}

func (b StringColumn) AppendValues(v []string, valid ...[]bool) error {
	return b.Do(func(b array.StringBuilder) error {
		b.AppendValues(v, valid...)
		return nil
	})
}

func (b StringColumn) Do(fn func(b array.StringBuilder) error) error {
	if b.BaseBuilder == nil {
		return fmt.Errorf("%s is part of the group key and cannot be modified", b.Name)
	}

	// Try typecasting to the correct type.
	builder, ok := b.BaseBuilder.(array.StringBuilder)
	if !ok {
		return fmt.Errorf("incompatible column type: %s", b.BaseBuilder.Type())
	}
	return fn(builder)
}

type BoolColumn struct {
	// Name contains the name of the column.
	Name string

	// Index contains the current index of the column.
	Index int

	// Created is set to true when the column has been newly created.
	// This is useful if an algorithm needs to access columns multiple times and expects
	// the schema to change during the algorithm.
	Created bool

	// BaseBuilder holds the underlying builder. If this column is part of the group key,
	// then the builder will be nil.
	array.BaseBuilder
}

func (b *Instance) Bools(name string) BoolColumn {
	idx, created, builder := b.column(name, b.newBoolBuilder)
	return BoolColumn{
		Name:        name,
		Index:       idx,
		Created:     created,
		BaseBuilder: builder,
	}
}

func (b *Instance) newBoolBuilder() array.BaseBuilder {
	return staticarray.BooleanBuilder(b.alloc)
}

func (b BoolColumn) Append(v bool) error {
	return b.Do(func(b array.BooleanBuilder) error {
		b.Append(v)
		return nil
	})
}

func (b BoolColumn) AppendValues(v []bool, valid ...[]bool) error {
	return b.Do(func(b array.BooleanBuilder) error {
		b.AppendValues(v, valid...)
		return nil
	})
}

func (b BoolColumn) Do(fn func(b array.BooleanBuilder) error) error {
	if b.BaseBuilder == nil {
		return fmt.Errorf("%s is part of the group key and cannot be modified", b.Name)
	}

	// Try typecasting to the correct type.
	builder, ok := b.BaseBuilder.(array.BooleanBuilder)
	if !ok {
		return fmt.Errorf("incompatible column type: %s", b.BaseBuilder.Type())
	}
	return fn(builder)
}

type TimeColumn struct {
	// Name contains the name of the column.
	Name string

	// Index contains the current index of the column.
	Index int

	// Created is set to true when the column has been newly created.
	// This is useful if an algorithm needs to access columns multiple times and expects
	// the schema to change during the algorithm.
	Created bool

	// BaseBuilder holds the underlying builder. If this column is part of the group key,
	// then the builder will be nil.
	array.BaseBuilder
}

func (b *Instance) Times(name string) TimeColumn {
	idx, created, builder := b.column(name, b.newTimeBuilder)
	return TimeColumn{
		Name:        name,
		Index:       idx,
		Created:     created,
		BaseBuilder: builder,
	}
}

func (b *Instance) newTimeBuilder() array.BaseBuilder {
	return staticarray.TimeBuilder(b.alloc)
}

func (b TimeColumn) Append(v values.Time) error {
	return b.Do(func(b array.TimeBuilder) error {
		b.Append(v)
		return nil
	})
}

func (b TimeColumn) AppendValues(v []values.Time, valid ...[]bool) error {
	return b.Do(func(b array.TimeBuilder) error {
		b.AppendValues(v, valid...)
		return nil
	})
}

func (b TimeColumn) Do(fn func(b array.TimeBuilder) error) error {
	if b.BaseBuilder == nil {
		return fmt.Errorf("%s is part of the group key and cannot be modified", b.Name)
	}

	// Try typecasting to the correct type.
	builder, ok := b.BaseBuilder.(array.TimeBuilder)
	if !ok {
		return fmt.Errorf("incompatible column type: %s", b.BaseBuilder.Type())
	}
	return fn(builder)
}

type ValueColumn struct {
	// Name contains the name of the column.
	Name string

	// Index contains the current index of the column.
	Index int

	// Created is set to true when the column has been newly created.
	// This is useful if an algorithm needs to access columns multiple times and expects
	// the schema to change during the algorithm.
	Created bool

	// Err holds any errors that happened on column creation.
	Err error

	// ValueBuilder holds the underlying builder. If this column is part of the group key,
	// then the builder will be nil. If there was an error in column
	// creation, this may not contain a value value.
	*array.ValueBuilder
}

// Values will return a ValueBuilder interface. If the column does not
// exist, the passed in column type will be used.
func (b *Instance) Values(name string, typ flux.ColType) ValueColumn {
	idx, created, builder := b.column(name, b.newValueBuilder(typ))
	column := ValueColumn{
		Name:         name,
		Index:        idx,
		Created:      created,
		ValueBuilder: &array.ValueBuilder{Builder: builder},
	}
	if !created {
		if btype := flux.ColumnType(builder.Type()); typ != btype {
			column.Err = fmt.Errorf("conflicting column types: %s != %s", btype, typ)
		}
	}
	return column
}

func (b ValueColumn) Do(fn func(b *array.ValueBuilder) error) error {
	if b.Err != nil {
		return b.Err
	}
	return fn(b.ValueBuilder)
}

func (b *Instance) newValueBuilder(typ flux.ColType) func() array.BaseBuilder {
	switch typ {
	case flux.TFloat:
		return b.newFloatBuilder
	case flux.TInt:
		return b.newIntBuilder
	case flux.TUInt:
		return b.newUIntBuilder
	case flux.TString:
		return b.newStringBuilder
	case flux.TBool:
		return b.newBoolBuilder
	case flux.TTime:
		return b.newTimeBuilder
	default:
		// TODO(jsternberg): Probably find some way to have this not panic.
		panic(fmt.Sprintf("unsupported column type: %s", typ))
	}
}

func (b *Instance) AppendFloat(name string, v float64) error {
	return b.Floats(name).Append(v)
}

func (b *Instance) AppendInt(name string, v int64) error {
	return b.Ints(name).Append(v)
}

func (b *Instance) AppendUInt(name string, v uint64) error {
	return b.UInts(name).Append(v)
}

func (b *Instance) AppendString(name string, v string) error {
	return b.Strings(name).Append(v)
}

func (b *Instance) AppendBool(name string, v bool) error {
	return b.Bools(name).Append(v)
}

func (b *Instance) AppendTime(name string, v values.Time) error {
	return b.Times(name).Append(v)
}

func (b *Instance) AppendValue(name string, v values.Value) error {
	return b.Values(name, flux.ColumnType(v.Type())).Append(v)
}

// Table will construct the table and return it. If there was an error constructing the table,
// this will return an error.
func (b *Instance) Table() (flux.Table, error) {
	// A table with no columns cannot be built.
	if len(b.columns) == 0 {
		return nil, errors.New("table is empty")
	}

	// Ensure that all columns are the same length. If the builder for a column is nil,
	// it's part of the group key so skip it.
	sz := -1
	for _, c := range b.columns {
		if c == nil {
			// This is a group key so skip it.
		} else if sz == -1 {
			sz = c.Len()
		} else if c.Len() != sz {
			return nil, fmt.Errorf("incompatible column lengths")
		}
	}

	if sz == -1 {
		// Only use the specified size if we don't have a better value from the columns.
		sz = b.size
	}

	// Generate the key.
	key := b.key
	if key == nil {
		k, err := b.keyBuilder.Build()
		if err != nil {
			return nil, err
		}
		key = k
	}

	// Instantiate the tag columns and copy them to a new slice.
	columns := make([]array.Base, len(b.columns))
	colMeta := make([]flux.ColMeta, len(b.columns))
	for name, idx := range b.indexes {
		if c := b.columns[idx]; c != nil {
			columns[idx] = c.BuildArray()
		} else {
			switch v := key.LabelValue(name); v.Type() {
			case semantic.Float:
				arr := make([]float64, sz)
				for i := 0; i < sz; i++ {
					arr[i] = v.Float()
				}
				columns[idx] = staticarray.Float(arr)
			case semantic.Int:
				arr := make([]int64, sz)
				for i := 0; i < sz; i++ {
					arr[i] = v.Int()
				}
				columns[idx] = staticarray.Int(arr)
			case semantic.UInt:
				arr := make([]uint64, sz)
				for i := 0; i < sz; i++ {
					arr[i] = v.UInt()
				}
				columns[idx] = staticarray.UInt(arr)
			case semantic.String:
				arr := make([]string, sz)
				for i := 0; i < sz; i++ {
					arr[i] = v.Str()
				}
				columns[idx] = staticarray.String(arr)
			case semantic.Bool:
				arr := make([]bool, sz)
				for i := 0; i < sz; i++ {
					arr[i] = v.Bool()
				}
				columns[idx] = staticarray.Boolean(arr)
			default:
				panic(fmt.Sprintf("invalid type: %s", v.Type()))
			}
		}

		// Set the column metadata.
		colMeta[idx] = flux.ColMeta{
			Label: name,
			Type:  flux.ColumnType(columns[idx].Type()),
		}
	}

	// Return the table.
	return &table{
		key:     key,
		colMeta: colMeta,
		columns: columns,
		sz:      sz,
	}, nil
}
