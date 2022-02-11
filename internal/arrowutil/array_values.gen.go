// Generated by tmpl
// https://github.com/benbjohnson/tmpl
//
// DO NOT EDIT!
// Source: array_values.gen.go.tmpl

package arrowutil

import (
	"fmt"
	"regexp"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

func NewArrayValue(arr array.Interface, typ flux.ColType) values.Array {
	switch elemType := flux.SemanticType(typ); elemType {

	case semantic.BasicInt:
		return NewIntArrayValue(arr.(*array.Int))

	case semantic.BasicUint:
		return NewUintArrayValue(arr.(*array.Uint))

	case semantic.BasicFloat:
		return NewFloatArrayValue(arr.(*array.Float))

	case semantic.BasicBool:
		return NewBooleanArrayValue(arr.(*array.Boolean))

	case semantic.BasicString:
		return NewStringArrayValue(arr.(*array.String))

	default:
		panic(fmt.Errorf("unsupported column data type: %s", typ))
	}
}

var _ values.Value = IntArrayValue{}
var _ values.Array = IntArrayValue{}

type IntArrayValue struct {
	arr *array.Int
	typ semantic.MonoType
}

func NewIntArrayValue(arr *array.Int) values.Array {
	return IntArrayValue{
		arr: arr,
		typ: semantic.NewArrayType(semantic.BasicInt),
	}
}

func (v IntArrayValue) Type() semantic.MonoType { return v.typ }
func (v IntArrayValue) IsNull() bool            { return false }
func (v IntArrayValue) Str() string             { panic(values.UnexpectedKind(semantic.Array, semantic.String)) }
func (v IntArrayValue) Bytes() []byte           { panic(values.UnexpectedKind(semantic.Array, semantic.Bytes)) }
func (v IntArrayValue) Int() int64              { panic(values.UnexpectedKind(semantic.Array, semantic.Int)) }
func (v IntArrayValue) UInt() uint64            { panic(values.UnexpectedKind(semantic.Array, semantic.UInt)) }
func (v IntArrayValue) Float() float64          { panic(values.UnexpectedKind(semantic.Array, semantic.Float)) }
func (v IntArrayValue) Bool() bool              { panic(values.UnexpectedKind(semantic.Array, semantic.Bool)) }
func (v IntArrayValue) Time() values.Time {
	panic(values.UnexpectedKind(semantic.Array, semantic.Time))
}
func (v IntArrayValue) Duration() values.Duration {
	panic(values.UnexpectedKind(semantic.Array, semantic.Duration))
}
func (v IntArrayValue) Regexp() *regexp.Regexp {
	panic(values.UnexpectedKind(semantic.Array, semantic.Regexp))
}
func (v IntArrayValue) Array() values.Array { return v }
func (v IntArrayValue) Object() values.Object {
	panic(values.UnexpectedKind(semantic.Array, semantic.Object))
}
func (v IntArrayValue) Function() values.Function {
	panic(values.UnexpectedKind(semantic.Array, semantic.Function))
}
func (v IntArrayValue) Dict() values.Dictionary {
	panic(values.UnexpectedKind(semantic.Array, semantic.Dictionary))
}
func (v IntArrayValue) Vector() values.Vector {
	panic(values.UnexpectedKind(semantic.Array, semantic.Vector))
}

func (v IntArrayValue) Equal(other values.Value) bool {
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

func (v IntArrayValue) Get(i int) values.Value {
	if v.arr.IsNull(i) {
		return values.Null
	}
	return values.New(v.arr.Value(i))
}

func (v IntArrayValue) Set(i int, value values.Value) { panic("cannot set value on immutable array") }
func (v IntArrayValue) Append(value values.Value)     { panic("cannot append to immutable array") }

func (v IntArrayValue) Len() int { return v.arr.Len() }
func (v IntArrayValue) Range(f func(i int, v values.Value)) {
	for i, n := 0, v.arr.Len(); i < n; i++ {
		f(i, v.Get(i))
	}
}

func (v IntArrayValue) Sort(f func(i values.Value, j values.Value) bool) {
	panic("cannot sort immutable array")
}

func (v IntArrayValue) Retain() {
	v.arr.Retain()
}

func (v IntArrayValue) Release() {
	v.arr.Release()
}

var _ values.Value = UintArrayValue{}
var _ values.Array = UintArrayValue{}

type UintArrayValue struct {
	arr *array.Uint
	typ semantic.MonoType
}

func NewUintArrayValue(arr *array.Uint) values.Array {
	return UintArrayValue{
		arr: arr,
		typ: semantic.NewArrayType(semantic.BasicUint),
	}
}

func (v UintArrayValue) Type() semantic.MonoType { return v.typ }
func (v UintArrayValue) IsNull() bool            { return false }
func (v UintArrayValue) Str() string             { panic(values.UnexpectedKind(semantic.Array, semantic.String)) }
func (v UintArrayValue) Bytes() []byte           { panic(values.UnexpectedKind(semantic.Array, semantic.Bytes)) }
func (v UintArrayValue) Int() int64              { panic(values.UnexpectedKind(semantic.Array, semantic.Int)) }
func (v UintArrayValue) UInt() uint64            { panic(values.UnexpectedKind(semantic.Array, semantic.UInt)) }
func (v UintArrayValue) Float() float64          { panic(values.UnexpectedKind(semantic.Array, semantic.Float)) }
func (v UintArrayValue) Bool() bool              { panic(values.UnexpectedKind(semantic.Array, semantic.Bool)) }
func (v UintArrayValue) Time() values.Time {
	panic(values.UnexpectedKind(semantic.Array, semantic.Time))
}
func (v UintArrayValue) Duration() values.Duration {
	panic(values.UnexpectedKind(semantic.Array, semantic.Duration))
}
func (v UintArrayValue) Regexp() *regexp.Regexp {
	panic(values.UnexpectedKind(semantic.Array, semantic.Regexp))
}
func (v UintArrayValue) Array() values.Array { return v }
func (v UintArrayValue) Object() values.Object {
	panic(values.UnexpectedKind(semantic.Array, semantic.Object))
}
func (v UintArrayValue) Function() values.Function {
	panic(values.UnexpectedKind(semantic.Array, semantic.Function))
}
func (v UintArrayValue) Dict() values.Dictionary {
	panic(values.UnexpectedKind(semantic.Array, semantic.Dictionary))
}
func (v UintArrayValue) Vector() values.Vector {
	panic(values.UnexpectedKind(semantic.Array, semantic.Vector))
}

func (v UintArrayValue) Equal(other values.Value) bool {
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

func (v UintArrayValue) Get(i int) values.Value {
	if v.arr.IsNull(i) {
		return values.Null
	}
	return values.New(v.arr.Value(i))
}

func (v UintArrayValue) Set(i int, value values.Value) { panic("cannot set value on immutable array") }
func (v UintArrayValue) Append(value values.Value)     { panic("cannot append to immutable array") }

func (v UintArrayValue) Len() int { return v.arr.Len() }
func (v UintArrayValue) Range(f func(i int, v values.Value)) {
	for i, n := 0, v.arr.Len(); i < n; i++ {
		f(i, v.Get(i))
	}
}

func (v UintArrayValue) Sort(f func(i values.Value, j values.Value) bool) {
	panic("cannot sort immutable array")
}

func (v UintArrayValue) Retain() {
	v.arr.Retain()
}

func (v UintArrayValue) Release() {
	v.arr.Release()
}

var _ values.Value = FloatArrayValue{}
var _ values.Array = FloatArrayValue{}

type FloatArrayValue struct {
	arr *array.Float
	typ semantic.MonoType
}

func NewFloatArrayValue(arr *array.Float) values.Array {
	return FloatArrayValue{
		arr: arr,
		typ: semantic.NewArrayType(semantic.BasicFloat),
	}
}

func (v FloatArrayValue) Type() semantic.MonoType { return v.typ }
func (v FloatArrayValue) IsNull() bool            { return false }
func (v FloatArrayValue) Str() string             { panic(values.UnexpectedKind(semantic.Array, semantic.String)) }
func (v FloatArrayValue) Bytes() []byte           { panic(values.UnexpectedKind(semantic.Array, semantic.Bytes)) }
func (v FloatArrayValue) Int() int64              { panic(values.UnexpectedKind(semantic.Array, semantic.Int)) }
func (v FloatArrayValue) UInt() uint64            { panic(values.UnexpectedKind(semantic.Array, semantic.UInt)) }
func (v FloatArrayValue) Float() float64 {
	panic(values.UnexpectedKind(semantic.Array, semantic.Float))
}
func (v FloatArrayValue) Bool() bool { panic(values.UnexpectedKind(semantic.Array, semantic.Bool)) }
func (v FloatArrayValue) Time() values.Time {
	panic(values.UnexpectedKind(semantic.Array, semantic.Time))
}
func (v FloatArrayValue) Duration() values.Duration {
	panic(values.UnexpectedKind(semantic.Array, semantic.Duration))
}
func (v FloatArrayValue) Regexp() *regexp.Regexp {
	panic(values.UnexpectedKind(semantic.Array, semantic.Regexp))
}
func (v FloatArrayValue) Array() values.Array { return v }
func (v FloatArrayValue) Object() values.Object {
	panic(values.UnexpectedKind(semantic.Array, semantic.Object))
}
func (v FloatArrayValue) Function() values.Function {
	panic(values.UnexpectedKind(semantic.Array, semantic.Function))
}
func (v FloatArrayValue) Dict() values.Dictionary {
	panic(values.UnexpectedKind(semantic.Array, semantic.Dictionary))
}
func (v FloatArrayValue) Vector() values.Vector {
	panic(values.UnexpectedKind(semantic.Array, semantic.Vector))
}

func (v FloatArrayValue) Equal(other values.Value) bool {
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

func (v FloatArrayValue) Get(i int) values.Value {
	if v.arr.IsNull(i) {
		return values.Null
	}
	return values.New(v.arr.Value(i))
}

func (v FloatArrayValue) Set(i int, value values.Value) { panic("cannot set value on immutable array") }
func (v FloatArrayValue) Append(value values.Value)     { panic("cannot append to immutable array") }

func (v FloatArrayValue) Len() int { return v.arr.Len() }
func (v FloatArrayValue) Range(f func(i int, v values.Value)) {
	for i, n := 0, v.arr.Len(); i < n; i++ {
		f(i, v.Get(i))
	}
}

func (v FloatArrayValue) Sort(f func(i values.Value, j values.Value) bool) {
	panic("cannot sort immutable array")
}

func (v FloatArrayValue) Retain() {
	v.arr.Retain()
}

func (v FloatArrayValue) Release() {
	v.arr.Release()
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
func (v BooleanArrayValue) Dict() values.Dictionary {
	panic(values.UnexpectedKind(semantic.Array, semantic.Dictionary))
}
func (v BooleanArrayValue) Vector() values.Vector {
	panic(values.UnexpectedKind(semantic.Array, semantic.Vector))
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

func (v BooleanArrayValue) Retain() {
	v.arr.Retain()
}

func (v BooleanArrayValue) Release() {
	v.arr.Release()
}

var _ values.Value = StringArrayValue{}
var _ values.Array = StringArrayValue{}

type StringArrayValue struct {
	arr *array.String
	typ semantic.MonoType
}

func NewStringArrayValue(arr *array.String) values.Array {
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
func (v StringArrayValue) Dict() values.Dictionary {
	panic(values.UnexpectedKind(semantic.Array, semantic.Dictionary))
}
func (v StringArrayValue) Vector() values.Vector {
	panic(values.UnexpectedKind(semantic.Array, semantic.Vector))
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
	return values.New(v.arr.Value(i))
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

func (v StringArrayValue) Retain() {
	v.arr.Retain()
}

func (v StringArrayValue) Release() {
	v.arr.Release()
}
