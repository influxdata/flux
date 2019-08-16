// Package values declares the flux data types and implements them.
package values

import (
	"bytes"
	"fmt"
	"regexp"
	"runtime/debug"
	"strconv"

	"github.com/influxdata/flux/semantic"
	"github.com/pkg/errors"
)

type Typer interface {
	Type() semantic.Type
	PolyType() semantic.PolyType
}

type Value interface {
	Typer
	IsNull() bool
	Str() string
	Bytes() []byte
	Int() int64
	UInt() uint64
	Float() float64
	Bool() bool
	Time() Time
	Duration() Duration
	Regexp() *regexp.Regexp
	Array() Array
	Object() Object
	Function() Function
	Equal(Value) bool
}

type ValueStringer interface {
	String() string
}

type value struct {
	t semantic.Type
	v interface{}
}

func (v value) Type() semantic.Type {
	return v.t
}
func (v value) PolyType() semantic.PolyType {
	return v.t.PolyType()
}
func (v value) IsNull() bool {
	return v.v == nil
}
func (v value) Str() string {
	CheckKind(v.t.Nature(), semantic.String)
	return v.v.(string)
}
func (v value) Bytes() []byte {
	CheckKind(v.t.Nature(), semantic.Bytes)
	return v.v.([]byte)
}
func (v value) Int() int64 {
	CheckKind(v.t.Nature(), semantic.Int)
	return v.v.(int64)
}
func (v value) UInt() uint64 {
	CheckKind(v.t.Nature(), semantic.UInt)
	return v.v.(uint64)
}
func (v value) Float() float64 {
	CheckKind(v.t.Nature(), semantic.Float)
	return v.v.(float64)
}
func (v value) Bool() bool {
	CheckKind(v.t.Nature(), semantic.Bool)
	return v.v.(bool)
}
func (v value) Time() Time {
	CheckKind(v.t.Nature(), semantic.Time)
	return v.v.(Time)
}
func (v value) Duration() Duration {
	CheckKind(v.t.Nature(), semantic.Duration)
	return v.v.(Duration)
}
func (v value) Regexp() *regexp.Regexp {
	CheckKind(v.t.Nature(), semantic.Regexp)
	return v.v.(*regexp.Regexp)
}
func (v value) Array() Array {
	CheckKind(v.t.Nature(), semantic.Array)
	return v.v.(Array)
}
func (v value) Object() Object {
	CheckKind(v.t.Nature(), semantic.Object)
	return v.v.(Object)
}
func (v value) Function() Function {
	CheckKind(v.t.Nature(), semantic.Function)
	return v.v.(Function)
}
func (v value) Equal(r Value) bool {
	if v.Type() != r.Type() {
		return false
	}

	if v.IsNull() || r.IsNull() {
		return false
	}

	switch k := v.Type().Nature(); k {
	case semantic.Bool:
		return v.Bool() == r.Bool()
	case semantic.UInt:
		return v.UInt() == r.UInt()
	case semantic.Int:
		return v.Int() == r.Int()
	case semantic.Float:
		return v.Float() == r.Float()
	case semantic.String:
		return v.Str() == r.Str()
	case semantic.Bytes:
		return bytes.Equal(v.Bytes(), r.Bytes())
	case semantic.Time:
		return v.Time() == r.Time()
	case semantic.Duration:
		return v.Duration() == r.Duration()
	case semantic.Regexp:
		return v.Regexp().String() == r.Regexp().String()
	case semantic.Object:
		return v.Object().Equal(r.Object())
	case semantic.Array:
		return v.Array().Equal(r.Array())
	case semantic.Function:
		return v.Function().Equal(r.Function())
	default:
		return false
	}
}

func (v value) String() string {
	return fmt.Sprintf("%v", v.v)
}

var (
	// InvalidValue is a non nil value who's type is semantic.Invalid
	InvalidValue = value{t: semantic.Invalid}

	// Null is an untyped nil value.
	Null = value{t: semantic.Nil}
)

// New constructs a new Value by inferring the type from the interface. If the interface
// does not translate to a valid Value type, then InvalidValue is returned.
func New(v interface{}) Value {
	if v == nil {
		return Null
	}

	switch v := v.(type) {
	case string:
		return NewString(v)
	case []byte:
		return NewBytes(v)
	case int64:
		return NewInt(v)
	case uint64:
		return NewUInt(v)
	case float64:
		return NewFloat(v)
	case bool:
		return NewBool(v)
	case Time:
		return NewTime(v)
	case Duration:
		return NewDuration(v)
	case *regexp.Regexp:
		return NewRegexp(v)
	default:
		return InvalidValue
	}
}

func NewNull(t semantic.Type) Value {
	return value{
		t: t,
		v: nil,
	}
}

func NewFromString(t semantic.Type, s string) (Value, error) {
	var err error
	v := value{t: t}
	switch t {
	case semantic.String:
		v.v = s
	case semantic.Int:
		v.v, err = strconv.ParseInt(s, 10, 64)
		if err != nil {
			return nil, err
		}
	case semantic.UInt:
		v.v, err = strconv.ParseUint(s, 10, 64)
		if err != nil {
			return nil, err
		}
	case semantic.Float:
		v.v, err = strconv.ParseFloat(s, 64)
		if err != nil {
			return nil, err
		}
	case semantic.Bool:
		v.v, err = strconv.ParseBool(s)
		if err != nil {
			return nil, err
		}
	case semantic.Time:
		v.v, err = ParseTime(s)
		if err != nil {
			return nil, err
		}
	case semantic.Duration:
		v.v, err = ParseDuration(s)
		if err != nil {
			return nil, err
		}

	default:
		return nil, errors.New("invalid type for value stringer")
	}
	return v, nil
}

func NewString(v string) Value {
	return value{
		t: semantic.String,
		v: v,
	}
}
func NewBytes(v []byte) Value {
	return value{
		t: semantic.Bytes,
		v: v,
	}
}
func NewInt(v int64) Value {
	return value{
		t: semantic.Int,
		v: v,
	}
}
func NewUInt(v uint64) Value {
	return value{
		t: semantic.UInt,
		v: v,
	}
}
func NewFloat(v float64) Value {
	return value{
		t: semantic.Float,
		v: v,
	}
}
func NewBool(v bool) Value {
	return value{
		t: semantic.Bool,
		v: v,
	}
}
func NewTime(v Time) Value {
	return value{
		t: semantic.Time,
		v: v,
	}
}
func NewDuration(v Duration) Value {
	return value{
		t: semantic.Duration,
		v: v,
	}
}
func NewRegexp(v *regexp.Regexp) Value {
	return value{
		t: semantic.Regexp,
		v: v,
	}
}

// AssignableTo returns true if type V is assignable to type T.
func AssignableTo(V, T semantic.Type) bool {
	switch tn := T.Nature(); tn {
	case semantic.Int,
		semantic.UInt,
		semantic.Float,
		semantic.String,
		semantic.Bool,
		semantic.Time,
		semantic.Duration:
		vn := V.Nature()
		return vn == tn || vn == semantic.Nil
	case semantic.Array:
		if V.Nature() != semantic.Array {
			return false
		}
		// Exact match is required at the moment.
		return V.ElementType() == T.ElementType()
	case semantic.Object:
		if V.Nature() != semantic.Object {
			return false
		}
		properties := V.Properties()
		for name, ttyp := range T.Properties() {
			vtyp, ok := properties[name]
			if !ok {
				vtyp = semantic.Nil
			}

			if !AssignableTo(vtyp, ttyp) {
				return false
			}
		}
		return true
	default:
		return V.Nature() == T.Nature()
	}
}

func UnexpectedKind(got, exp semantic.Nature) error {
	return fmt.Errorf("unexpected kind: got %q expected %q, trace: %s", got, exp, string(debug.Stack()))
}

// CheckKind panics if got != exp.
func CheckKind(got, exp semantic.Nature) {
	if got == exp {
		return
	}

	// Try to see if the two natures are functionally
	// equivalent to see if we are allowed to assign
	// this type to the other type.
	equiv := func(l, r semantic.Nature) bool {
		switch l {
		case semantic.Nil:
			switch r {
			case semantic.Int,
				semantic.UInt,
				semantic.Float,
				semantic.String,
				semantic.Bool,
				semantic.Time,
				semantic.Duration:
				return true
			}
		}
		return false
	}

	// If got and exp are not equivalent in either
	// direction, then panic because we got the wrong
	// kind.
	if !equiv(got, exp) && !equiv(exp, got) {
		panic(UnexpectedKind(got, exp))
	}
}
