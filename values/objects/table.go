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
	SchemaMonoType = semantic.NewObjectType([]semantic.PropertyType{
		{
			Key:   []byte("label"),
			Value: semantic.BasicString,
		},
		{
			Key:   []byte("grouped"),
			Value: semantic.BasicBool,
		},
		{
			Key:   []byte("type"),
			Value: semantic.BasicString,
		},
	})
	TableMonoType = semantic.NewObjectType([]semantic.PropertyType{{
		Key:   []byte("schema"),
		Value: semantic.NewArrayType(SchemaMonoType),
	}})
)

// Table is a values.Value that represents a table in Flux.
// (Unlike flux.TableObject which represents a stream of tables.)
type Table struct {
	flux.BufferedTable
	schema values.Array
}

func NewTable(tbl flux.Table) (*Table, error) {
	bt, err := execute.CopyTable(tbl)
	if err != nil {
		return nil, err
	}
	t := &Table{BufferedTable: bt}
	t.schema = values.NewArray(semantic.NewArrayType(SchemaMonoType))
	for _, c := range tbl.Cols() {
		t.schema.Append(values.NewObjectWithValues(map[string]values.Value{
			"label":   values.New(c.Label),
			"grouped": values.New(tbl.Key().HasCol(c.Label)),
			"type":    values.New(c.Type.String()),
		}))
	}
	return t, nil
}

func (t *Table) Get(name string) (values.Value, bool) {
	// TODO(affo): uncomment this block and remove the return statement once we decide to expose the schema.
	/*
		if name != "schema" {
			return nil, false
		}
		return t.schema, true
	*/
	return values.Null, false
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

func (t *Table) Type() semantic.MonoType {
	return TableMonoType
}

func (t *Table) IsNull() bool {
	return false
}

func (t *Table) Str() string {
	panic(values.UnexpectedKind(semantic.Object, semantic.String))
}

func (t *Table) Bytes() []byte {
	panic(values.UnexpectedKind(semantic.Object, semantic.Bytes))
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

// Table returns a copy of the Table that can be called
// with Do. Either Do or Done must be called on the
// returned Table.
func (t *Table) Table() flux.Table {
	return t.Copy()
}

func (t *Table) Equal(other values.Value) bool {
	ot, ok := other.(*Table)
	if !ok || !t.schema.Equal(ot.schema) {
		return false
	}
	eq, err := execute.TablesEqual(t.BufferedTable, ot.BufferedTable, &memory.Allocator{})
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
