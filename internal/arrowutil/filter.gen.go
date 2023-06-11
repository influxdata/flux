// Generated by tmpl
// https://github.com/benbjohnson/tmpl
//
// DO NOT EDIT!
// Source: filter.gen.go.tmpl

package arrowutil

import (
	"fmt"

	"github.com/InfluxCommunity/flux/array"
	"github.com/apache/arrow/go/v7/arrow/bitutil"
	"github.com/apache/arrow/go/v7/arrow/memory"
)

func Filter(arr array.Array, bitset []byte, mem memory.Allocator) array.Array {
	switch arr := arr.(type) {

	case *array.Int:
		return FilterInts(arr, bitset, mem)

	case *array.Uint:
		return FilterUints(arr, bitset, mem)

	case *array.Float:
		return FilterFloats(arr, bitset, mem)

	case *array.Boolean:
		return FilterBooleans(arr, bitset, mem)

	case *array.String:
		return FilterStrings(arr, bitset, mem)

	default:
		panic(fmt.Errorf("unsupported array data type: %s", arr.DataType()))
	}
}

func FilterInts(arr *array.Int, bitset []byte, mem memory.Allocator) *array.Int {
	n := bitutil.CountSetBits(bitset, 0, len(bitset))
	b := NewIntBuilder(mem)
	b.Resize(n)
	for i := 0; i < len(bitset); i++ {
		if bitutil.BitIsSet(bitset, i) {
			if arr.IsValid(i) {
				b.Append(arr.Value(i))
			} else {
				b.AppendNull()
			}
		}
	}
	return b.NewIntArray()
}

func FilterUints(arr *array.Uint, bitset []byte, mem memory.Allocator) *array.Uint {
	n := bitutil.CountSetBits(bitset, 0, len(bitset))
	b := NewUintBuilder(mem)
	b.Resize(n)
	for i := 0; i < len(bitset); i++ {
		if bitutil.BitIsSet(bitset, i) {
			if arr.IsValid(i) {
				b.Append(arr.Value(i))
			} else {
				b.AppendNull()
			}
		}
	}
	return b.NewUintArray()
}

func FilterFloats(arr *array.Float, bitset []byte, mem memory.Allocator) *array.Float {
	n := bitutil.CountSetBits(bitset, 0, len(bitset))
	b := NewFloatBuilder(mem)
	b.Resize(n)
	for i := 0; i < len(bitset); i++ {
		if bitutil.BitIsSet(bitset, i) {
			if arr.IsValid(i) {
				b.Append(arr.Value(i))
			} else {
				b.AppendNull()
			}
		}
	}
	return b.NewFloatArray()
}

func FilterBooleans(arr *array.Boolean, bitset []byte, mem memory.Allocator) *array.Boolean {
	n := bitutil.CountSetBits(bitset, 0, len(bitset))
	b := NewBooleanBuilder(mem)
	b.Resize(n)
	for i := 0; i < len(bitset); i++ {
		if bitutil.BitIsSet(bitset, i) {
			if arr.IsValid(i) {
				b.Append(arr.Value(i))
			} else {
				b.AppendNull()
			}
		}
	}
	return b.NewBooleanArray()
}

func FilterStrings(arr *array.String, bitset []byte, mem memory.Allocator) *array.String {
	n := bitutil.CountSetBits(bitset, 0, len(bitset))
	b := NewStringBuilder(mem)
	b.Resize(n)
	for i := 0; i < len(bitset); i++ {
		if bitutil.BitIsSet(bitset, i) {
			if arr.IsValid(i) {
				b.Append(arr.Value(i))
			} else {
				b.AppendNull()
			}
		}
	}
	return b.NewStringArray()
}
