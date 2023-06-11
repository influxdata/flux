// Generated by tmpl
// https://github.com/benbjohnson/tmpl
//
// DO NOT EDIT!
// Source: pivot.gen.go.tmpl

package universe

import (
	"github.com/InfluxCommunity/flux"
	"github.com/InfluxCommunity/flux/array"
	"github.com/InfluxCommunity/flux/codes"
	"github.com/InfluxCommunity/flux/internal/arrowutil"
	"github.com/InfluxCommunity/flux/internal/errors"
	"github.com/apache/arrow/go/v7/arrow/memory"
)

//lint:file-ignore U1000 Ignore all unused code, it's generated

// mergeKeys finds all the unique values of the row key across each buffer,
// and return them in a single array sorted in ascending order.
func (gr *pivotTableGroup) mergeKeys(mem memory.Allocator) array.Array {
	switch gr.rowCol.Type {

	case flux.TInt:
		return gr.mergeIntKeys(mem)

	case flux.TUInt:
		return gr.mergeUintKeys(mem)

	case flux.TFloat:
		return gr.mergeFloatKeys(mem)

	case flux.TString:
		return gr.mergeStringKeys(mem)

	case flux.TTime:
		return gr.mergeTimeKeys(mem)

	default:
		panic(errors.Newf(codes.Unimplemented, "row column merge not implemented for %s", gr.rowCol.Type))
	}
}

func (gr *pivotTableGroup) buildColumn(keys array.Array, buf *pivotTableBuffer, mem memory.Allocator) array.Array {
	switch gr.rowCol.Type {

	case flux.TInt:
		return gr.buildColumnFromInts(keys.(*array.Int), buf, mem)

	case flux.TUInt:
		return gr.buildColumnFromUints(keys.(*array.Uint), buf, mem)

	case flux.TFloat:
		return gr.buildColumnFromFloats(keys.(*array.Float), buf, mem)

	case flux.TString:
		return gr.buildColumnFrom(keys.(*array.String), buf, mem)

	case flux.TTime:
		return gr.buildColumnFromTimes(keys.(*array.Int), buf, mem)

	default:
		panic(errors.Newf(codes.Unimplemented, "row column merge not implemented for %s", gr.rowCol.Type))
	}
}

func (gr *pivotTableGroup) mergeIntKeys(mem memory.Allocator) array.Array {
	buffers := make([][]array.Array, 0, len(gr.buffers))
	for _, buf := range gr.buffers {
		buffers = append(buffers, buf.keys)
	}

	count := 0
	gr.forEachInt(buffers, func(v int64) {
		count++
	})

	b := arrowutil.NewIntBuilder(mem)
	b.Resize(count)
	gr.forEachInt(buffers, b.Append)
	return b.NewArray()
}

func (gr *pivotTableGroup) forEachInt(buffers [][]array.Array, fn func(v int64)) {
	iterators := make([]*arrowutil.IntIterator, 0, len(buffers))
	for _, vs := range buffers {
		itr := arrowutil.IterateInts(vs)
		if !itr.Next() {
			continue
		}
		iterators = append(iterators, &itr)
	}

	// Count the number of common keys.
	for len(iterators) > 0 {
		next := iterators[0].Value()
		for _, itr := range iterators[1:] {
			if v := itr.Value(); v < next {
				next = v
			}
		}

		// This counts as a row.
		fn(next)

		// Advance any iterators to the next non-null value
		// that match the next value.
		for i := 0; i < len(iterators); {
			itr := iterators[i]
			if itr.Value() != next {
				i++
				continue
			}

			// Advance to the next non-null value.
			for {
				if !itr.Next() {
					// Remove this iterator from the list.
					copy(iterators[i:], iterators[i+1:])
					iterators = iterators[:len(iterators)-1]
					break
				}

				if itr.IsValid() && itr.Value() != next {
					// The next value is valid so advance
					// to the next iterator.
					i++
					break
				}
			}
		}
	}
}

func (gr *pivotTableGroup) buildColumnFromInts(keys *array.Int, buf *pivotTableBuffer, mem memory.Allocator) array.Array {

	switch buf.valueType {

	case flux.TInt:
		return gr.buildIntColumnFromInts(keys, buf, mem)

	case flux.TUInt:
		return gr.buildUintColumnFromInts(keys, buf, mem)

	case flux.TFloat:
		return gr.buildFloatColumnFromInts(keys, buf, mem)

	case flux.TBool:
		return gr.buildBooleanColumnFromInts(keys, buf, mem)

	case flux.TString:
		return gr.buildStringColumnFromInts(keys, buf, mem)

	case flux.TTime:
		return gr.buildTimeColumnFromInts(keys, buf, mem)

	default:
		panic("unimplemented")
	}

}

func (gr *pivotTableGroup) buildIntColumnFromInts(keys *array.Int, buf *pivotTableBuffer, mem memory.Allocator) array.Array {
	b := arrowutil.NewIntBuilder(mem)
	b.Resize(keys.Len())

	kitr := arrowutil.IterateInts(buf.keys)
	vitr := arrowutil.IterateInts(buf.values)
	for i := 0; kitr.Next() && vitr.Next(); {
		for ; i < keys.Len(); i++ {
			if kitr.Value() == keys.Value(i) {
				if vitr.IsValid() {
					b.Append(vitr.Value())
				} else {
					b.AppendNull()
				}
				i++
				break
			}
			b.AppendNull()
		}
	}
	for i := b.Len(); i < keys.Len(); i++ {
		b.AppendNull()
	}
	return b.NewArray()
}

func (gr *pivotTableGroup) buildUintColumnFromInts(keys *array.Int, buf *pivotTableBuffer, mem memory.Allocator) array.Array {
	b := arrowutil.NewUintBuilder(mem)
	b.Resize(keys.Len())

	kitr := arrowutil.IterateInts(buf.keys)
	vitr := arrowutil.IterateUints(buf.values)
	for i := 0; kitr.Next() && vitr.Next(); {
		for ; i < keys.Len(); i++ {
			if kitr.Value() == keys.Value(i) {
				if vitr.IsValid() {
					b.Append(vitr.Value())
				} else {
					b.AppendNull()
				}
				i++
				break
			}
			b.AppendNull()
		}
	}
	for i := b.Len(); i < keys.Len(); i++ {
		b.AppendNull()
	}
	return b.NewArray()
}

func (gr *pivotTableGroup) buildFloatColumnFromInts(keys *array.Int, buf *pivotTableBuffer, mem memory.Allocator) array.Array {
	b := arrowutil.NewFloatBuilder(mem)
	b.Resize(keys.Len())

	kitr := arrowutil.IterateInts(buf.keys)
	vitr := arrowutil.IterateFloats(buf.values)
	for i := 0; kitr.Next() && vitr.Next(); {
		for ; i < keys.Len(); i++ {
			if kitr.Value() == keys.Value(i) {
				if vitr.IsValid() {
					b.Append(vitr.Value())
				} else {
					b.AppendNull()
				}
				i++
				break
			}
			b.AppendNull()
		}
	}
	for i := b.Len(); i < keys.Len(); i++ {
		b.AppendNull()
	}
	return b.NewArray()
}

func (gr *pivotTableGroup) buildBooleanColumnFromInts(keys *array.Int, buf *pivotTableBuffer, mem memory.Allocator) array.Array {
	b := arrowutil.NewBooleanBuilder(mem)
	b.Resize(keys.Len())

	kitr := arrowutil.IterateInts(buf.keys)
	vitr := arrowutil.IterateBooleans(buf.values)
	for i := 0; kitr.Next() && vitr.Next(); {
		for ; i < keys.Len(); i++ {
			if kitr.Value() == keys.Value(i) {
				if vitr.IsValid() {
					b.Append(vitr.Value())
				} else {
					b.AppendNull()
				}
				i++
				break
			}
			b.AppendNull()
		}
	}
	for i := b.Len(); i < keys.Len(); i++ {
		b.AppendNull()
	}
	return b.NewArray()
}

func (gr *pivotTableGroup) buildStringColumnFromInts(keys *array.Int, buf *pivotTableBuffer, mem memory.Allocator) array.Array {
	b := arrowutil.NewStringBuilder(mem)
	b.Resize(keys.Len())

	kitr := arrowutil.IterateInts(buf.keys)
	vitr := arrowutil.IterateStrings(buf.values)
	for i := 0; kitr.Next() && vitr.Next(); {
		for ; i < keys.Len(); i++ {
			if kitr.Value() == keys.Value(i) {
				if vitr.IsValid() {
					b.Append(vitr.Value())
				} else {
					b.AppendNull()
				}
				i++
				break
			}
			b.AppendNull()
		}
	}
	for i := b.Len(); i < keys.Len(); i++ {
		b.AppendNull()
	}
	return b.NewArray()
}

func (gr *pivotTableGroup) buildTimeColumnFromInts(keys *array.Int, buf *pivotTableBuffer, mem memory.Allocator) array.Array {
	b := arrowutil.NewIntBuilder(mem)
	b.Resize(keys.Len())

	kitr := arrowutil.IterateInts(buf.keys)
	vitr := arrowutil.IterateInts(buf.values)
	for i := 0; kitr.Next() && vitr.Next(); {
		for ; i < keys.Len(); i++ {
			if kitr.Value() == keys.Value(i) {
				if vitr.IsValid() {
					b.Append(vitr.Value())
				} else {
					b.AppendNull()
				}
				i++
				break
			}
			b.AppendNull()
		}
	}
	for i := b.Len(); i < keys.Len(); i++ {
		b.AppendNull()
	}
	return b.NewArray()
}

func (gr *pivotTableGroup) mergeUintKeys(mem memory.Allocator) array.Array {
	buffers := make([][]array.Array, 0, len(gr.buffers))
	for _, buf := range gr.buffers {
		buffers = append(buffers, buf.keys)
	}

	count := 0
	gr.forEachUint(buffers, func(v uint64) {
		count++
	})

	b := arrowutil.NewUintBuilder(mem)
	b.Resize(count)
	gr.forEachUint(buffers, b.Append)
	return b.NewArray()
}

func (gr *pivotTableGroup) forEachUint(buffers [][]array.Array, fn func(v uint64)) {
	iterators := make([]*arrowutil.UintIterator, 0, len(buffers))
	for _, vs := range buffers {
		itr := arrowutil.IterateUints(vs)
		if !itr.Next() {
			continue
		}
		iterators = append(iterators, &itr)
	}

	// Count the number of common keys.
	for len(iterators) > 0 {
		next := iterators[0].Value()
		for _, itr := range iterators[1:] {
			if v := itr.Value(); v < next {
				next = v
			}
		}

		// This counts as a row.
		fn(next)

		// Advance any iterators to the next non-null value
		// that match the next value.
		for i := 0; i < len(iterators); {
			itr := iterators[i]
			if itr.Value() != next {
				i++
				continue
			}

			// Advance to the next non-null value.
			for {
				if !itr.Next() {
					// Remove this iterator from the list.
					copy(iterators[i:], iterators[i+1:])
					iterators = iterators[:len(iterators)-1]
					break
				}

				if itr.IsValid() && itr.Value() != next {
					// The next value is valid so advance
					// to the next iterator.
					i++
					break
				}
			}
		}
	}
}

func (gr *pivotTableGroup) buildColumnFromUints(keys *array.Uint, buf *pivotTableBuffer, mem memory.Allocator) array.Array {

	switch buf.valueType {

	case flux.TInt:
		return gr.buildIntColumnFromUints(keys, buf, mem)

	case flux.TUInt:
		return gr.buildUintColumnFromUints(keys, buf, mem)

	case flux.TFloat:
		return gr.buildFloatColumnFromUints(keys, buf, mem)

	case flux.TBool:
		return gr.buildBooleanColumnFromUints(keys, buf, mem)

	case flux.TString:
		return gr.buildStringColumnFromUints(keys, buf, mem)

	case flux.TTime:
		return gr.buildTimeColumnFromUints(keys, buf, mem)

	default:
		panic("unimplemented")
	}

}

func (gr *pivotTableGroup) buildIntColumnFromUints(keys *array.Uint, buf *pivotTableBuffer, mem memory.Allocator) array.Array {
	b := arrowutil.NewIntBuilder(mem)
	b.Resize(keys.Len())

	kitr := arrowutil.IterateUints(buf.keys)
	vitr := arrowutil.IterateInts(buf.values)
	for i := 0; kitr.Next() && vitr.Next(); {
		for ; i < keys.Len(); i++ {
			if kitr.Value() == keys.Value(i) {
				if vitr.IsValid() {
					b.Append(vitr.Value())
				} else {
					b.AppendNull()
				}
				i++
				break
			}
			b.AppendNull()
		}
	}
	for i := b.Len(); i < keys.Len(); i++ {
		b.AppendNull()
	}
	return b.NewArray()
}

func (gr *pivotTableGroup) buildUintColumnFromUints(keys *array.Uint, buf *pivotTableBuffer, mem memory.Allocator) array.Array {
	b := arrowutil.NewUintBuilder(mem)
	b.Resize(keys.Len())

	kitr := arrowutil.IterateUints(buf.keys)
	vitr := arrowutil.IterateUints(buf.values)
	for i := 0; kitr.Next() && vitr.Next(); {
		for ; i < keys.Len(); i++ {
			if kitr.Value() == keys.Value(i) {
				if vitr.IsValid() {
					b.Append(vitr.Value())
				} else {
					b.AppendNull()
				}
				i++
				break
			}
			b.AppendNull()
		}
	}
	for i := b.Len(); i < keys.Len(); i++ {
		b.AppendNull()
	}
	return b.NewArray()
}

func (gr *pivotTableGroup) buildFloatColumnFromUints(keys *array.Uint, buf *pivotTableBuffer, mem memory.Allocator) array.Array {
	b := arrowutil.NewFloatBuilder(mem)
	b.Resize(keys.Len())

	kitr := arrowutil.IterateUints(buf.keys)
	vitr := arrowutil.IterateFloats(buf.values)
	for i := 0; kitr.Next() && vitr.Next(); {
		for ; i < keys.Len(); i++ {
			if kitr.Value() == keys.Value(i) {
				if vitr.IsValid() {
					b.Append(vitr.Value())
				} else {
					b.AppendNull()
				}
				i++
				break
			}
			b.AppendNull()
		}
	}
	for i := b.Len(); i < keys.Len(); i++ {
		b.AppendNull()
	}
	return b.NewArray()
}

func (gr *pivotTableGroup) buildBooleanColumnFromUints(keys *array.Uint, buf *pivotTableBuffer, mem memory.Allocator) array.Array {
	b := arrowutil.NewBooleanBuilder(mem)
	b.Resize(keys.Len())

	kitr := arrowutil.IterateUints(buf.keys)
	vitr := arrowutil.IterateBooleans(buf.values)
	for i := 0; kitr.Next() && vitr.Next(); {
		for ; i < keys.Len(); i++ {
			if kitr.Value() == keys.Value(i) {
				if vitr.IsValid() {
					b.Append(vitr.Value())
				} else {
					b.AppendNull()
				}
				i++
				break
			}
			b.AppendNull()
		}
	}
	for i := b.Len(); i < keys.Len(); i++ {
		b.AppendNull()
	}
	return b.NewArray()
}

func (gr *pivotTableGroup) buildStringColumnFromUints(keys *array.Uint, buf *pivotTableBuffer, mem memory.Allocator) array.Array {
	b := arrowutil.NewStringBuilder(mem)
	b.Resize(keys.Len())

	kitr := arrowutil.IterateUints(buf.keys)
	vitr := arrowutil.IterateStrings(buf.values)
	for i := 0; kitr.Next() && vitr.Next(); {
		for ; i < keys.Len(); i++ {
			if kitr.Value() == keys.Value(i) {
				if vitr.IsValid() {
					b.Append(vitr.Value())
				} else {
					b.AppendNull()
				}
				i++
				break
			}
			b.AppendNull()
		}
	}
	for i := b.Len(); i < keys.Len(); i++ {
		b.AppendNull()
	}
	return b.NewArray()
}

func (gr *pivotTableGroup) buildTimeColumnFromUints(keys *array.Uint, buf *pivotTableBuffer, mem memory.Allocator) array.Array {
	b := arrowutil.NewIntBuilder(mem)
	b.Resize(keys.Len())

	kitr := arrowutil.IterateUints(buf.keys)
	vitr := arrowutil.IterateInts(buf.values)
	for i := 0; kitr.Next() && vitr.Next(); {
		for ; i < keys.Len(); i++ {
			if kitr.Value() == keys.Value(i) {
				if vitr.IsValid() {
					b.Append(vitr.Value())
				} else {
					b.AppendNull()
				}
				i++
				break
			}
			b.AppendNull()
		}
	}
	for i := b.Len(); i < keys.Len(); i++ {
		b.AppendNull()
	}
	return b.NewArray()
}

func (gr *pivotTableGroup) mergeFloatKeys(mem memory.Allocator) array.Array {
	buffers := make([][]array.Array, 0, len(gr.buffers))
	for _, buf := range gr.buffers {
		buffers = append(buffers, buf.keys)
	}

	count := 0
	gr.forEachFloat(buffers, func(v float64) {
		count++
	})

	b := arrowutil.NewFloatBuilder(mem)
	b.Resize(count)
	gr.forEachFloat(buffers, b.Append)
	return b.NewArray()
}

func (gr *pivotTableGroup) forEachFloat(buffers [][]array.Array, fn func(v float64)) {
	iterators := make([]*arrowutil.FloatIterator, 0, len(buffers))
	for _, vs := range buffers {
		itr := arrowutil.IterateFloats(vs)
		if !itr.Next() {
			continue
		}
		iterators = append(iterators, &itr)
	}

	// Count the number of common keys.
	for len(iterators) > 0 {
		next := iterators[0].Value()
		for _, itr := range iterators[1:] {
			if v := itr.Value(); v < next {
				next = v
			}
		}

		// This counts as a row.
		fn(next)

		// Advance any iterators to the next non-null value
		// that match the next value.
		for i := 0; i < len(iterators); {
			itr := iterators[i]
			if itr.Value() != next {
				i++
				continue
			}

			// Advance to the next non-null value.
			for {
				if !itr.Next() {
					// Remove this iterator from the list.
					copy(iterators[i:], iterators[i+1:])
					iterators = iterators[:len(iterators)-1]
					break
				}

				if itr.IsValid() && itr.Value() != next {
					// The next value is valid so advance
					// to the next iterator.
					i++
					break
				}
			}
		}
	}
}

func (gr *pivotTableGroup) buildColumnFromFloats(keys *array.Float, buf *pivotTableBuffer, mem memory.Allocator) array.Array {

	switch buf.valueType {

	case flux.TInt:
		return gr.buildIntColumnFromFloats(keys, buf, mem)

	case flux.TUInt:
		return gr.buildUintColumnFromFloats(keys, buf, mem)

	case flux.TFloat:
		return gr.buildFloatColumnFromFloats(keys, buf, mem)

	case flux.TBool:
		return gr.buildBooleanColumnFromFloats(keys, buf, mem)

	case flux.TString:
		return gr.buildStringColumnFromFloats(keys, buf, mem)

	case flux.TTime:
		return gr.buildTimeColumnFromFloats(keys, buf, mem)

	default:
		panic("unimplemented")
	}

}

func (gr *pivotTableGroup) buildIntColumnFromFloats(keys *array.Float, buf *pivotTableBuffer, mem memory.Allocator) array.Array {
	b := arrowutil.NewIntBuilder(mem)
	b.Resize(keys.Len())

	kitr := arrowutil.IterateFloats(buf.keys)
	vitr := arrowutil.IterateInts(buf.values)
	for i := 0; kitr.Next() && vitr.Next(); {
		for ; i < keys.Len(); i++ {
			if kitr.Value() == keys.Value(i) {
				if vitr.IsValid() {
					b.Append(vitr.Value())
				} else {
					b.AppendNull()
				}
				i++
				break
			}
			b.AppendNull()
		}
	}
	for i := b.Len(); i < keys.Len(); i++ {
		b.AppendNull()
	}
	return b.NewArray()
}

func (gr *pivotTableGroup) buildUintColumnFromFloats(keys *array.Float, buf *pivotTableBuffer, mem memory.Allocator) array.Array {
	b := arrowutil.NewUintBuilder(mem)
	b.Resize(keys.Len())

	kitr := arrowutil.IterateFloats(buf.keys)
	vitr := arrowutil.IterateUints(buf.values)
	for i := 0; kitr.Next() && vitr.Next(); {
		for ; i < keys.Len(); i++ {
			if kitr.Value() == keys.Value(i) {
				if vitr.IsValid() {
					b.Append(vitr.Value())
				} else {
					b.AppendNull()
				}
				i++
				break
			}
			b.AppendNull()
		}
	}
	for i := b.Len(); i < keys.Len(); i++ {
		b.AppendNull()
	}
	return b.NewArray()
}

func (gr *pivotTableGroup) buildFloatColumnFromFloats(keys *array.Float, buf *pivotTableBuffer, mem memory.Allocator) array.Array {
	b := arrowutil.NewFloatBuilder(mem)
	b.Resize(keys.Len())

	kitr := arrowutil.IterateFloats(buf.keys)
	vitr := arrowutil.IterateFloats(buf.values)
	for i := 0; kitr.Next() && vitr.Next(); {
		for ; i < keys.Len(); i++ {
			if kitr.Value() == keys.Value(i) {
				if vitr.IsValid() {
					b.Append(vitr.Value())
				} else {
					b.AppendNull()
				}
				i++
				break
			}
			b.AppendNull()
		}
	}
	for i := b.Len(); i < keys.Len(); i++ {
		b.AppendNull()
	}
	return b.NewArray()
}

func (gr *pivotTableGroup) buildBooleanColumnFromFloats(keys *array.Float, buf *pivotTableBuffer, mem memory.Allocator) array.Array {
	b := arrowutil.NewBooleanBuilder(mem)
	b.Resize(keys.Len())

	kitr := arrowutil.IterateFloats(buf.keys)
	vitr := arrowutil.IterateBooleans(buf.values)
	for i := 0; kitr.Next() && vitr.Next(); {
		for ; i < keys.Len(); i++ {
			if kitr.Value() == keys.Value(i) {
				if vitr.IsValid() {
					b.Append(vitr.Value())
				} else {
					b.AppendNull()
				}
				i++
				break
			}
			b.AppendNull()
		}
	}
	for i := b.Len(); i < keys.Len(); i++ {
		b.AppendNull()
	}
	return b.NewArray()
}

func (gr *pivotTableGroup) buildStringColumnFromFloats(keys *array.Float, buf *pivotTableBuffer, mem memory.Allocator) array.Array {
	b := arrowutil.NewStringBuilder(mem)
	b.Resize(keys.Len())

	kitr := arrowutil.IterateFloats(buf.keys)
	vitr := arrowutil.IterateStrings(buf.values)
	for i := 0; kitr.Next() && vitr.Next(); {
		for ; i < keys.Len(); i++ {
			if kitr.Value() == keys.Value(i) {
				if vitr.IsValid() {
					b.Append(vitr.Value())
				} else {
					b.AppendNull()
				}
				i++
				break
			}
			b.AppendNull()
		}
	}
	for i := b.Len(); i < keys.Len(); i++ {
		b.AppendNull()
	}
	return b.NewArray()
}

func (gr *pivotTableGroup) buildTimeColumnFromFloats(keys *array.Float, buf *pivotTableBuffer, mem memory.Allocator) array.Array {
	b := arrowutil.NewIntBuilder(mem)
	b.Resize(keys.Len())

	kitr := arrowutil.IterateFloats(buf.keys)
	vitr := arrowutil.IterateInts(buf.values)
	for i := 0; kitr.Next() && vitr.Next(); {
		for ; i < keys.Len(); i++ {
			if kitr.Value() == keys.Value(i) {
				if vitr.IsValid() {
					b.Append(vitr.Value())
				} else {
					b.AppendNull()
				}
				i++
				break
			}
			b.AppendNull()
		}
	}
	for i := b.Len(); i < keys.Len(); i++ {
		b.AppendNull()
	}
	return b.NewArray()
}

func (gr *pivotTableGroup) buildIntColumnFromBooleans(keys *array.Boolean, buf *pivotTableBuffer, mem memory.Allocator) array.Array {
	b := arrowutil.NewIntBuilder(mem)
	b.Resize(keys.Len())

	kitr := arrowutil.IterateBooleans(buf.keys)
	vitr := arrowutil.IterateInts(buf.values)
	for i := 0; kitr.Next() && vitr.Next(); {
		for ; i < keys.Len(); i++ {
			if kitr.Value() == keys.Value(i) {
				if vitr.IsValid() {
					b.Append(vitr.Value())
				} else {
					b.AppendNull()
				}
				i++
				break
			}
			b.AppendNull()
		}
	}
	for i := b.Len(); i < keys.Len(); i++ {
		b.AppendNull()
	}
	return b.NewArray()
}

func (gr *pivotTableGroup) buildUintColumnFromBooleans(keys *array.Boolean, buf *pivotTableBuffer, mem memory.Allocator) array.Array {
	b := arrowutil.NewUintBuilder(mem)
	b.Resize(keys.Len())

	kitr := arrowutil.IterateBooleans(buf.keys)
	vitr := arrowutil.IterateUints(buf.values)
	for i := 0; kitr.Next() && vitr.Next(); {
		for ; i < keys.Len(); i++ {
			if kitr.Value() == keys.Value(i) {
				if vitr.IsValid() {
					b.Append(vitr.Value())
				} else {
					b.AppendNull()
				}
				i++
				break
			}
			b.AppendNull()
		}
	}
	for i := b.Len(); i < keys.Len(); i++ {
		b.AppendNull()
	}
	return b.NewArray()
}

func (gr *pivotTableGroup) buildFloatColumnFromBooleans(keys *array.Boolean, buf *pivotTableBuffer, mem memory.Allocator) array.Array {
	b := arrowutil.NewFloatBuilder(mem)
	b.Resize(keys.Len())

	kitr := arrowutil.IterateBooleans(buf.keys)
	vitr := arrowutil.IterateFloats(buf.values)
	for i := 0; kitr.Next() && vitr.Next(); {
		for ; i < keys.Len(); i++ {
			if kitr.Value() == keys.Value(i) {
				if vitr.IsValid() {
					b.Append(vitr.Value())
				} else {
					b.AppendNull()
				}
				i++
				break
			}
			b.AppendNull()
		}
	}
	for i := b.Len(); i < keys.Len(); i++ {
		b.AppendNull()
	}
	return b.NewArray()
}

func (gr *pivotTableGroup) buildBooleanColumnFromBooleans(keys *array.Boolean, buf *pivotTableBuffer, mem memory.Allocator) array.Array {
	b := arrowutil.NewBooleanBuilder(mem)
	b.Resize(keys.Len())

	kitr := arrowutil.IterateBooleans(buf.keys)
	vitr := arrowutil.IterateBooleans(buf.values)
	for i := 0; kitr.Next() && vitr.Next(); {
		for ; i < keys.Len(); i++ {
			if kitr.Value() == keys.Value(i) {
				if vitr.IsValid() {
					b.Append(vitr.Value())
				} else {
					b.AppendNull()
				}
				i++
				break
			}
			b.AppendNull()
		}
	}
	for i := b.Len(); i < keys.Len(); i++ {
		b.AppendNull()
	}
	return b.NewArray()
}

func (gr *pivotTableGroup) buildStringColumnFromBooleans(keys *array.Boolean, buf *pivotTableBuffer, mem memory.Allocator) array.Array {
	b := arrowutil.NewStringBuilder(mem)
	b.Resize(keys.Len())

	kitr := arrowutil.IterateBooleans(buf.keys)
	vitr := arrowutil.IterateStrings(buf.values)
	for i := 0; kitr.Next() && vitr.Next(); {
		for ; i < keys.Len(); i++ {
			if kitr.Value() == keys.Value(i) {
				if vitr.IsValid() {
					b.Append(vitr.Value())
				} else {
					b.AppendNull()
				}
				i++
				break
			}
			b.AppendNull()
		}
	}
	for i := b.Len(); i < keys.Len(); i++ {
		b.AppendNull()
	}
	return b.NewArray()
}

func (gr *pivotTableGroup) buildTimeColumnFromBooleans(keys *array.Boolean, buf *pivotTableBuffer, mem memory.Allocator) array.Array {
	b := arrowutil.NewIntBuilder(mem)
	b.Resize(keys.Len())

	kitr := arrowutil.IterateBooleans(buf.keys)
	vitr := arrowutil.IterateInts(buf.values)
	for i := 0; kitr.Next() && vitr.Next(); {
		for ; i < keys.Len(); i++ {
			if kitr.Value() == keys.Value(i) {
				if vitr.IsValid() {
					b.Append(vitr.Value())
				} else {
					b.AppendNull()
				}
				i++
				break
			}
			b.AppendNull()
		}
	}
	for i := b.Len(); i < keys.Len(); i++ {
		b.AppendNull()
	}
	return b.NewArray()
}

func (gr *pivotTableGroup) mergeStringKeys(mem memory.Allocator) array.Array {
	buffers := make([][]array.Array, 0, len(gr.buffers))
	for _, buf := range gr.buffers {
		buffers = append(buffers, buf.keys)
	}

	count := 0
	gr.forEachString(buffers, func(v string) {
		count++
	})

	b := arrowutil.NewStringBuilder(mem)
	b.Resize(count)
	gr.forEachString(buffers, b.Append)
	return b.NewArray()
}

func (gr *pivotTableGroup) forEachString(buffers [][]array.Array, fn func(v string)) {
	iterators := make([]*arrowutil.StringIterator, 0, len(buffers))
	for _, vs := range buffers {
		itr := arrowutil.IterateStrings(vs)
		if !itr.Next() {
			continue
		}
		iterators = append(iterators, &itr)
	}

	// Count the number of common keys.
	for len(iterators) > 0 {
		next := iterators[0].Value()
		for _, itr := range iterators[1:] {
			if v := itr.Value(); v < next {
				next = v
			}
		}

		// This counts as a row.
		fn(next)

		// Advance any iterators to the next non-null value
		// that match the next value.
		for i := 0; i < len(iterators); {
			itr := iterators[i]
			if itr.Value() != next {
				i++
				continue
			}

			// Advance to the next non-null value.
			for {
				if !itr.Next() {
					// Remove this iterator from the list.
					copy(iterators[i:], iterators[i+1:])
					iterators = iterators[:len(iterators)-1]
					break
				}

				if itr.IsValid() && itr.Value() != next {
					// The next value is valid so advance
					// to the next iterator.
					i++
					break
				}
			}
		}
	}
}

func (gr *pivotTableGroup) buildColumnFrom(keys *array.String, buf *pivotTableBuffer, mem memory.Allocator) array.Array {

	switch buf.valueType {

	case flux.TInt:
		return gr.buildIntColumnFrom(keys, buf, mem)

	case flux.TUInt:
		return gr.buildUintColumnFrom(keys, buf, mem)

	case flux.TFloat:
		return gr.buildFloatColumnFrom(keys, buf, mem)

	case flux.TBool:
		return gr.buildBooleanColumnFrom(keys, buf, mem)

	case flux.TString:
		return gr.buildStringColumnFrom(keys, buf, mem)

	case flux.TTime:
		return gr.buildTimeColumnFrom(keys, buf, mem)

	default:
		panic("unimplemented")
	}

}

func (gr *pivotTableGroup) buildIntColumnFrom(keys *array.String, buf *pivotTableBuffer, mem memory.Allocator) array.Array {
	b := arrowutil.NewIntBuilder(mem)
	b.Resize(keys.Len())

	kitr := arrowutil.IterateStrings(buf.keys)
	vitr := arrowutil.IterateInts(buf.values)
	for i := 0; kitr.Next() && vitr.Next(); {
		for ; i < keys.Len(); i++ {
			if kitr.Value() == keys.Value(i) {
				if vitr.IsValid() {
					b.Append(vitr.Value())
				} else {
					b.AppendNull()
				}
				i++
				break
			}
			b.AppendNull()
		}
	}
	for i := b.Len(); i < keys.Len(); i++ {
		b.AppendNull()
	}
	return b.NewArray()
}

func (gr *pivotTableGroup) buildUintColumnFrom(keys *array.String, buf *pivotTableBuffer, mem memory.Allocator) array.Array {
	b := arrowutil.NewUintBuilder(mem)
	b.Resize(keys.Len())

	kitr := arrowutil.IterateStrings(buf.keys)
	vitr := arrowutil.IterateUints(buf.values)
	for i := 0; kitr.Next() && vitr.Next(); {
		for ; i < keys.Len(); i++ {
			if kitr.Value() == keys.Value(i) {
				if vitr.IsValid() {
					b.Append(vitr.Value())
				} else {
					b.AppendNull()
				}
				i++
				break
			}
			b.AppendNull()
		}
	}
	for i := b.Len(); i < keys.Len(); i++ {
		b.AppendNull()
	}
	return b.NewArray()
}

func (gr *pivotTableGroup) buildFloatColumnFrom(keys *array.String, buf *pivotTableBuffer, mem memory.Allocator) array.Array {
	b := arrowutil.NewFloatBuilder(mem)
	b.Resize(keys.Len())

	kitr := arrowutil.IterateStrings(buf.keys)
	vitr := arrowutil.IterateFloats(buf.values)
	for i := 0; kitr.Next() && vitr.Next(); {
		for ; i < keys.Len(); i++ {
			if kitr.Value() == keys.Value(i) {
				if vitr.IsValid() {
					b.Append(vitr.Value())
				} else {
					b.AppendNull()
				}
				i++
				break
			}
			b.AppendNull()
		}
	}
	for i := b.Len(); i < keys.Len(); i++ {
		b.AppendNull()
	}
	return b.NewArray()
}

func (gr *pivotTableGroup) buildBooleanColumnFrom(keys *array.String, buf *pivotTableBuffer, mem memory.Allocator) array.Array {
	b := arrowutil.NewBooleanBuilder(mem)
	b.Resize(keys.Len())

	kitr := arrowutil.IterateStrings(buf.keys)
	vitr := arrowutil.IterateBooleans(buf.values)
	for i := 0; kitr.Next() && vitr.Next(); {
		for ; i < keys.Len(); i++ {
			if kitr.Value() == keys.Value(i) {
				if vitr.IsValid() {
					b.Append(vitr.Value())
				} else {
					b.AppendNull()
				}
				i++
				break
			}
			b.AppendNull()
		}
	}
	for i := b.Len(); i < keys.Len(); i++ {
		b.AppendNull()
	}
	return b.NewArray()
}

func (gr *pivotTableGroup) buildStringColumnFrom(keys *array.String, buf *pivotTableBuffer, mem memory.Allocator) array.Array {
	b := arrowutil.NewStringBuilder(mem)
	b.Resize(keys.Len())

	kitr := arrowutil.IterateStrings(buf.keys)
	vitr := arrowutil.IterateStrings(buf.values)
	for i := 0; kitr.Next() && vitr.Next(); {
		for ; i < keys.Len(); i++ {
			if kitr.Value() == keys.Value(i) {
				if vitr.IsValid() {
					b.Append(vitr.Value())
				} else {
					b.AppendNull()
				}
				i++
				break
			}
			b.AppendNull()
		}
	}
	for i := b.Len(); i < keys.Len(); i++ {
		b.AppendNull()
	}
	return b.NewArray()
}

func (gr *pivotTableGroup) buildTimeColumnFrom(keys *array.String, buf *pivotTableBuffer, mem memory.Allocator) array.Array {
	b := arrowutil.NewIntBuilder(mem)
	b.Resize(keys.Len())

	kitr := arrowutil.IterateStrings(buf.keys)
	vitr := arrowutil.IterateInts(buf.values)
	for i := 0; kitr.Next() && vitr.Next(); {
		for ; i < keys.Len(); i++ {
			if kitr.Value() == keys.Value(i) {
				if vitr.IsValid() {
					b.Append(vitr.Value())
				} else {
					b.AppendNull()
				}
				i++
				break
			}
			b.AppendNull()
		}
	}
	for i := b.Len(); i < keys.Len(); i++ {
		b.AppendNull()
	}
	return b.NewArray()
}

func (gr *pivotTableGroup) mergeTimeKeys(mem memory.Allocator) array.Array {
	buffers := make([][]array.Array, 0, len(gr.buffers))
	for _, buf := range gr.buffers {
		buffers = append(buffers, buf.keys)
	}

	count := 0
	gr.forEachTime(buffers, func(v int64) {
		count++
	})

	b := arrowutil.NewIntBuilder(mem)
	b.Resize(count)
	gr.forEachTime(buffers, b.Append)
	return b.NewArray()
}

func (gr *pivotTableGroup) forEachTime(buffers [][]array.Array, fn func(v int64)) {
	iterators := make([]*arrowutil.IntIterator, 0, len(buffers))
	for _, vs := range buffers {
		itr := arrowutil.IterateInts(vs)
		if !itr.Next() {
			continue
		}
		iterators = append(iterators, &itr)
	}

	// Count the number of common keys.
	for len(iterators) > 0 {
		next := iterators[0].Value()
		for _, itr := range iterators[1:] {
			if v := itr.Value(); v < next {
				next = v
			}
		}

		// This counts as a row.
		fn(next)

		// Advance any iterators to the next non-null value
		// that match the next value.
		for i := 0; i < len(iterators); {
			itr := iterators[i]
			if itr.Value() != next {
				i++
				continue
			}

			// Advance to the next non-null value.
			for {
				if !itr.Next() {
					// Remove this iterator from the list.
					copy(iterators[i:], iterators[i+1:])
					iterators = iterators[:len(iterators)-1]
					break
				}

				if itr.IsValid() && itr.Value() != next {
					// The next value is valid so advance
					// to the next iterator.
					i++
					break
				}
			}
		}
	}
}

func (gr *pivotTableGroup) buildColumnFromTimes(keys *array.Int, buf *pivotTableBuffer, mem memory.Allocator) array.Array {

	switch buf.valueType {

	case flux.TInt:
		return gr.buildIntColumnFromTimes(keys, buf, mem)

	case flux.TUInt:
		return gr.buildUintColumnFromTimes(keys, buf, mem)

	case flux.TFloat:
		return gr.buildFloatColumnFromTimes(keys, buf, mem)

	case flux.TBool:
		return gr.buildBooleanColumnFromTimes(keys, buf, mem)

	case flux.TString:
		return gr.buildStringColumnFromTimes(keys, buf, mem)

	case flux.TTime:
		return gr.buildTimeColumnFromTimes(keys, buf, mem)

	default:
		panic("unimplemented")
	}

}

func (gr *pivotTableGroup) buildIntColumnFromTimes(keys *array.Int, buf *pivotTableBuffer, mem memory.Allocator) array.Array {
	b := arrowutil.NewIntBuilder(mem)
	b.Resize(keys.Len())

	kitr := arrowutil.IterateInts(buf.keys)
	vitr := arrowutil.IterateInts(buf.values)
	for i := 0; kitr.Next() && vitr.Next(); {
		for ; i < keys.Len(); i++ {
			if kitr.Value() == keys.Value(i) {
				if vitr.IsValid() {
					b.Append(vitr.Value())
				} else {
					b.AppendNull()
				}
				i++
				break
			}
			b.AppendNull()
		}
	}
	for i := b.Len(); i < keys.Len(); i++ {
		b.AppendNull()
	}
	return b.NewArray()
}

func (gr *pivotTableGroup) buildUintColumnFromTimes(keys *array.Int, buf *pivotTableBuffer, mem memory.Allocator) array.Array {
	b := arrowutil.NewUintBuilder(mem)
	b.Resize(keys.Len())

	kitr := arrowutil.IterateInts(buf.keys)
	vitr := arrowutil.IterateUints(buf.values)
	for i := 0; kitr.Next() && vitr.Next(); {
		for ; i < keys.Len(); i++ {
			if kitr.Value() == keys.Value(i) {
				if vitr.IsValid() {
					b.Append(vitr.Value())
				} else {
					b.AppendNull()
				}
				i++
				break
			}
			b.AppendNull()
		}
	}
	for i := b.Len(); i < keys.Len(); i++ {
		b.AppendNull()
	}
	return b.NewArray()
}

func (gr *pivotTableGroup) buildFloatColumnFromTimes(keys *array.Int, buf *pivotTableBuffer, mem memory.Allocator) array.Array {
	b := arrowutil.NewFloatBuilder(mem)
	b.Resize(keys.Len())

	kitr := arrowutil.IterateInts(buf.keys)
	vitr := arrowutil.IterateFloats(buf.values)
	for i := 0; kitr.Next() && vitr.Next(); {
		for ; i < keys.Len(); i++ {
			if kitr.Value() == keys.Value(i) {
				if vitr.IsValid() {
					b.Append(vitr.Value())
				} else {
					b.AppendNull()
				}
				i++
				break
			}
			b.AppendNull()
		}
	}
	for i := b.Len(); i < keys.Len(); i++ {
		b.AppendNull()
	}
	return b.NewArray()
}

func (gr *pivotTableGroup) buildBooleanColumnFromTimes(keys *array.Int, buf *pivotTableBuffer, mem memory.Allocator) array.Array {
	b := arrowutil.NewBooleanBuilder(mem)
	b.Resize(keys.Len())

	kitr := arrowutil.IterateInts(buf.keys)
	vitr := arrowutil.IterateBooleans(buf.values)
	for i := 0; kitr.Next() && vitr.Next(); {
		for ; i < keys.Len(); i++ {
			if kitr.Value() == keys.Value(i) {
				if vitr.IsValid() {
					b.Append(vitr.Value())
				} else {
					b.AppendNull()
				}
				i++
				break
			}
			b.AppendNull()
		}
	}
	for i := b.Len(); i < keys.Len(); i++ {
		b.AppendNull()
	}
	return b.NewArray()
}

func (gr *pivotTableGroup) buildStringColumnFromTimes(keys *array.Int, buf *pivotTableBuffer, mem memory.Allocator) array.Array {
	b := arrowutil.NewStringBuilder(mem)
	b.Resize(keys.Len())

	kitr := arrowutil.IterateInts(buf.keys)
	vitr := arrowutil.IterateStrings(buf.values)
	for i := 0; kitr.Next() && vitr.Next(); {
		for ; i < keys.Len(); i++ {
			if kitr.Value() == keys.Value(i) {
				if vitr.IsValid() {
					b.Append(vitr.Value())
				} else {
					b.AppendNull()
				}
				i++
				break
			}
			b.AppendNull()
		}
	}
	for i := b.Len(); i < keys.Len(); i++ {
		b.AppendNull()
	}
	return b.NewArray()
}

func (gr *pivotTableGroup) buildTimeColumnFromTimes(keys *array.Int, buf *pivotTableBuffer, mem memory.Allocator) array.Array {
	b := arrowutil.NewIntBuilder(mem)
	b.Resize(keys.Len())

	kitr := arrowutil.IterateInts(buf.keys)
	vitr := arrowutil.IterateInts(buf.values)
	for i := 0; kitr.Next() && vitr.Next(); {
		for ; i < keys.Len(); i++ {
			if kitr.Value() == keys.Value(i) {
				if vitr.IsValid() {
					b.Append(vitr.Value())
				} else {
					b.AppendNull()
				}
				i++
				break
			}
			b.AppendNull()
		}
	}
	for i := b.Len(); i < keys.Len(); i++ {
		b.AppendNull()
	}
	return b.NewArray()
}
