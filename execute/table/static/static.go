// Package static provides utilities for easily constructing static
// tables that are meant for tests.
//
// The primary type is Table which will be a mapping of columns to their data.
// The data is defined in a columnar format instead of a row-based one.
//
// The implementations in this package are not performant and are not meant
// to be used in production code. They are good enough for small datasets that
// are present in tests to ensure code correctness.
package static

import (
	"fmt"
	"sort"
	"time"

	stdarrow "github.com/apache/arrow/go/arrow"
	"github.com/apache/arrow/go/arrow/array"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/values"
)

// Table is a statically constructed table.
// It is a mapping between column names and the column.
//
// This is not a performant section of code and it is primarily
// meant to make writing unit tests easy. Do not use in
// production code.
type Table map[string]Column

func (s Table) Key() flux.GroupKey {
	var cols []flux.ColMeta
	for label, c := range s {
		if c.IsKey() {
			cols = append(cols, flux.ColMeta{
				Label: label,
				Type:  c.Type(),
			})
		}
	}
	sort.Slice(cols, func(i, j int) bool {
		return cols[i].Label < cols[j].Label
	})

	vs := make([]values.Value, len(cols))
	for i, col := range cols {
		vs[i] = s[col.Label].KeyValue()
	}
	return execute.NewGroupKey(cols, vs)
}

func (s Table) Cols() []flux.ColMeta {
	cols := make([]flux.ColMeta, 0, len(s))
	for label, c := range s {
		cols = append(cols, flux.ColMeta{
			Label: label,
			Type:  c.Type(),
		})
	}
	sort.Slice(cols, func(i, j int) bool {
		return cols[i].Label < cols[j].Label
	})
	return cols
}

func (s Table) Do(f func(flux.ColReader) error) error {
	buffer := arrow.TableBuffer{
		GroupKey: s.Key(),
		Columns:  s.Cols(),
	}

	// Determine the size by looking at the first non-key column.
	n := 0
	for _, c := range s {
		if c.IsKey() {
			continue
		}
		n = c.Len()
		break
	}

	// Table is empty.
	if n == 0 {
		return nil
	}

	// Construct each of the buffers.
	buffer.Values = make([]array.Interface, len(buffer.Columns))
	for i, col := range buffer.Columns {
		buffer.Values[i] = s[col.Label].Make(n)
	}

	if err := buffer.Validate(); err != nil {
		return err
	}
	return f(&buffer)
}

func (s Table) Done() {}

func (s Table) Empty() bool {
	for _, c := range s {
		if !c.IsKey() && c.Len() > 0 {
			return false
		}
	}
	return true
}

// Column is the definition for how to construct a column for the table.
type Column interface {
	// Type returns the column type for this column.
	Type() flux.ColType

	// Make will construct an array with the given length
	// if it is possible.
	Make(n int) array.Interface

	// Len will return the length of this column.
	// If no length is known, this will return -1.
	Len() int

	// IsKey will return true if this is part of the group key.
	IsKey() bool

	// KeyValue will return the key value if this column is part
	// of the group key.
	KeyValue() values.Value
}

// IntKey will construct a group key with the integer type.
// The value can be an int, int64, or nil.
func IntKey(v interface{}) Column {
	if iv, ok := mustIntValue(v); ok {
		return keyColumn{v: iv, t: flux.TInt}
	}
	return keyColumn{t: flux.TInt}
}

// UintKey will construct a group key with the unsigned type.
// The value can be a uint, uint64, int, int64, or nil.
func UintKey(v interface{}) Column {
	if iv, ok := mustUintValue(v); ok {
		return keyColumn{v: iv, t: flux.TUInt}
	}
	return keyColumn{t: flux.TUInt}
}

// FloatKey will construct a group key with the float type.
// The value can be a float64, int, int64, or nil.
func FloatKey(v interface{}) Column {
	if iv, ok := mustFloatValue(v); ok {
		return keyColumn{v: iv, t: flux.TFloat}
	}
	return keyColumn{t: flux.TFloat}
}

// StringKey will construct a group key with the string type.
// The value can be a string or nil.
func StringKey(v interface{}) Column {
	if iv, ok := mustStringValue(v); ok {
		return keyColumn{v: iv, t: flux.TString}
	}
	return keyColumn{t: flux.TString}
}

// BooleanKey will construct a group key with the boolean type.
// The value can be a bool or nil.
func BooleanKey(v interface{}) Column {
	if iv, ok := mustBooleanValue(v); ok {
		return keyColumn{v: iv, t: flux.TBool}
	}
	return keyColumn{t: flux.TBool}
}

// TimeKey will construct a group key with the given time using either a
// string or an integer. If an integer is used, then it is in seconds.
func TimeKey(v interface{}) Column {
	if iv, _, ok := mustTimeValue(v, 0, time.Second); ok {
		return keyColumn{v: execute.Time(iv), t: flux.TTime}
	}
	return keyColumn{t: flux.TTime}
}

type keyColumn struct {
	v interface{}
	t flux.ColType
}

func (s keyColumn) Make(n int) array.Interface {
	return arrow.Repeat(s.KeyValue(), n, memory.DefaultAllocator)
}

func (s keyColumn) Type() flux.ColType     { return s.t }
func (s keyColumn) Len() int               { return -1 }
func (s keyColumn) IsKey() bool            { return true }
func (s keyColumn) KeyValue() values.Value { return values.New(s.v) }

// Ints will construct an array of integers.
// Each value can be an int, int64, or nil.
func Ints(v ...interface{}) Column {
	c := intColumn{v: make([]int64, len(v))}
	for i, iv := range v {
		val, ok := mustIntValue(iv)
		if !ok {
			if c.valid == nil {
				c.valid = make([]bool, len(v))
				for i := range c.valid {
					c.valid[i] = true
				}
			}
			c.valid[i] = false
		}
		c.v[i] = val
	}
	return c
}

type intColumn struct {
	v     []int64
	valid []bool
}

func (s intColumn) Make(n int) array.Interface {
	b := array.NewInt64Builder(memory.DefaultAllocator)
	b.Resize(len(s.v))
	b.AppendValues(s.v, s.valid)
	return b.NewArray()
}

func (s intColumn) Type() flux.ColType     { return flux.TInt }
func (s intColumn) Len() int               { return len(s.v) }
func (s intColumn) IsKey() bool            { return false }
func (s intColumn) KeyValue() values.Value { return values.InvalidValue }

func mustIntValue(v interface{}) (int64, bool) {
	if v == nil {
		return 0, false
	}

	switch v := v.(type) {
	case int:
		return int64(v), true
	case int64:
		return v, true
	default:
		panic(fmt.Sprintf("unable to convert type %T to an int value", v))
	}
}

// Uints will construct an array of unsigned integers.
// Each value can be a uint, uint64, int, int64, or nil.
func Uints(v ...interface{}) Column {
	c := uintColumn{v: make([]uint64, len(v))}
	for i, iv := range v {
		val, ok := mustUintValue(iv)
		if !ok {
			if c.valid == nil {
				c.valid = make([]bool, len(v))
				for i := range c.valid {
					c.valid[i] = true
				}
			}
			c.valid[i] = false
		}
		c.v[i] = val
	}
	return c
}

type uintColumn struct {
	v     []uint64
	valid []bool
}

func (s uintColumn) Make(n int) array.Interface {
	b := array.NewUint64Builder(memory.DefaultAllocator)
	b.Resize(len(s.v))
	b.AppendValues(s.v, s.valid)
	return b.NewArray()
}

func (s uintColumn) Type() flux.ColType     { return flux.TUInt }
func (s uintColumn) Len() int               { return len(s.v) }
func (s uintColumn) IsKey() bool            { return false }
func (s uintColumn) KeyValue() values.Value { return values.InvalidValue }

func mustUintValue(v interface{}) (uint64, bool) {
	if v == nil {
		return 0, false
	}

	switch v := v.(type) {
	case int:
		return uint64(v), true
	case int64:
		return uint64(v), true
	case uint:
		return uint64(v), true
	case uint64:
		return v, true
	default:
		panic(fmt.Sprintf("unable to convert type %T to a uint value", v))
	}
}

// Floats will construct an array of floats.
// Each value can be a float64, int, int64, or nil.
func Floats(v ...interface{}) Column {
	c := floatColumn{v: make([]float64, len(v))}
	for i, iv := range v {
		val, ok := mustFloatValue(iv)
		if !ok {
			if c.valid == nil {
				c.valid = make([]bool, len(v))
				for i := range c.valid {
					c.valid[i] = true
				}
			}
			c.valid[i] = false
		}
		c.v[i] = val
	}
	return c
}

type floatColumn struct {
	v     []float64
	valid []bool
}

func (s floatColumn) Make(n int) array.Interface {
	b := array.NewFloat64Builder(memory.DefaultAllocator)
	b.Resize(len(s.v))
	b.AppendValues(s.v, s.valid)
	return b.NewArray()
}

func (s floatColumn) Type() flux.ColType     { return flux.TFloat }
func (s floatColumn) Len() int               { return len(s.v) }
func (s floatColumn) IsKey() bool            { return false }
func (s floatColumn) KeyValue() values.Value { return values.InvalidValue }

func mustFloatValue(v interface{}) (float64, bool) {
	if v == nil {
		return 0, false
	}

	switch v := v.(type) {
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case float64:
		return v, true
	default:
		panic(fmt.Sprintf("unable to convert type %T to a float value", v))
	}
}

// Strings will construct an array of strings.
// Each value can be a string or nil.
func Strings(v ...interface{}) Column {
	c := stringColumn{v: make([]string, len(v))}
	for i, iv := range v {
		val, ok := mustStringValue(iv)
		if !ok {
			if c.valid == nil {
				c.valid = make([]bool, len(v))
				for i := range c.valid {
					c.valid[i] = true
				}
			}
			c.valid[i] = false
		}
		c.v[i] = val
	}
	return c
}

type stringColumn struct {
	v     []string
	valid []bool
}

func (s stringColumn) Make(n int) array.Interface {
	b := array.NewBinaryBuilder(memory.DefaultAllocator, stdarrow.BinaryTypes.String)
	b.Resize(len(s.v))
	b.AppendStringValues(s.v, s.valid)
	return b.NewArray()
}

func (s stringColumn) Type() flux.ColType     { return flux.TString }
func (s stringColumn) Len() int               { return len(s.v) }
func (s stringColumn) IsKey() bool            { return false }
func (s stringColumn) KeyValue() values.Value { return values.InvalidValue }

func mustStringValue(v interface{}) (string, bool) {
	if v == nil {
		return "", false
	}

	switch v := v.(type) {
	case string:
		return v, true
	default:
		panic(fmt.Sprintf("unable to convert type %T to a string value", v))
	}
}

// Booleans will construct an array of booleans.
// Each value can be a bool or nil.
func Booleans(v ...interface{}) Column {
	c := booleanColumn{v: make([]bool, len(v))}
	for i, iv := range v {
		val, ok := mustBooleanValue(iv)
		if !ok {
			if c.valid == nil {
				c.valid = make([]bool, len(v))
				for i := range c.valid {
					c.valid[i] = true
				}
			}
			c.valid[i] = false
		}
		c.v[i] = val
	}
	return c
}

type booleanColumn struct {
	v     []bool
	valid []bool
}

func (s booleanColumn) Make(n int) array.Interface {
	b := array.NewBooleanBuilder(memory.DefaultAllocator)
	b.Resize(len(s.v))
	b.AppendValues(s.v, s.valid)
	return b.NewArray()
}

func (s booleanColumn) Type() flux.ColType     { return flux.TBool }
func (s booleanColumn) Len() int               { return len(s.v) }
func (s booleanColumn) IsKey() bool            { return false }
func (s booleanColumn) KeyValue() values.Value { return values.InvalidValue }

func mustBooleanValue(v interface{}) (bool, bool) {
	if v == nil {
		return false, false
	}

	switch v := v.(type) {
	case bool:
		return v, true
	default:
		panic(fmt.Sprintf("unable to convert type %T to a boolean value", v))
	}
}

// Times will construct an array of times with the given time using either a
// string or an integer. If an integer is used, then it is in seconds.
//
// If strings and integers are mixed, the integers will be treates as offsets
// from the last string time that was used.
func Times(v ...interface{}) Column {
	var offset int64
	c := timeColumn{v: make([]int64, len(v))}
	for i, iv := range v {
		val, abs, ok := mustTimeValue(iv, offset, time.Second)
		if !ok {
			if c.valid == nil {
				c.valid = make([]bool, len(v))
				for i := range c.valid {
					c.valid[i] = true
				}
			}
			c.valid[i] = false
		}
		if abs {
			offset = val
		}
		c.v[i] = val
	}
	return c
}

type timeColumn struct {
	v     []int64
	valid []bool
}

func (s timeColumn) Make(n int) array.Interface {
	b := array.NewInt64Builder(memory.DefaultAllocator)
	b.Resize(len(s.v))
	b.AppendValues(s.v, s.valid)
	return b.NewArray()
}

func (s timeColumn) Type() flux.ColType     { return flux.TTime }
func (s timeColumn) Len() int               { return len(s.v) }
func (s timeColumn) IsKey() bool            { return false }
func (s timeColumn) KeyValue() values.Value { return values.InvalidValue }

// mustTimeValue will convert the interface into a time value.
// This must either be an int-like value or a string that can be
// parsed as a time in RFC3339 format.
//
// This will panic otherwise.
func mustTimeValue(v interface{}, offset int64, unit time.Duration) (t int64, abs, ok bool) {
	if v == nil {
		return 0, false, false
	}

	switch v := v.(type) {
	case int:
		return offset + int64(v)*int64(unit), false, true
	case int64:
		return offset + v*int64(unit), false, true
	case string:
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			if t, err = time.Parse(time.RFC3339Nano, v); err != nil {
				panic(err)
			}
		}
		return t.UnixNano(), true, true
	default:
		panic(fmt.Sprintf("unable to convert type %T to a time value", v))
	}
}

// Extend will copy the spec from the table and add or override
// any columns for the extended table.
func (s Table) Extend(table Table) Table {
	ns := make(Table, len(s)+len(table))
	for label, column := range s {
		ns[label] = column
	}
	for label, column := range table {
		ns[label] = column
	}
	return ns
}
