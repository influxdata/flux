package compiler

import (
	"context"
	"math"
	"regexp"
	"runtime/debug"
	"strconv"

	"github.com/benbjohnson/immutable"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

type Value struct {
	// t holds the type nature of this value.
	// This determines how the value is read.
	t semantic.Nature
	// v holds the byte representation of this value
	// if it has one.
	v uint64
	// data holds any allocated memory for more complex
	// types such as containers.
	data interface{}
}

func (v Value) Nature() semantic.Nature {
	return v.t
}

func (v Value) IsValid() bool {
	return !v.IsNull()
}

func (v Value) IsNull() bool {
	return v.t == semantic.Invalid
}

func (v Value) Str() string {
	CheckKind(v.t, semantic.String)
	return v.data.(string)
}

func (v Value) Bytes() []byte {
	CheckKind(v.t, semantic.Bytes)
	return v.data.([]byte)
}

func (v Value) Int() int64 {
	CheckKind(v.t, semantic.Int)
	return int64(v.v)
}

func (v Value) Uint() uint64 {
	CheckKind(v.t, semantic.UInt)
	return v.v
}

func (v Value) Float() float64 {
	CheckKind(v.t, semantic.Float)
	return math.Float64frombits(v.v)
}

func (v Value) Bool() bool {
	CheckKind(v.t, semantic.Bool)
	return v.v != 0
}

func (v Value) Time() values.Time {
	CheckKind(v.t, semantic.Time)
	return values.Time(v.v)
}

func (v Value) Duration() values.Duration {
	CheckKind(v.t, semantic.Duration)
	return v.data.(values.Duration)
}

func (v Value) Regexp() *regexp.Regexp {
	CheckKind(v.t, semantic.Regexp)
	return v.data.(*regexp.Regexp)
}

func (v Value) Index(i int) Value {
	CheckKind(v.t, semantic.Array)
	return v.data.(*array).Index(i)
}

func (v Value) Len() int {
	CheckKind(v.t, semantic.Array)
	return v.data.(*array).Len()
}

func (v Value) Set(i int, val Value) {
	CheckKind(v.t, semantic.Object)
	v.data.(*record).Set(i, val)
}

func (v Value) Get(i int) Value {
	CheckKind(v.t, semantic.Object)
	return v.data.(*record).Get(i)
}

func (v Value) NumFields() int {
	CheckKind(v.t, semantic.Object)
	return v.data.(*record).NumFields()
}

func (v Value) FieldByName(name string) int {
	CheckKind(v.t, semantic.Object)
	return v.data.(*record).FieldByName(name)
}

func (v Value) DictGet(key Value, def Value) Value {
	CheckKind(v.t, semantic.Dictionary)
	return v.data.(*dict).Get(key, def)
}

func (v Value) DictInsert(key Value, val Value) (Value, error) {
	CheckKind(v.t, semantic.Dictionary)
	return v.data.(*dict).Insert(key, val)
}

func (v Value) DictRemove(key Value) Value {
	CheckKind(v.t, semantic.Dictionary)
	return v.data.(*dict).Remove(key)
}

func (v Value) Call(ctx context.Context, args []MaybeValue) (Value, error) {
	CheckKind(v.t, semantic.Function)
	return v.data.(Function).Call(ctx, args)
}

func UnexpectedKind(got, exp semantic.Nature) error {
	return errors.Newf(codes.Internal, "unexpected kind: got %q expected %q, trace: %s", got, exp, string(debug.Stack()))
}

// CheckKind panics if got != exp.
func CheckKind(got, exp semantic.Nature) {
	if got != exp {
		panic(UnexpectedKind(got, exp))
	}
}

func NewString(v string) Value {
	return Value{
		t:    semantic.String,
		data: v,
	}
}

func NewBytes(v []byte) Value {
	return Value{
		t:    semantic.Bytes,
		data: v,
	}
}

func NewInt(v int64) Value {
	return Value{
		t: semantic.Int,
		v: uint64(v),
	}
}

func NewUint(v uint64) Value {
	return Value{
		t: semantic.UInt,
		v: v,
	}
}

func NewFloat(v float64) Value {
	return Value{
		t: semantic.Float,
		v: math.Float64bits(v),
	}
}

func NewBool(v bool) Value {
	return Value{
		t: semantic.Bool,
		v: boolbit(v),
	}
}

func boolbit(v bool) uint64 {
	if v {
		return 1
	} else {
		return 0
	}
}

func NewTime(v values.Time) Value {
	return Value{
		t: semantic.Time,
		v: uint64(v),
	}
}

func NewDuration(v values.Duration) Value {
	return Value{
		t:    semantic.Duration,
		data: v,
	}
}

func NewRegexp(v *regexp.Regexp) Value {
	return Value{
		t:    semantic.Regexp,
		data: v,
	}
}

func NewArray(typ semantic.MonoType, values []Value) Value {
	return Value{
		t: semantic.Array,
		data: &array{
			typ:    typ,
			values: values,
		},
	}
}

type array struct {
	values []Value
	typ    semantic.MonoType
}

func (a *array) Index(i int) Value {
	return a.values[i]
}

func (a *array) Len() int {
	return len(a.values)
}

func NewRecord(typ semantic.MonoType) Value {
	n, err := typ.NumProperties()
	if err != nil {
		panic(err)
	}
	labels := make([]string, n)
	for i := 0; i < len(labels); i++ {
		rp, err := typ.RecordProperty(i)
		if err != nil {
			panic(err)
		}
		labels[i] = rp.Name()
	}
	return Value{
		t: semantic.Object,
		data: &record{
			labels: labels,
			values: make([]Value, n),
			typ:    typ,
		},
	}
}

type record struct {
	labels []string
	values []Value
	typ    semantic.MonoType
}

func (r *record) Set(i int, v Value) {
	r.values[i] = v
}

func (r *record) Get(i int) Value {
	return r.values[i]
}

func (r *record) NumFields() int {
	return len(r.labels)
}

func (r *record) FieldByName(name string) int {
	for i, label := range r.labels {
		if label == name {
			return i
		}
	}
	return -1
}

func NewDict(typ semantic.MonoType) Value {
	return Value{
		t: semantic.Dictionary,
		data: &dict{
			data: immutable.NewSortedMap(
				dictComparer(typ),
			),
			typ: typ,
		},
	}
}

type dict struct {
	data *immutable.SortedMap
	typ  semantic.MonoType
}

func (d *dict) Get(key, def Value) Value {
	if !key.IsNull() {
		v, ok := d.data.Get(key)
		if ok {
			return v.(Value)
		}
	}
	return def
}

func (d *dict) Insert(key, value Value) (Value, error) {
	if key.IsNull() {
		return Value{}, errors.New(codes.Invalid, "null value cannot be used as a dictionary key")
	}
	data := d.data.Set(key, value)
	return Value{
		t: semantic.Dictionary,
		data: &dict{
			data: data,
			typ:  d.typ,
		},
	}, nil
}

func (d *dict) Remove(key Value) Value {
	if key.IsNull() {
		return Value{t: semantic.Dictionary, data: d}
	}
	data := d.data.Delete(key)
	return Value{
		t: semantic.Dictionary,
		data: &dict{
			data: data,
			typ:  d.typ,
		},
	}
}

func dictComparer(dictType semantic.MonoType) immutable.Comparer {
	if dictType.Nature() != semantic.Dictionary {
		panic(UnexpectedKind(dictType.Nature(), semantic.Dictionary))
	}
	keyType, err := dictType.KeyType()
	if err != nil {
		panic(err)
	}
	switch n := keyType.Nature(); n {
	case semantic.Int:
		return intComparer{}
	case semantic.UInt:
		return uintComparer{}
	case semantic.Float:
		return floatComparer{}
	case semantic.String:
		return stringComparer{}
	case semantic.Time:
		return timeComparer{}
	default:
		panic(errors.Newf(codes.Internal, "invalid key nature: %s", n))
	}
}

type (
	intComparer    struct{}
	uintComparer   struct{}
	floatComparer  struct{}
	stringComparer struct{}
	timeComparer   struct{}
)

func (c intComparer) Compare(a, b interface{}) int {
	if i, j := a.(Value).Int(), b.(Value).Int(); i < j {
		return -1
	} else if i > j {
		return 1
	}
	return 0
}

func (c uintComparer) Compare(a, b interface{}) int {
	if i, j := a.(Value).Uint(), b.(Value).Uint(); i < j {
		return -1
	} else if i > j {
		return 1
	}
	return 0
}

func (c floatComparer) Compare(a, b interface{}) int {
	if i, j := a.(Value).Float(), b.(Value).Float(); i < j {
		return -1
	} else if i > j {
		return 1
	}
	return 0
}

func (c stringComparer) Compare(a, b interface{}) int {
	if i, j := a.(Value).Str(), b.(Value).Str(); i < j {
		return -1
	} else if i > j {
		return 1
	}
	return 0
}

func (c timeComparer) Compare(a, b interface{}) int {
	if i, j := a.(Value).Time(), b.(Value).Time(); i < j {
		return -1
	} else if i > j {
		return 1
	}
	return 0
}

// DictionaryBuilder can be used to construct a Dictionary
// with in-place memory instead of successive Insert calls
// that create new Dictionary values.
type DictionaryBuilder struct {
	t semantic.MonoType
	b *immutable.SortedMapBuilder
}

// NewDictBuilder will create a new DictionaryBuilder for the given
// key type.
func NewDictBuilder(typ semantic.MonoType) DictionaryBuilder {
	builder := immutable.NewSortedMapBuilder(dictComparer(typ))
	return DictionaryBuilder{t: typ, b: builder}
}

// Dict will construct a new Dictionary using the inserted values.
func (d *DictionaryBuilder) Dict() Value {
	return Value{
		t: semantic.Dictionary,
		data: &dict{
			data: d.b.Map(),
			typ:  d.t,
		},
	}
}

// Get will retrieve a Value if it is present.
func (d *DictionaryBuilder) Get(key Value) (Value, bool) {
	v, ok := d.b.Get(key)
	if !ok {
		return Value{}, false
	}
	return v.(Value), true
}

// Insert will insert a new key/value pair into the Dictionary.
func (d *DictionaryBuilder) Insert(key, value Value) error {
	if key.IsNull() {
		return errors.New(codes.Invalid, "null value cannot be used as a dictionary key")
	}
	d.b.Set(key, value)
	return nil
}

// Remove will remove a key/value pair from the Dictionary.
func (d *DictionaryBuilder) Remove(key Value) {
	if !key.IsNull() {
		d.b.Delete(key)
	}
}

type Function interface {
	Call(ctx context.Context, args []MaybeValue) (Value, error)
}

func NewFunction(fn Function) Value {
	return Value{
		t:    semantic.Function,
		data: fn,
	}
}

// convertToValue will convert a Value from this package
// into a values.Value. This is a temporary conversion method
// while we test the Value struct so that we don't have to modify
// the portions that use the compiler and can rely on the feature
// flag to turn on or off this section.
//
// This method comes with a cost that may undermine the point of
// this value compiler. If this is the case, we should remove this.
// If we switch to using this compiler as the default for all
// runtime compilations, this should be removed during that conversion.
func convertToValue(v Value) values.Value {
	if v.IsNull() {
		return values.Null
	}

	switch v.Nature() {
	case semantic.Bool:
		return values.NewBool(v.Bool())
	case semantic.Int:
		return values.NewInt(v.Int())
	case semantic.UInt:
		return values.NewUInt(v.Uint())
	case semantic.Float:
		return values.NewFloat(v.Float())
	case semantic.Time:
		return values.NewTime(v.Time())
	case semantic.Duration:
		return values.NewDuration(v.Duration())
	case semantic.String:
		return values.NewString(v.Str())
	case semantic.Bytes:
		return values.NewBytes(v.Bytes())
	case semantic.Regexp:
		return values.NewRegexp(v.Regexp())
	case semantic.Array:
		a := v.data.(*array)
		elements := make([]values.Value, len(a.values))
		for i, value := range a.values {
			elements[i] = convertToValue(value)
		}
		return values.NewArrayWithBacking(a.typ, elements)
	case semantic.Object:
		r := v.data.(*record)
		nr := values.NewObject(r.typ)
		for i, label := range r.labels {
			value := convertToValue(r.values[i])
			nr.Set(label, value)
		}
		return nr
	case semantic.Dictionary:
		d := v.data.(*dict)
		dict := values.NewDictBuilder(d.typ)
		for itr := d.data.Iterator(); ; {
			key, value := itr.Next()
			if key == nil {
				break
			}
			_ = dict.Insert(
				convertToValue(key.(Value)),
				convertToValue(value.(Value)),
			)
		}
		return dict.Dict()
	default:
		panic("unreachable")
	}
}

func convertFromValue(v values.Value) Value {
	if v.IsNull() {
		return Value{}
	}

	switch v.Type().Nature() {
	case semantic.Bool:
		return NewBool(v.Bool())
	case semantic.Int:
		return NewInt(v.Int())
	case semantic.UInt:
		return NewUint(v.UInt())
	case semantic.Float:
		return NewFloat(v.Float())
	case semantic.Time:
		return NewTime(v.Time())
	case semantic.Duration:
		return NewDuration(v.Duration())
	case semantic.String:
		return NewString(v.Str())
	case semantic.Bytes:
		return NewBytes(v.Bytes())
	case semantic.Regexp:
		return NewRegexp(v.Regexp())
	case semantic.Array:
		arr := v.Array()
		elements := make([]Value, arr.Len())
		arr.Range(func(i int, v values.Value) {
			elements[i] = convertFromValue(v)
		})
		return NewArray(arr.Type(), elements)
	case semantic.Object:
		i := 0
		record := NewRecord(v.Type())
		v.Object().Range(func(name string, v values.Value) {
			record.Set(i, convertFromValue(v))
			i++
		})
		return record
	case semantic.Dictionary:
		dict := NewDictBuilder(v.Type())
		v.Dict().Range(func(key, value values.Value) {
			k := convertFromValue(key)
			v := convertFromValue(value)
			_ = dict.Insert(k, v)
		})
		return dict.Dict()
	default:
		panic("unreachable")
	}
}

func stringify(v Value) (Value, error) {
	switch v.t {
	case semantic.Bool:
		return NewString(strconv.FormatBool(v.Bool())), nil
	case semantic.Int:
		return NewString(strconv.FormatInt(v.Int(), 10)), nil
	case semantic.UInt:
		return NewString(strconv.FormatUint(v.Uint(), 10)), nil
	case semantic.Float:
		return NewString(strconv.FormatFloat(v.Float(), 'f', -1, 64)), nil
	case semantic.Time:
		return NewString(v.Time().String()), nil
	case semantic.Duration:
		return NewString(v.Duration().String()), nil
	case semantic.String:
		return v, nil
	}
	return Value{}, errors.Newf(codes.Invalid, "invalid interpolation type")
}

type MaybeValue struct {
	Value Value
	Valid bool
}

func SomeValue(v Value) MaybeValue {
	return MaybeValue{
		Value: v,
		Valid: true,
	}
}

func (v Value) Equal(other Value) bool {
	if v.t != other.t {
		return false
	}

	if v.IsNull() && other.IsNull() {
		return true
	} else if v.IsNull() || other.IsNull() {
		return false
	}

	switch v.t {
	case semantic.String:
		return v.Str() == other.Str()
	case semantic.Int:
		return v.Int() == other.Int()
	case semantic.UInt:
		return v.Uint() == other.Uint()
	case semantic.Float:
		return v.Float() == other.Float()
	case semantic.Bool:
		return v.Bool() == other.Bool()
	case semantic.Array:
		left, right := v.data.(*array), other.data.(*array)
		if left.Len() != right.Len() {
			return false
		}

		for i, n := 0, left.Len(); i < n; i++ {
			if !left.values[i].Equal(right.values[i]) {
				return false
			}
		}
		return true
	case semantic.Object:
		left, right := v.data.(*record), other.data.(*record)
		if len(left.labels) != len(right.labels) {
			return false
		}

	OUTER:
		for i, label1 := range left.labels {
			lv := left.values[i]
			for j, label2 := range right.labels {
				if label1 == label2 {
					rv := right.values[j]
					if !lv.Equal(rv) {
						return false
					}
					continue OUTER
				}
			}

			// Label not found on the right side.
			return false
		}
		return true
	default:
		panic("implement me")
	}
}
