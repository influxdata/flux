package values

import (
	"fmt"
	"regexp"

	"github.com/influxdata/flux/semantic"
)

// String represents a string value.
type String struct {
	Value string
	Null  bool
}

func (String) Type() semantic.Type {
	return semantic.String
}

func (String) PolyType() semantic.PolyType {
	return semantic.String
}

func (v String) Str() string {
	return v.Value
}

func (v String) IsNull() bool {
	return v.Null
}

func (String) Int() int64             { panic(UnexpectedKind(semantic.Int, semantic.String)) }
func (String) UInt() uint64           { panic(UnexpectedKind(semantic.UInt, semantic.String)) }
func (String) Float() float64         { panic(UnexpectedKind(semantic.Float, semantic.String)) }
func (String) Bool() bool             { panic(UnexpectedKind(semantic.Bool, semantic.String)) }
func (String) Time() Time             { panic(UnexpectedKind(semantic.Time, semantic.String)) }
func (String) Duration() Duration     { panic(UnexpectedKind(semantic.Duration, semantic.String)) }
func (String) Regexp() *regexp.Regexp { panic(UnexpectedKind(semantic.Regexp, semantic.String)) }
func (String) Array() Array           { panic(UnexpectedKind(semantic.Array, semantic.String)) }
func (String) Object() Object         { panic(UnexpectedKind(semantic.Object, semantic.String)) }
func (String) Function() Function     { panic(UnexpectedKind(semantic.Function, semantic.String)) }

func (v String) Equal(other Value) bool {
	if v.Type() != other.Type() {
		return false
	} else if v.IsNull() || other.IsNull() {
		return false
	}
	return v.Str() == other.Str()
}

func (v String) String() string {
	if v.Null {
		return "<nil>"
	}
	return fmt.Sprintf("%v", v.Value)
}
