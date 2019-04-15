package objects

import (
	"bytes"
	"fmt"
	"regexp"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

// A Table is an object with a schema.
// Schema is the only attribute that the table exposes. To extract records or columns
// one must use the provided functions (getRecord/getColumn).
// TODO(affo): we decided not to expose the schema, for now.
var (
	TableType = semantic.NewObjectPolyType(
		map[string]semantic.PolyType{
			"schema": semantic.NewArrayPolyType(SchemaType),
		},
		semantic.LabelSet{"schema"},
		semantic.LabelSet{"schema"},
	)
	TableMonoType, _ = TableType.MonoType()
	SchemaType       = semantic.NewObjectPolyType(
		map[string]semantic.PolyType{
			"label":   semantic.String,
			"grouped": semantic.Bool,
			// TODO(affo): we cannot express types as values in Flux, by now, so we use strings.
			"type": semantic.String,
		},
		semantic.LabelSet{"label", "type", "grouped"},
		semantic.LabelSet{"label", "type", "grouped"},
	)
	SchemaMonoType, _ = SchemaType.MonoType()
)

// Table is a values.Value that represents a table in Flux.
// (Unlike flux.TableObject which represents a stream of tables.)
type Table struct {
	flux.Table
	schema values.Array
}

func NewTable(tbl flux.Table) *Table {
	// We need to cache the content of the table in order to make subsequent calls
	// to getRecord/Column idempotent. If we don't cache the table, then it would be
	// consumed by calls to `Do`, and subsequent calls to getRecord/Column would find
	// an empty table.
	t := &Table{Table: &cachedTable{Table: tbl}}
	t.schema = values.NewArray(SchemaMonoType)
	for _, c := range tbl.Cols() {
		t.schema.Append(values.NewObjectWithValues(map[string]values.Value{
			"label":   values.New(c.Label),
			"grouped": values.New(tbl.Key().HasCol(c.Label)),
			"type":    values.New(c.Type.String()),
		}))
	}
	return t
}

func (t *Table) Get(name string) (values.Value, bool) {
	// TODO(affo): uncomment this block and remove the return statement once we decide to expose the schema.
	/*
		if name != "schema" {
			return nil, false
		}
		return t.schema, true
	*/
	return nil, false
}

func (t *Table) Set(name string, v values.Value) {
	// immutable
}

func (t *Table) Len() int {
	return 1
}

func (t *Table) Range(fn func(name string, v values.Value)) {
	fn("schema", t.schema)
}

func (t *Table) Type() semantic.Type {
	return TableMonoType
}

func (t *Table) PolyType() semantic.PolyType {
	return TableType
}

func (t *Table) IsNull() bool {
	return false
}

func (t *Table) Str() string {
	panic(values.UnexpectedKind(semantic.Object, semantic.String))
}

func (t *Table) Int() int64 {
	panic(values.UnexpectedKind(semantic.Object, semantic.Int))
}

func (t *Table) UInt() uint64 {
	panic(values.UnexpectedKind(semantic.Object, semantic.UInt))
}

func (t *Table) Float() float64 {
	panic(values.UnexpectedKind(semantic.Object, semantic.Float))
}

func (t *Table) Bool() bool {
	panic(values.UnexpectedKind(semantic.Object, semantic.Bool))
}

func (t *Table) Time() values.Time {
	panic(values.UnexpectedKind(semantic.Object, semantic.Time))
}

func (t *Table) Duration() values.Duration {
	panic(values.UnexpectedKind(semantic.Object, semantic.Duration))
}

func (t *Table) Regexp() *regexp.Regexp {
	panic(values.UnexpectedKind(semantic.Object, semantic.Regexp))
}

func (t *Table) Array() values.Array {
	panic(values.UnexpectedKind(semantic.Object, semantic.Array))
}

func (t *Table) Object() values.Object {
	return t
}

func (t *Table) Function() values.Function {
	panic(values.UnexpectedKind(semantic.Object, semantic.Function))
}

func (t *Table) Equal(other values.Value) bool {
	ot, ok := other.(*Table)
	if !ok || !t.schema.Equal(ot.schema) {
		return false
	}
	eq, err := execute.TablesEqual(t.Table, ot.Table, &memory.Allocator{})
	if err != nil || !eq {
		return false
	}
	return true
}

func (t *Table) String() string {
	w := bytes.NewBuffer([]byte{})
	if _, err := execute.NewFormatter(t, nil).WriteTo(w); err != nil {
		return fmt.Sprintf("error while formatting table: %v", err)
	}
	return w.String()
}

// cachedTable caches the column reader extracted from a table on `Do` for further usage.
type cachedTable struct {
	flux.Table
	cr flux.ColReader
}

func (ct *cachedTable) Do(f func(flux.ColReader) error) error {
	if ct.cr != nil {
		return f(ct.cr)
	}
	return ct.Table.Do(func(cr flux.ColReader) error {
		ct.cr = cr
		return f(cr)
	})
}
