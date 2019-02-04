package values

import (
	"regexp"

	"github.com/influxdata/flux/semantic"
)

// Float represents a float value.
type Float struct {
	Value float64
	Null  bool
}

func (Float) Type() semantic.Type {
	return semantic.Float
}

func (Float) PolyType() semantic.PolyType {
	return semantic.Float
}

func (v Float) Float() float64 {
	return v.Value
}

func (v Float) IsNull() bool {
	return v.Null
}

func (Float) Str() string            { panic(UnexpectedKind(semantic.String, semantic.Float)) }
func (Float) Int() int64             { panic(UnexpectedKind(semantic.Int, semantic.Float)) }
func (Float) UInt() uint64           { panic(UnexpectedKind(semantic.UInt, semantic.Float)) }
func (Float) Bool() bool             { panic(UnexpectedKind(semantic.Bool, semantic.Float)) }
func (Float) Time() Time             { panic(UnexpectedKind(semantic.Time, semantic.Float)) }
func (Float) Duration() Duration     { panic(UnexpectedKind(semantic.Duration, semantic.Float)) }
func (Float) Regexp() *regexp.Regexp { panic(UnexpectedKind(semantic.Regexp, semantic.Float)) }
func (Float) Array() Array           { panic(UnexpectedKind(semantic.Array, semantic.Float)) }
func (Float) Object() Object         { panic(UnexpectedKind(semantic.Object, semantic.Float)) }
func (Float) Function() Function     { panic(UnexpectedKind(semantic.Function, semantic.Float)) }

func (v Float) Equal(other Value) bool {
	if v.Type() != other.Type() {
		return false
	} else if v.IsNull() || other.IsNull() {
		return false
	}
	return v.Float() == other.Float()
}
