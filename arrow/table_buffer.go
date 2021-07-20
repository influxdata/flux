package arrow

import (
	"reflect"

	"github.com/apache/arrow/go/arrow/array"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
)

// TableBuffer represents the buffered component of an arrow table.
//
// TableBuffer is a low-level structure for creating
// a table that implements the flux.ColReader interface.
// It does not have very many guiding blocks to ensure it is
// used correctly.
//
// A valid TableBuffer will have a number of columns that
// is equal in length to the number of values arrays.
// All of the values arrays will have the same length.
type TableBuffer struct {
	GroupKey flux.GroupKey
	Columns  []flux.ColMeta
	Values   []array.Interface
}

var _ flux.ColReader = (*TableBuffer)(nil)

func (t *TableBuffer) Len() int {
	if len(t.Columns) == 0 {
		return 0
	}
	return t.Values[0].Len()
}

func (t *TableBuffer) Bools(j int) *array.Boolean {
	return AsBools(t.Values[j])
}
func (t *TableBuffer) Ints(j int) *array.Int64 {
	return AsInts(t.Values[j])
}
func (t *TableBuffer) UInts(j int) *array.Uint64 {
	return AsUints(t.Values[j])
}
func (t *TableBuffer) Floats(j int) *array.Float64 {
	return AsFloats(t.Values[j])
}
func (t *TableBuffer) Strings(j int) *array.Binary {
	return AsStrings(t.Values[j])
}
func (t *TableBuffer) Times(j int) *array.Int64 {
	return AsInts(t.Values[j])
}

func (t *TableBuffer) Retain() {
	for _, vs := range t.Values {
		vs.Retain()
	}
}

func (t *TableBuffer) Release() {
	for _, vs := range t.Values {
		vs.Release()
	}
}

func (t *TableBuffer) Key() flux.GroupKey {
	return t.GroupKey
}

func (t *TableBuffer) Cols() []flux.ColMeta {
	return t.Columns
}

// Validate will validate that this TableBuffer has the
// proper structure.
func (t *TableBuffer) Validate() error {
	if len(t.Columns) != len(t.Values) {
		return errors.Newf(codes.Internal, "mismatched number of columns and arrays: %d != %d", len(t.Columns), len(t.Values))
	}

	// If a table has no columns, do not validate the length.
	if len(t.Columns) == 0 {
		return nil
	}

	sz := t.Values[0].Len()
	for i := 1; i < len(t.Values); i++ {
		if t.Values[i].Len() != sz {
			return errors.New(codes.Internal, "column size mismatch")
		}
	}
	return nil
}

// As will attempt to use arr as the target interface.
// The target should be a pointer to a concrete arrow
// array structure.
func As(arr array.Interface, target interface{}) bool {
	if a, ok := arr.(interface{ As(target interface{}) bool }); ok {
		if a.As(target) {
			return true
		}
	}

	targetV := reflect.ValueOf(target)
	v := reflect.ValueOf(arr)
	if v.Type().AssignableTo(targetV.Elem().Type()) {
		targetV.Elem().Set(v)
		return true
	}
	return false
}
