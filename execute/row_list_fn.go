package execute

import (
	"fmt"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/compiler"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
	"regexp"
)

type dynamicListFn struct {
	compilationCache *compiler.CompilationCache
	inRecordList     values.Object

	preparedFn compiler.Func

	recordListName string
	recordList     *RecordList

	recordCols map[string]int
	references []string
}

func newDynamicListFn(fn *semantic.FunctionExpression) dynamicListFn {
	scope := flux.BuiltIns()
	return dynamicListFn{
		compilationCache: compiler.NewCompilationCache(fn, scope),
		inRecordList:     values.NewObject(),
		recordListName:   fn.Block.Parameters.List[0].Key.Name,
		references:       findColReferences(fn),
		recordCols:       make(map[string]int),
	}
}

func (f *dynamicListFn) prepare(cols []flux.ColMeta, extraTypes map[string]semantic.Type) error {
	propertyTypes := make(map[string]semantic.Type, len(f.references))
	f.recordCols = make(map[string]int)

	// Should they have different ObjectTypes?
	for j, c := range cols {
		propertyTypes[c.Label] = ConvertToKind(c.Type)
		f.recordCols[c.Label] = j
	}

	f.recordList = NewRecordList(semantic.NewObjectType(propertyTypes))
	if extraTypes == nil {
		extraTypes = map[string]semantic.Type{
			f.recordListName: f.recordList.Type(),
		}
	} else {
		extraTypes[f.recordListName] = f.recordList.Type()
	}

	f.recordList.cols = f.recordCols

	fn, err := f.compilationCache.Compile(
		semantic.NewObjectType(extraTypes),
	)
	if err != nil {
		return err
	}
	f.preparedFn = fn
	return nil
}

type rowListFn struct {
	dynamicListFn
}

func newRowListFn(fn *semantic.FunctionExpression) (rowListFn, error) {
	return rowListFn{
		dynamicListFn: newDynamicListFn(fn),
	}, nil
}

func (f *rowListFn) eval(values [][]values.Value, extraParams map[string]values.Value) (values.Value, error) {
	for i, row := range values {
		r := NewRecord(f.recordList.Type())
		for n, p := range f.recordCols {
			r.Set(n, row[p])
		}
		f.recordList.Set(string(i), r)
	}
	f.inRecordList.Set(f.recordListName, f.recordList)
	for k, v := range extraParams {
		f.inRecordList.Set(k, v)
	}

	return f.preparedFn.Eval(f.inRecordList)
}

type RowListReduceFn struct {
	rowListFn
}

func NewRowListReduceFn(fn *semantic.FunctionExpression) (*RowListReduceFn, error) {
	r, err := newRowListFn(fn)
	if err != nil {
		return nil, err
	}
	return &RowListReduceFn{
		rowListFn: r,
	}, nil
}

func (f *RowListReduceFn) Prepare(cols []flux.ColMeta) error {
	err := f.rowListFn.prepare(cols, nil)
	if err != nil {
		return err
	}
	k := f.preparedFn.Type().Nature()
	if k != semantic.Object {
		return fmt.Errorf("row list reduce function must return an object, got %s", k.String())
	}
	return nil
}

func (f *RowListReduceFn) Eval(values [][]values.Value, extraParams map[string]values.Value) (values.Object, error) {
	v, err := f.rowListFn.eval(values, extraParams)
	if err != nil {
		return nil, err
	}
	return v.Object(), nil
}

type RecordList struct {
	t semantic.Type
	records map[string]*Record

	cols map[string]int
}

func NewRecordList(t semantic.Type) *RecordList {
	return &RecordList{
		t: t,
		records: make(map[string]*Record),
	}
}

func (r *RecordList) Type() semantic.Type {
	return r.t
}

func (r *RecordList) PolyType() semantic.PolyType {
	return r.t.PolyType()
}

func (r *RecordList) IsNull() bool {
	return false
}

func (r *RecordList) Str() string {
	panic(values.UnexpectedKind(semantic.Object, semantic.String))
}
func (r *RecordList) Int() int64 {
	panic(values.UnexpectedKind(semantic.Object, semantic.Int))
}
func (r *RecordList) UInt() uint64 {
	panic(values.UnexpectedKind(semantic.Object, semantic.UInt))
}
func (r *RecordList) Float() float64 {
	panic(values.UnexpectedKind(semantic.Object, semantic.Float))
}
func (r *RecordList) Bool() bool {
	panic(values.UnexpectedKind(semantic.Object, semantic.Bool))
}
func (r *RecordList) Time() values.Time {
	panic(values.UnexpectedKind(semantic.Object, semantic.Time))
}
func (r *RecordList) Duration() values.Duration {
	panic(values.UnexpectedKind(semantic.Object, semantic.Duration))
}
func (r *RecordList) Regexp() *regexp.Regexp {
	panic(values.UnexpectedKind(semantic.Object, semantic.Regexp))
}
func (r *RecordList) Array() values.Array {
	panic(values.UnexpectedKind(semantic.Object, semantic.Array))
}
func (r *RecordList) Object() values.Object {
	return r
}
func (r *RecordList) Function() values.Function {
	panic(values.UnexpectedKind(semantic.Object, semantic.Function))
}
func (r *RecordList) Equal(rhs values.Value) bool {
	if r.Type() != rhs.Type() {
		return false
	}
	obj := rhs.Object()
	if r.Len() != obj.Len() {
		return false
	}

	for k, v := range r.records {
		val, ok := obj.Get(k)
		if !ok || !v.Equal(val) {
			return false
		}
	}
	return true
}

func (r *RecordList) Set(name string, v values.Value) {
	r.records[name] = v.(*Record)
}

func (r *RecordList) Get(name string) (values.Value, bool) {
	v, ok := r.records[name]
	if !ok {
		return values.Null, false
	}
	return v, true
}

func (r *RecordList) Len() int {
	return len(r.records)
}

func (r *RecordList) Range(f func(name string, v values.Value)) {
	for k, v := range r.records {
		f(k, v)
	}
}
