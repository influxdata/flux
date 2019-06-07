package values

import (
	"fmt"
	"regexp"

	"github.com/influxdata/flux/semantic"
)

// Int represents an int value.
type Int struct {
	Value int64
	Null  bool
}

func (Int) Type() semantic.Type {
	return semantic.Int
}

func (Int) PolyType() semantic.PolyType {
	return semantic.Int
}

func (v Int) Int() int64 {
	return v.Value
}

func (v Int) IsNull() bool {
	return v.Null
}

func (Int) Str() string            { panic(UnexpectedKind(semantic.String, semantic.Int)) }
func (Int) UInt() uint64           { panic(UnexpectedKind(semantic.UInt, semantic.Int)) }
func (Int) Float() float64         { panic(UnexpectedKind(semantic.Float, semantic.Int)) }
func (Int) Bool() bool             { panic(UnexpectedKind(semantic.Bool, semantic.Int)) }
func (Int) Time() Time             { panic(UnexpectedKind(semantic.Time, semantic.Int)) }
func (Int) Duration() Duration     { panic(UnexpectedKind(semantic.Duration, semantic.Int)) }
func (Int) Regexp() *regexp.Regexp { panic(UnexpectedKind(semantic.Regexp, semantic.Int)) }
func (Int) Array() Array           { panic(UnexpectedKind(semantic.Array, semantic.Int)) }
func (Int) Object() Object         { panic(UnexpectedKind(semantic.Object, semantic.Int)) }
func (Int) Function() Function     { panic(UnexpectedKind(semantic.Function, semantic.Int)) }

func (v Int) Equal(other Value) bool {
	if v.Type() != other.Type() {
		return false
	} else if v.IsNull() || other.IsNull() {
		return false
	}
	return v.Int() == other.Int()
}

func (v Int) String() string {
	if v.Null {
		return "<nil>"
	}
	return fmt.Sprintf("%v", v.Value)
}
