package compiler

import (
	"math"
	"runtime/debug"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/semantic"
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

func (v Value) Set(i int, val Value) {
	CheckKind(v.t, semantic.Object)
	v.data.(*object).Set(i, val)
}

func (v Value) Get(i int) (Value, bool) {
	CheckKind(v.t, semantic.Object)
	return v.data.(*object).Get(i)
}

func (v Value) Range(f func(i int, name string, value Value)) {
	CheckKind(v.t, semantic.Object)
	v.data.(*object).Range(f)
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

func NewInt(v int64) Value {
	return Value{
		t: semantic.Int,
		v: uint64(v),
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

func NewObject(typ semantic.MonoType) Value {
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
		data: &object{
			labels: labels,
			values: make([]Value, n),
			typ:    typ,
		},
	}
}

type object struct {
	labels []string
	values []Value
	typ    semantic.MonoType
}

func (o *object) Set(i int, v Value) {
	o.values[i] = v
}

func (o *object) Get(i int) (Value, bool) {
	return o.values[i], true
}

func (o *object) Range(f func(i int, name string, v Value)) {
	for i, l := range o.labels {
		f(i, l, o.values[i])
	}
}

type BinaryFunction func(l, r Value) (Value, error)

type BinaryFuncSignature struct {
	Operator    ast.OperatorKind
	Left, Right semantic.Nature
}

// LookupBinaryFunction returns an appropriate binary function that evaluates two values and returns another value.
// If the two types are not compatible with the given operation, this returns an error.
func LookupBinaryFunction(sig BinaryFuncSignature) (BinaryFunction, error) {
	f, ok := binaryFuncLookup[sig]
	if !ok {
		return nil, errors.Newf(codes.Invalid, "unsupported binary expression %v %v %v", sig.Left, sig.Operator, sig.Right)
	}
	return binaryFuncNullCheck(f), nil
}

// binaryFuncNullCheck will wrap any BinaryFunction and
// check that both of the arguments are non-nil.
//
// If either value is null, then it will return null.
// Otherwise, it will invoke the function to retrieve the result.
func binaryFuncNullCheck(fn BinaryFunction) BinaryFunction {
	return fn
	// return func(lv, rv Value) (Value, error) {
	// 	// if lv.IsNull() || rv.IsNull() {
	// 	// 	return Null, nil
	// 	// }
	// 	return fn(lv, rv)
	// }
}

var binaryFuncLookup = map[BinaryFuncSignature]BinaryFunction{
	{Operator: ast.AdditionOperator, Left: semantic.Float, Right: semantic.Float}: func(lv, rv Value) (Value, error) {
		l := lv.Float()
		r := rv.Float()
		return NewFloat(l + r), nil
	},
	{Operator: ast.GreaterThanOperator, Left: semantic.Float, Right: semantic.Float}: func(lv, rv Value) (Value, error) {
		l := lv.Float()
		r := rv.Float()
		return NewBool(l > r), nil
	},
}

func ValueForRow(cr flux.ColReader, i, j int) Value {
	t := cr.Cols()[j].Type
	switch t {
	case flux.TString:
		if cr.Strings(j).IsNull(i) {
			return Value{}
		}
		return NewString(cr.Strings(j).Value(i))
	case flux.TInt:
		if cr.Ints(j).IsNull(i) {
			return Value{}
		}
		return NewInt(cr.Ints(j).Value(i))
	case flux.TUInt:
		return Value{}
	case flux.TFloat:
		if cr.Floats(j).IsNull(i) {
			return Value{}
		}
		return NewFloat(cr.Floats(j).Value(i))
	case flux.TBool:
		if cr.Bools(j).IsNull(i) {
			return Value{}
		}
		return NewBool(cr.Bools(j).Value(i))
	case flux.TTime:
		return Value{}
	default:
		return Value{}
	}
}
