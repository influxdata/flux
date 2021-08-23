package arrow

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/array"
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
	return t.Values[j].(*array.Boolean)
}
func (t *TableBuffer) Ints(j int) *array.Int {
	return t.Values[j].(*array.Int)
}
func (t *TableBuffer) UInts(j int) *array.Uint {
	return t.Values[j].(*array.Uint)
}
func (t *TableBuffer) Floats(j int) *array.Float {
	return t.Values[j].(*array.Float)
}
func (t *TableBuffer) Strings(j int) *array.String {
	return t.Values[j].(*array.String)
}
func (t *TableBuffer) Times(j int) *array.Int {
	return t.Values[j].(*array.Int)
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

	// Retrieve the size of the first column if non-nil to use for checking column size.
	// If the first column is nil, the check below will catch it so we don't care about
	// the size then.
	var sz int
	if t.Values[0] != nil {
		sz = t.Values[0].Len()
	}

	for i := 0; i < len(t.Values); i++ {
		if t.Values[i] == nil {
			return errors.New(codes.Internal, "column data was not initialized")
		}
		if i > 0 && t.Values[i].Len() != sz {
			// Some column was mismatched so generate a nicer error message.
			// We failed anyway. We can spend extra time on this.
			sizes := make([]interface{}, len(t.Values))
			for i, arr := range t.Values {
				if arr != nil {
					sizes[i] = arr.Len()
				}
			}
			return errors.Newf(codes.Internal, "column size mismatch: %v", sizes)
		}
		if ok := t.checkCol(t.Columns[i].Type, t.Values[i]); !ok {
			return errors.Newf(codes.Internal, "column %s of type %s is incompatible with data array %T", t.Columns[i].Label, t.Columns[i].Type, t.Values[i])
		}
	}
	return nil
}

func (t *TableBuffer) checkCol(typ flux.ColType, arr array.Interface) bool {
	switch typ {
	case flux.TInt, flux.TTime:
		_, ok := arr.(*array.Int)
		return ok
	case flux.TUInt:
		_, ok := arr.(*array.Uint)
		return ok
	case flux.TFloat:
		_, ok := arr.(*array.Float)
		return ok
	case flux.TString:
		_, ok := arr.(*array.String)
		return ok
	case flux.TBool:
		_, ok := arr.(*array.Boolean)
		return ok
	default:
		return false
	}
}
