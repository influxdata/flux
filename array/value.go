package array

import (
	"fmt"

	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

// ValueBuilder is a wrapper around the various builders
// that allows appending to a builder using values.Value types.
type ValueBuilder struct {
	// Builder holds the BaseBuilder that will be used.
	Builder BaseBuilder
}

func (b *ValueBuilder) Type() semantic.Type {
	return b.Builder.Type()
}

func (b *ValueBuilder) Len() int {
	return b.Builder.Len()
}

func (b *ValueBuilder) Cap() int {
	return b.Builder.Cap()
}

func (b *ValueBuilder) Reserve(n int) {
	b.Builder.Reserve(n)
}

func (b *ValueBuilder) AppendNull() {
	b.Builder.AppendNull()
}

func (b *ValueBuilder) BuildArray() Base {
	return b.Builder.BuildArray()
}

func (b *ValueBuilder) Append(v values.Value) error {
	if vtype := v.Type(); b.Type() != vtype {
		return fmt.Errorf("invalid value type: %s != %s", b.Type(), vtype)
	}

	switch b := b.Builder.(type) {
	case FloatBuilder:
		b.Append(v.Float())
	case IntBuilder:
		b.Append(v.Int())
	case UIntBuilder:
		b.Append(v.UInt())
	case StringBuilder:
		b.Append(v.Str())
	case BooleanBuilder:
		b.Append(v.Bool())
	case TimeBuilder:
		b.Append(v.Time())
	default:
		return fmt.Errorf("unsupported value type: %s", b.Type())
	}
	return nil
}
