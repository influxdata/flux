// Generated by tmpl
// https://github.com/benbjohnson/tmpl
//
// DO NOT EDIT!
// Source: builder.gen.go.tmpl

package arrowutil

import (
	"github.com/apache/arrow-go/v18/arrow/memory"
	"github.com/influxdata/flux/array"
)

func NewIntBuilder(mem memory.Allocator) *array.IntBuilder {
	return array.NewIntBuilder(mem)
}

func NewUintBuilder(mem memory.Allocator) *array.UintBuilder {
	return array.NewUintBuilder(mem)
}

func NewFloatBuilder(mem memory.Allocator) *array.FloatBuilder {
	return array.NewFloatBuilder(mem)
}

func NewBooleanBuilder(mem memory.Allocator) *array.BooleanBuilder {
	return array.NewBooleanBuilder(mem)
}

func NewStringBuilder(mem memory.Allocator) *array.StringBuilder {
	return array.NewStringBuilder(mem)
}
