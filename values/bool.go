package values

import (
	"fmt"
	"regexp"

	"github.com/influxdata/flux/semantic"
)

// Bool represents a float value.
type Bool struct {
	Value bool
	Null  bool
}

func (Bool) Type() semantic.Type {
	return semantic.Bool
}

func (Bool) PolyType() semantic.PolyType {
	return semantic.Bool
}

func (v Bool) Bool() bool {
	return v.Value
}

func (v Bool) IsNull() bool {
	return v.Null
}

func (Bool) Str() string            { panic(UnexpectedKind(semantic.String, semantic.Bool)) }
func (Bool) Int() int64             { panic(UnexpectedKind(semantic.Int, semantic.Bool)) }
func (Bool) UInt() uint64           { panic(UnexpectedKind(semantic.UInt, semantic.Bool)) }
func (Bool) Float() float64         { panic(UnexpectedKind(semantic.Float, semantic.Bool)) }
func (Bool) Time() Time             { panic(UnexpectedKind(semantic.Time, semantic.Bool)) }
func (Bool) Duration() Duration     { panic(UnexpectedKind(semantic.Duration, semantic.Bool)) }
func (Bool) Regexp() *regexp.Regexp { panic(UnexpectedKind(semantic.Regexp, semantic.Bool)) }
func (Bool) Array() Array           { panic(UnexpectedKind(semantic.Array, semantic.Bool)) }
func (Bool) Object() Object         { panic(UnexpectedKind(semantic.Object, semantic.Bool)) }
func (Bool) Function() Function     { panic(UnexpectedKind(semantic.Function, semantic.Bool)) }

func (v Bool) Equal(other Value) bool {
	if v.Type() != other.Type() {
		return false
	} else if v.IsNull() || other.IsNull() {
		return false
	}
	return v.Bool() == other.Bool()
}

func (v Bool) String() string {
	if v.Null {
		return "<nil>"
	}
	return fmt.Sprintf("%v", v.Value)
}
