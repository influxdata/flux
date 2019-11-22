package arrow

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/apache/arrow/go/arrow"
	"github.com/apache/arrow/go/arrow/array"
	"github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

// Repeat will construct an arrow array that repeats the value n times.
func Repeat(v values.Value, n int, mem memory.Allocator) array.Interface {
	switch v.Type() {
	case semantic.Int:
		return RepeatInt(v.Int(), n, mem)
	case semantic.UInt:
		return RepeatUint(v.UInt(), n, mem)
	case semantic.Float:
		return RepeatFloat(v.Float(), n, mem)
	case semantic.String:
		return RepeatString(v.Str(), n, mem)
	case semantic.Bool:
		return RepeatBoolean(v.Bool(), n, mem)
	case semantic.Time:
		return RepeatInt(int64(v.Time()), n, mem)
	default:
		panic(fmt.Errorf("unknown builder for type: %s", v.Type()))
	}
}

// RepeatInt will return an array that repeats an integer n times.
func RepeatInt(v int64, n int, mem memory.Allocator) array.Interface {
	return &lazyArray{
		refCount: 1,
		dataType: arrow.PrimitiveTypes.Int64,
		length:   n,
		init: func(mem memory.Allocator) array.Interface {
			b := array.NewInt64Builder(mem)
			b.Resize(n)
			for i := 0; i < n; i++ {
				b.Append(v)
			}
			return b.NewArray()
		},
		mem: mem,
	}
}

// RepeatUint will return an array that repeats an unsigned integer n times.
func RepeatUint(v uint64, n int, mem memory.Allocator) array.Interface {
	return &lazyArray{
		refCount: 1,
		dataType: arrow.PrimitiveTypes.Uint64,
		length:   n,
		init: func(mem memory.Allocator) array.Interface {
			b := array.NewUint64Builder(mem)
			b.Resize(n)
			for i := 0; i < n; i++ {
				b.Append(v)
			}
			return b.NewArray()
		},
		mem: mem,
	}
}

// RepeatFloat will return an array that repeats a float n times.
func RepeatFloat(v float64, n int, mem memory.Allocator) array.Interface {
	return &lazyArray{
		refCount: 1,
		dataType: arrow.PrimitiveTypes.Float64,
		length:   n,
		init: func(mem memory.Allocator) array.Interface {
			b := array.NewFloat64Builder(mem)
			b.Resize(n)
			for i := 0; i < n; i++ {
				b.Append(v)
			}
			return b.NewArray()
		},
		mem: mem,
	}
}

// RepeatString will return an array that repeats a string n times.
func RepeatString(v string, n int, mem memory.Allocator) array.Interface {
	return &lazyArray{
		refCount: 1,
		dataType: arrow.BinaryTypes.String,
		length:   n,
		init: func(mem memory.Allocator) array.Interface {
			b := array.NewBinaryBuilder(mem, arrow.BinaryTypes.String)
			b.Resize(n)
			b.ReserveData(n * len(v))
			for i := 0; i < n; i++ {
				b.AppendString(v)
			}
			return b.NewArray()
		},
		mem: mem,
	}
}

// RepeatBoolean will return an array that repeats a boolean n times.
func RepeatBoolean(v bool, n int, mem memory.Allocator) array.Interface {
	return &lazyArray{
		refCount: 1,
		dataType: arrow.FixedWidthTypes.Boolean,
		length:   n,
		init: func(mem memory.Allocator) array.Interface {
			// TODO(jsternberg): I think this can be optimized further
			// because this should just use memset.
			b := array.NewBooleanBuilder(mem)
			b.Resize(n)
			for i := 0; i < n; i++ {
				b.Append(v)
			}
			return b.NewArray()
		},
		mem: mem,
	}
}

// lazyArray implements a lazily initialized array.
// When a method on the array.Interface gets called, this will
// materialize the array exactly once.
//
// This can be used for marking that an array of a certain type
// exists, but wait until it is used until it is initialized.
// In order to appropriately use these with their realized values,
// the AsXXX methods should be used.
type lazyArray struct {
	refCount int32
	init     func(memory.Allocator) array.Interface
	mem      memory.Allocator

	dataType arrow.DataType
	array    array.Interface
	length   int
	mu       sync.RWMutex
}

func (l *lazyArray) get() array.Interface {
	l.mu.RLock()
	if l.array != nil {
		arr := l.array
		l.mu.RUnlock()
		return arr
	}
	l.mu.RUnlock()

	// The array has not been created.
	// Take out the write lock and try to make it.
	l.mu.Lock()
	if l.array == nil {
		l.array = l.init(l.mem)
	}
	arr := l.array
	l.mu.Unlock()
	return arr
}

func (l *lazyArray) DataType() arrow.DataType {
	return l.dataType
}

func (l *lazyArray) NullN() int {
	return l.get().NullN()
}

func (l *lazyArray) NullBitmapBytes() []byte {
	return l.get().NullBitmapBytes()
}

func (l *lazyArray) IsNull(i int) bool {
	return l.get().IsNull(i)
}

func (l *lazyArray) IsValid(i int) bool {
	return l.get().IsValid(i)
}

func (l *lazyArray) Data() *array.Data {
	return l.get().Data()
}

func (l *lazyArray) Len() int {
	return l.length
}

func (l *lazyArray) Retain() {
	atomic.AddInt32(&l.refCount, 1)
}

func (l *lazyArray) Release() {
	if atomic.AddInt32(&l.refCount, -1) == 0 {
		l.mu.Lock()
		if l.array != nil {
			l.array.Release()
			l.array = nil
		}
		l.mu.Unlock()
	}
}

// As will assign the underlying array to the target.
func (l *lazyArray) As(target interface{}) bool {
	return As(l.get(), target)
}
