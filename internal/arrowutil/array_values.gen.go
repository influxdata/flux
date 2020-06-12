// Generated by tmpl
// https://github.com/benbjohnson/tmpl
//
// DO NOT EDIT!
// Source: array_values.gen.go.tmpl

package arrowutil

import (
	"fmt"
	"regexp"

	"github.com/apache/arrow/go/arrow/array"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

func NewArrayValue(arr array.Interface, typ flux.ColType) values.Array {
	switch elemType := flux.SemanticType(typ); elemType {

	case semantic.BasicInt:
		return NewInt64ArrayValue(arr.(*array.Int64))

	case semantic.BasicUint:
		return NewUint64ArrayValue(arr.(*array.Uint64))

	case semantic.BasicFloat:
		return NewFloat64ArrayValue(arr.(*array.Float64))

	case semantic.BasicBool:
		return NewBooleanArrayValue(arr.(*array.Boolean))

	case semantic.BasicString:
		return NewStringArrayValue(arr.(*array.Binary))

	default:
		panic(fmt.Errorf("unsupported column data type: %s", typ))
	}
}

var _ values.Value = Int64ArrayValue{}
var _ values.Array = Int64ArrayValue{}

type Int64ArrayValue struct {
	arr *array.Int64
	typ semantic.MonoType
}

func NewInt64ArrayValue(arr *array.Int64) values.Array {
	return Int64ArrayValue{
		arr: arr,
		typ: semantic.NewArrayType(semantic.BasicInt),
	}
}

func (v Int64ArrayValue) Type() semantic.MonoType { return v.typ }
func (v Int64ArrayValue) IsNull() bool            { return false }
func (v Int64ArrayValue) Str() string             { panic(values.UnexpectedKind(semantic.Array, semantic.String)) }
func (v Int64ArrayValue) Bytes() []byte           { panic(values.UnexpectedKind(semantic.Array, semantic.Bytes)) }
func (v Int64ArrayValue) Int() int64              { panic(values.UnexpectedKind(semantic.Array, semantic.Int)) }
func (v Int64ArrayValue) UInt() uint64            { panic(values.UnexpectedKind(semantic.Array, semantic.UInt)) }
func (v Int64ArrayValue) Float() float64 {
	panic(values.UnexpectedKind(semantic.Array, semantic.Float))
}
func (v Int64ArrayValue) Bool() bool { panic(values.UnexpectedKind(semantic.Array, semantic.Bool)) }
func (v Int64ArrayValue) Time() values.Time {
	panic(values.UnexpectedKind(semantic.Array, semantic.Time))
}
func (v Int64ArrayValue) Duration() values.Duration {
	panic(values.UnexpectedKind(semantic.Array, semantic.Duration))
}
func (v Int64ArrayValue) Regexp() *regexp.Regexp {
	panic(values.UnexpectedKind(semantic.Array, semantic.Regexp))
}
func (v Int64ArrayValue) Array() values.Array { return v }
func (v Int64ArrayValue) Object() values.Object {
	panic(values.UnexpectedKind(semantic.Array, semantic.Object))
}
func (v Int64ArrayValue) Function() values.Function {
	panic(values.UnexpectedKind(semantic.Array, semantic.Function))
}

func (v Int64ArrayValue) Equal(other values.Value) bool {
	if other.Type().Nature() != semantic.Array {
		return false
	} else if v.arr.Len() != other.Array().Len() {
		return false
	}

	otherArray := other.Array()
	for i, n := 0, v.arr.Len(); i < n; i++ {
		if !v.Get(i).Equal(otherArray.Get(i)) {
			return false
		}
	}
	return true
}

func (v Int64ArrayValue) Get(i int) values.Value {
	if v.arr.IsNull(i) {
		return values.Null
	}
	return values.New(v.arr.Value(i))
}

func (v Int64ArrayValue) Set(i int, value values.Value) { panic("cannot set value on immutable array") }
func (v Int64ArrayValue) Append(value values.Value)     { panic("cannot append to immutable array") }

func (v Int64ArrayValue) Len() int { return v.arr.Len() }
func (v Int64ArrayValue) Range(f func(i int, v values.Value)) {
	for i, n := 0, v.arr.Len(); i < n; i++ {
		f(i, v.Get(i))
	}
}

func (v Int64ArrayValue) Sort(f func(i values.Value, j values.Value) bool) {
	panic("cannot sort immutable array")
}

var _ values.Value = Uint64ArrayValue{}
var _ values.Array = Uint64ArrayValue{}

type Uint64ArrayValue struct {
	arr *array.Uint64
	typ semantic.MonoType
}

func NewUint64ArrayValue(arr *array.Uint64) values.Array {
	return Uint64ArrayValue{
		arr: arr,
		typ: semantic.NewArrayType(semantic.BasicUint),
	}
}

func (v Uint64ArrayValue) Type() semantic.MonoType { return v.typ }
func (v Uint64ArrayValue) IsNull() bool            { return false }
func (v Uint64ArrayValue) Str() string             { panic(values.UnexpectedKind(semantic.Array, semantic.String)) }
func (v Uint64ArrayValue) Bytes() []byte {
	panic(values.UnexpectedKind(semantic.Array, semantic.Bytes))
}
func (v Uint64ArrayValue) Int() int64   { panic(values.UnexpectedKind(semantic.Array, semantic.Int)) }
func (v Uint64ArrayValue) UInt() uint64 { panic(values.UnexpectedKind(semantic.Array, semantic.UInt)) }
func (v Uint64ArrayValue) Float() float64 {
	panic(values.UnexpectedKind(semantic.Array, semantic.Float))
}
func (v Uint64ArrayValue) Bool() bool { panic(values.UnexpectedKind(semantic.Array, semantic.Bool)) }
func (v Uint64ArrayValue) Time() values.Time {
	panic(values.UnexpectedKind(semantic.Array, semantic.Time))
}
func (v Uint64ArrayValue) Duration() values.Duration {
	panic(values.UnexpectedKind(semantic.Array, semantic.Duration))
}
func (v Uint64ArrayValue) Regexp() *regexp.Regexp {
	panic(values.UnexpectedKind(semantic.Array, semantic.Regexp))
}
func (v Uint64ArrayValue) Array() values.Array { return v }
func (v Uint64ArrayValue) Object() values.Object {
	panic(values.UnexpectedKind(semantic.Array, semantic.Object))
}
func (v Uint64ArrayValue) Function() values.Function {
	panic(values.UnexpectedKind(semantic.Array, semantic.Function))
}

func (v Uint64ArrayValue) Equal(other values.Value) bool {
	if other.Type().Nature() != semantic.Array {
		return false
	} else if v.arr.Len() != other.Array().Len() {
		return false
	}

	otherArray := other.Array()
	for i, n := 0, v.arr.Len(); i < n; i++ {
		if !v.Get(i).Equal(otherArray.Get(i)) {
			return false
		}
	}
	return true
}

func (v Uint64ArrayValue) Get(i int) values.Value {
	if v.arr.IsNull(i) {
		return values.Null
	}
	return values.New(v.arr.Value(i))
}

func (v Uint64ArrayValue) Set(i int, value values.Value) {
	panic("cannot set value on immutable array")
}
func (v Uint64ArrayValue) Append(value values.Value) { panic("cannot append to immutable array") }

func (v Uint64ArrayValue) Len() int { return v.arr.Len() }
func (v Uint64ArrayValue) Range(f func(i int, v values.Value)) {
	for i, n := 0, v.arr.Len(); i < n; i++ {
		f(i, v.Get(i))
	}
}

func (v Uint64ArrayValue) Sort(f func(i values.Value, j values.Value) bool) {
	panic("cannot sort immutable array")
}

var _ values.Value = Float64ArrayValue{}
var _ values.Array = Float64ArrayValue{}

type Float64ArrayValue struct {
	arr *array.Float64
	typ semantic.MonoType
}

func NewFloat64ArrayValue(arr *array.Float64) values.Array {
	return Float64ArrayValue{
		arr: arr,
		typ: semantic.NewArrayType(semantic.BasicFloat),
	}
}

func (v Float64ArrayValue) Type() semantic.MonoType { return v.typ }
func (v Float64ArrayValue) IsNull() bool            { return false }
func (v Float64ArrayValue) Str() string {
	panic(values.UnexpectedKind(semantic.Array, semantic.String))
}
func (v Float64ArrayValue) Bytes() []byte {
	panic(values.UnexpectedKind(semantic.Array, semantic.Bytes))
}
func (v Float64ArrayValue) Int() int64   { panic(values.UnexpectedKind(semantic.Array, semantic.Int)) }
func (v Float64ArrayValue) UInt() uint64 { panic(values.UnexpectedKind(semantic.Array, semantic.UInt)) }
func (v Float64ArrayValue) Float() float64 {
	panic(values.UnexpectedKind(semantic.Array, semantic.Float))
}
func (v Float64ArrayValue) Bool() bool { panic(values.UnexpectedKind(semantic.Array, semantic.Bool)) }
func (v Float64ArrayValue) Time() values.Time {
	panic(values.UnexpectedKind(semantic.Array, semantic.Time))
}
func (v Float64ArrayValue) Duration() values.Duration {
	panic(values.UnexpectedKind(semantic.Array, semantic.Duration))
}
func (v Float64ArrayValue) Regexp() *regexp.Regexp {
	panic(values.UnexpectedKind(semantic.Array, semantic.Regexp))
}
func (v Float64ArrayValue) Array() values.Array { return v }
func (v Float64ArrayValue) Object() values.Object {
	panic(values.UnexpectedKind(semantic.Array, semantic.Object))
}
func (v Float64ArrayValue) Function() values.Function {
	panic(values.UnexpectedKind(semantic.Array, semantic.Function))
}

func (v Float64ArrayValue) Equal(other values.Value) bool {
	if other.Type().Nature() != semantic.Array {
		return false
	} else if v.arr.Len() != other.Array().Len() {
		return false
	}

	otherArray := other.Array()
	for i, n := 0, v.arr.Len(); i < n; i++ {
		if !v.Get(i).Equal(otherArray.Get(i)) {
			return false
		}
	}
	return true
}

func (v Float64ArrayValue) Get(i int) values.Value {
	if v.arr.IsNull(i) {
		return values.Null
	}
	return values.New(v.arr.Value(i))
}

func (v Float64ArrayValue) Set(i int, value values.Value) {
	panic("cannot set value on immutable array")
}
func (v Float64ArrayValue) Append(value values.Value) { panic("cannot append to immutable array") }

func (v Float64ArrayValue) Len() int { return v.arr.Len() }
func (v Float64ArrayValue) Range(f func(i int, v values.Value)) {
	for i, n := 0, v.arr.Len(); i < n; i++ {
		f(i, v.Get(i))
	}
}

func (v Float64ArrayValue) Sort(f func(i values.Value, j values.Value) bool) {
	panic("cannot sort immutable array")
}

var _ values.Value = BooleanArrayValue{}
var _ values.Array = BooleanArrayValue{}

type BooleanArrayValue struct {
	arr *array.Boolean
	typ semantic.MonoType
}

func NewBooleanArrayValue(arr *array.Boolean) values.Array {
	return BooleanArrayValue{
		arr: arr,
		typ: semantic.NewArrayType(semantic.BasicBool),
	}
}

func (v BooleanArrayValue) Type() semantic.MonoType { return v.typ }
func (v BooleanArrayValue) IsNull() bool            { return false }
func (v BooleanArrayValue) Str() string {
	panic(values.UnexpectedKind(semantic.Array, semantic.String))
}
func (v BooleanArrayValue) Bytes() []byte {
	panic(values.UnexpectedKind(semantic.Array, semantic.Bytes))
}
func (v BooleanArrayValue) Int() int64   { panic(values.UnexpectedKind(semantic.Array, semantic.Int)) }
func (v BooleanArrayValue) UInt() uint64 { panic(values.UnexpectedKind(semantic.Array, semantic.UInt)) }
func (v BooleanArrayValue) Float() float64 {
	panic(values.UnexpectedKind(semantic.Array, semantic.Float))
}
func (v BooleanArrayValue) Bool() bool { panic(values.UnexpectedKind(semantic.Array, semantic.Bool)) }
func (v BooleanArrayValue) Time() values.Time {
	panic(values.UnexpectedKind(semantic.Array, semantic.Time))
}
func (v BooleanArrayValue) Duration() values.Duration {
	panic(values.UnexpectedKind(semantic.Array, semantic.Duration))
}
func (v BooleanArrayValue) Regexp() *regexp.Regexp {
	panic(values.UnexpectedKind(semantic.Array, semantic.Regexp))
}
func (v BooleanArrayValue) Array() values.Array { return v }
func (v BooleanArrayValue) Object() values.Object {
	panic(values.UnexpectedKind(semantic.Array, semantic.Object))
}
func (v BooleanArrayValue) Function() values.Function {
	panic(values.UnexpectedKind(semantic.Array, semantic.Function))
}

func (v BooleanArrayValue) Equal(other values.Value) bool {
	if other.Type().Nature() != semantic.Array {
		return false
	} else if v.arr.Len() != other.Array().Len() {
		return false
	}

	otherArray := other.Array()
	for i, n := 0, v.arr.Len(); i < n; i++ {
		if !v.Get(i).Equal(otherArray.Get(i)) {
			return false
		}
	}
	return true
}

func (v BooleanArrayValue) Get(i int) values.Value {
	if v.arr.IsNull(i) {
		return values.Null
	}
	return values.New(v.arr.Value(i))
}

func (v BooleanArrayValue) Set(i int, value values.Value) {
	panic("cannot set value on immutable array")
}
func (v BooleanArrayValue) Append(value values.Value) { panic("cannot append to immutable array") }

func (v BooleanArrayValue) Len() int { return v.arr.Len() }
func (v BooleanArrayValue) Range(f func(i int, v values.Value)) {
	for i, n := 0, v.arr.Len(); i < n; i++ {
		f(i, v.Get(i))
	}
}

func (v BooleanArrayValue) Sort(f func(i values.Value, j values.Value) bool) {
	panic("cannot sort immutable array")
}

var _ values.Value = StringArrayValue{}
var _ values.Array = StringArrayValue{}

type StringArrayValue struct {
	arr *array.Binary
	typ semantic.MonoType
}

func NewStringArrayValue(arr *array.Binary) values.Array {
	return StringArrayValue{
		arr: arr,
		typ: semantic.NewArrayType(semantic.BasicString),
	}
}

func (v StringArrayValue) Type() semantic.MonoType { return v.typ }
func (v StringArrayValue) IsNull() bool            { return false }
func (v StringArrayValue) Str() string             { panic(values.UnexpectedKind(semantic.Array, semantic.String)) }
func (v StringArrayValue) Bytes() []byte {
	panic(values.UnexpectedKind(semantic.Array, semantic.Bytes))
}
func (v StringArrayValue) Int() int64   { panic(values.UnexpectedKind(semantic.Array, semantic.Int)) }
func (v StringArrayValue) UInt() uint64 { panic(values.UnexpectedKind(semantic.Array, semantic.UInt)) }
func (v StringArrayValue) Float() float64 {
	panic(values.UnexpectedKind(semantic.Array, semantic.Float))
}
func (v StringArrayValue) Bool() bool { panic(values.UnexpectedKind(semantic.Array, semantic.Bool)) }
func (v StringArrayValue) Time() values.Time {
	panic(values.UnexpectedKind(semantic.Array, semantic.Time))
}
func (v StringArrayValue) Duration() values.Duration {
	panic(values.UnexpectedKind(semantic.Array, semantic.Duration))
}
func (v StringArrayValue) Regexp() *regexp.Regexp {
	panic(values.UnexpectedKind(semantic.Array, semantic.Regexp))
}
func (v StringArrayValue) Array() values.Array { return v }
func (v StringArrayValue) Object() values.Object {
	panic(values.UnexpectedKind(semantic.Array, semantic.Object))
}
func (v StringArrayValue) Function() values.Function {
	panic(values.UnexpectedKind(semantic.Array, semantic.Function))
}

func (v StringArrayValue) Equal(other values.Value) bool {
	if other.Type().Nature() != semantic.Array {
		return false
	} else if v.arr.Len() != other.Array().Len() {
		return false
	}

	otherArray := other.Array()
	for i, n := 0, v.arr.Len(); i < n; i++ {
		if !v.Get(i).Equal(otherArray.Get(i)) {
			return false
		}
	}
	return true
}

func (v StringArrayValue) Get(i int) values.Value {
	if v.arr.IsNull(i) {
		return values.Null
	}
	return values.New(v.arr.ValueString(i))
}

func (v StringArrayValue) Set(i int, value values.Value) {
	panic("cannot set value on immutable array")
}
func (v StringArrayValue) Append(value values.Value) { panic("cannot append to immutable array") }

func (v StringArrayValue) Len() int { return v.arr.Len() }
func (v StringArrayValue) Range(f func(i int, v values.Value)) {
	for i, n := 0, v.arr.Len(); i < n; i++ {
		f(i, v.Get(i))
	}
}

func (v StringArrayValue) Sort(f func(i values.Value, j values.Value) bool) {
	panic("cannot sort immutable array")
}
