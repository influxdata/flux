package cloudwatch

import (
	"reflect"
	"time"

	"github.com/apache/arrow/go/arrow/array"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/execute/table"
	"github.com/influxdata/flux/memory"
	"github.com/pkg/errors"
)

type typeinfo struct {
	pos int
	typ flux.ColType
}

type NamedTableBuilder struct {
	arrow.TableBuffer
	valueBuilders []array.Builder
	allocator     execute.Allocator
	row           int
	nextFreeCol   int // uint16?
	writtenCols   map[string]bool
	typeMap       map[string]typeinfo
}

func NewNamedTableBuilder(groupKey flux.GroupKey, mem *memory.Allocator) *NamedTableBuilder {
	return &NamedTableBuilder{
		TableBuffer: arrow.TableBuffer{
			GroupKey: groupKey,
		},
		allocator:   execute.Allocator{mem},
		writtenCols: map[string]bool{},
		row:         0,
		typeMap:     map[string]typeinfo{},
	}
}

// SetCol sets a column by name. If the column doesn't already exist, it will be added.
// If the column is being added after the first row, previous rows for the column will be fulled with nulls.
// There is no protection against trying to set the same col twice (which won't work)
func (t *NamedTableBuilder) SetCol(name string, colType flux.ColType, value interface{}) error {
	// check to see if field exist
	typInfo, found := t.typeMap[name]
	// add if not
	if !found {
		pos := t.nextFreeCol
		t.nextFreeCol++
		t.Columns = append(t.Columns, flux.ColMeta{
			Label: name,
			Type:  colType,
		})
		t.valueBuilders = append(t.valueBuilders, t.newColsFromType(colType))
		typInfo = typeinfo{
			pos: pos,
			typ: colType,
		}
		t.typeMap[name] = typInfo
	}
	// type check
	valType := typeFromValue(value)
	if value != nil && valType != colType {
		// could try to coerce it. int => float, int => string, string => int may make sense.
		return errors.Errorf("expected type %s, but got type %s for value %q", colType, valType, value)
	}

	arr := t.valueBuilders[typInfo.pos]
	if value == nil {
		arr.AppendNull()
	} else {
		switch colType {
		case flux.TBool:
			b := arr.(*array.BooleanBuilder)
			b.Append(reflect.ValueOf(value).Bool())
		case flux.TInt:
			b := arr.(*array.Int64Builder)
			b.Append(reflect.ValueOf(value).Int())
		case flux.TUInt:
			b := arr.(*array.Uint64Builder)
			b.Append(reflect.ValueOf(value).Uint())
		case flux.TFloat:
			b := arr.(*array.Float64Builder)
			b.Append(reflect.ValueOf(value).Float())
		case flux.TString:
			b := arr.(*array.BinaryBuilder)
			b.AppendString(reflect.ValueOf(value).String())
		case flux.TTime:
			b := arr.(*array.Int64Builder)
			switch tm := value.(type) {
			case time.Time:
				b.Append(tm.UnixNano())
			default:
				b.Append(reflect.ValueOf(value).Int())
			}
		}
	}

	// add to this list of columns we've written to this row
	t.writtenCols[name] = true

	return nil
}

// NextRow closes out the current row, writing nulls to fields that weren't specifically written to.
func (t *NamedTableBuilder) NextRow() {
	// close out the current row
	for _, col := range t.Columns {
		if !t.writtenCols[col.Label] {
			t.SetCol(col.Label, col.Type, nil)
		}
	}

	// reset writtenCols
	t.writtenCols = map[string]bool{}

	t.row += 1
}

// Table builds and returns the final table.
func (t *NamedTableBuilder) Table() (flux.Table, error) {
	// generate all the values
	t.Values = make([]array.Interface, len(t.valueBuilders))
	for i, b := range t.valueBuilders {
		t.Values[i] = b.NewArray()
	}

	// validate
	if err := t.Validate(); err != nil {
		return nil, err // table state might not be any good after this point.
	}
	// convert to table.
	bb := table.NewBufferedBuilder(t.GroupKey, t.allocator.Allocator)
	err := bb.AppendBuffer(t)
	if err != nil {
		return nil, err
	}
	return bb.Table()
}

func typeFromValue(val interface{}) flux.ColType {
	switch val.(type) {
	case bool:
		return flux.TBool
	case int8, int, int16, int32, int64:
		return flux.TInt
	case uint8, uint, uint16, uint32, uint64:
		return flux.TUInt
	case float32, float64:
		return flux.TFloat
	case string:
		return flux.TString
	case time.Time:
		return flux.TTime
	default:
		return flux.TInvalid
	}
}

func (t *NamedTableBuilder) newColsFromType(ty flux.ColType) array.Builder {
	var b array.Builder
	switch ty {
	case flux.TBool:
		b = array.NewBooleanBuilder(t.allocator.Allocator)
	case flux.TInt:
		b = array.NewInt64Builder(t.allocator.Allocator)
	case flux.TUInt:
		b = array.NewUint64Builder(t.allocator.Allocator)
	case flux.TFloat:
		b = array.NewFloat64Builder(t.allocator.Allocator)
	case flux.TString:
		b = arrow.NewStringBuilder(t.allocator.Allocator)
	case flux.TTime:
		b = array.NewInt64Builder(t.allocator.Allocator)
	default:
		panic("Unknown column type " + ty.String())
	}
	// resize didn't work :(
	for i := 0; i < t.row; i++ {
		b.AppendNull()
	}
	return b
}
