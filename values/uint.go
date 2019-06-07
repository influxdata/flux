package values

import (
	"fmt"
	"regexp"

	"github.com/influxdata/flux/semantic"
)

// UInt represents a uint value.
type UInt struct {
	Value uint64
	Null  bool
}

func (UInt) Type() semantic.Type {
	return semantic.UInt
}

func (UInt) PolyType() semantic.PolyType {
	return semantic.UInt
}

func (v UInt) UInt() uint64 {
	return v.Value
}

func (v UInt) IsNull() bool {
	return v.Null
}

func (UInt) Str() string            { panic(UnexpectedKind(semantic.String, semantic.UInt)) }
func (UInt) Int() int64             { panic(UnexpectedKind(semantic.Int, semantic.UInt)) }
func (UInt) Float() float64         { panic(UnexpectedKind(semantic.Float, semantic.UInt)) }
func (UInt) Bool() bool             { panic(UnexpectedKind(semantic.Bool, semantic.UInt)) }
func (UInt) Time() Time             { panic(UnexpectedKind(semantic.Time, semantic.UInt)) }
func (UInt) Duration() Duration     { panic(UnexpectedKind(semantic.Duration, semantic.UInt)) }
func (UInt) Regexp() *regexp.Regexp { panic(UnexpectedKind(semantic.Regexp, semantic.UInt)) }
func (UInt) Array() Array           { panic(UnexpectedKind(semantic.Array, semantic.UInt)) }
func (UInt) Object() Object         { panic(UnexpectedKind(semantic.Object, semantic.UInt)) }
func (UInt) Function() Function     { panic(UnexpectedKind(semantic.Function, semantic.UInt)) }

func (v UInt) Equal(other Value) bool {
	if v.Type() != other.Type() {
		return false
	} else if v.IsNull() || other.IsNull() {
		return false
	}
	return v.UInt() == other.UInt()
}

func (v UInt) String() string {
	if v.Null {
		return "<nil>"
	}
	return fmt.Sprintf("%v", v.Value)
}
