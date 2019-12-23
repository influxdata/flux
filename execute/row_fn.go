package execute

import (
	"context"
	"regexp"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/compiler"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

type dynamicFn struct {
	compilationCache *compiler.CompilationCache
	inRecord         values.Object

	preparedFn compiler.Func

	recordName string
	record     *Record

	recordCols map[string]int
	references []string
}

func newDynamicFn(fn *semantic.FunctionExpression, scope compiler.Scope) dynamicFn {
	return dynamicFn{
		compilationCache: compiler.NewCompilationCache(fn, scope),
		inRecord:         values.NewObject(),
		recordName:       fn.Block.Parameters.List[0].Key.Name,
		references:       findColReferences(fn),
		recordCols:       make(map[string]int),
	}
}

func (f *dynamicFn) prepare(cols []flux.ColMeta, extraTypes map[string]semantic.MonoType) error {
	// Prepare types and recordCols
	propertyTypes := make(map[string]semantic.Nature, len(f.references))
	f.recordCols = make(map[string]int)
	for j, c := range cols {
		propertyTypes[c.Label] = ConvertToKind(c.Type)
		f.recordCols[c.Label] = j
	}

	f.record = NewRecord(semantic.NewObjectType())
	if extraTypes == nil {
		extraTypes = map[string]semantic.MonoType{
			f.recordName: f.record.Type(),
		}
	} else {
		extraTypes[f.recordName] = f.record.Type()
	}

	// Compile fn for given types
	fn, err := f.compilationCache.Compile(semantic.NewObjectType())
	if err != nil {
		return err
	}
	f.preparedFn = fn
	return nil
}

func ConvertToKind(t flux.ColType) semantic.Nature {
	// TODO make this an array lookup.
	switch t {
	case flux.TInvalid:
		return semantic.Invalid
	case flux.TBool:
		return semantic.Bool
	case flux.TInt:
		return semantic.Int
	case flux.TUInt:
		return semantic.UInt
	case flux.TFloat:
		return semantic.Float
	case flux.TString:
		return semantic.String
	case flux.TTime:
		return semantic.Time
	default:
		return semantic.Invalid
	}
}

func ConvertFromKind(k semantic.Nature) flux.ColType {
	// TODO make this an array lookup.
	switch k {
	case semantic.Invalid:
		return flux.TInvalid
	case semantic.Bool:
		return flux.TBool
	case semantic.Int:
		return flux.TInt
	case semantic.UInt:
		return flux.TUInt
	case semantic.Float:
		return flux.TFloat
	case semantic.String:
		return flux.TString
	case semantic.Time:
		return flux.TTime
	default:
		return flux.TInvalid
	}
}

type tableFn struct {
	dynamicFn
}

func newTableFn(fn *semantic.FunctionExpression, scope compiler.Scope) tableFn {
	return tableFn{
		dynamicFn: newDynamicFn(fn, scope),
	}
}

func (f *tableFn) eval(ctx context.Context, tbl flux.Table) (values.Value, error) {
	for r, col := range f.recordCols {
		f.record.Set(r, tbl.Key().Value(col))
	}
	f.inRecord.Set(f.recordName, f.record)
	return f.preparedFn.Eval(ctx, f.inRecord)
}

type TablePredicateFn struct {
	tableFn
}

func NewTablePredicateFn(fn *semantic.FunctionExpression, scope compiler.Scope) (*TablePredicateFn, error) {
	t := newTableFn(fn, scope)
	return &TablePredicateFn{tableFn: t}, nil
}

func (f *TablePredicateFn) Prepare(tbl flux.Table) error {
	if err := f.tableFn.prepare(tbl.Key().Cols(), nil); err != nil {
		return err
	}
	if f.preparedFn.Type().Nature() != semantic.Bool {
		return errors.New(codes.Invalid, "table predicate function does not evaluate to a boolean")
	}
	return nil
}

func (f *TablePredicateFn) Eval(ctx context.Context, tbl flux.Table) (bool, error) {
	v, err := f.tableFn.eval(ctx, tbl)
	if err != nil {
		return false, err
	}
	return !v.IsNull() && v.Bool(), nil
}

type rowFn struct {
	dynamicFn
}

func newRowFn(fn *semantic.FunctionExpression, scope compiler.Scope) (rowFn, error) {
	return rowFn{
		dynamicFn: newDynamicFn(fn, scope),
	}, nil
}

func (f *rowFn) eval(ctx context.Context, row int, cr flux.ColReader, extraParams map[string]values.Value) (values.Value, error) {
	for r, col := range f.recordCols {
		f.record.Set(r, ValueForRow(cr, row, col))
	}
	f.inRecord.Set(f.recordName, f.record)
	for k, v := range extraParams {
		f.inRecord.Set(k, v)
	}

	return f.preparedFn.Eval(ctx, f.inRecord)
}

type RowPredicateFn struct {
	rowFn
}

func NewRowPredicateFn(fn *semantic.FunctionExpression, scope compiler.Scope) (*RowPredicateFn, error) {
	r, err := newRowFn(fn, scope)
	if err != nil {
		return nil, err
	}
	return &RowPredicateFn{
		rowFn: r,
	}, nil
}

func (f *RowPredicateFn) Prepare(cols []flux.ColMeta) error {
	if err := f.rowFn.prepare(cols, nil); err != nil {
		return err
	}
	if f.preparedFn.Type().Nature() != semantic.Bool {
		return errors.New(codes.Invalid, "row predicate function does not evaluate to a boolean")
	}
	return nil
}

func (f *RowPredicateFn) InputType() semantic.MonoType {
	t := f.preparedFn.Type()
	arg, _ := t.Argument(0)
	at, _ := arg.TypeOf()
	return at
}

func (f *RowPredicateFn) EvalRow(ctx context.Context, row int, cr flux.ColReader) (bool, error) {
	v, err := f.rowFn.eval(ctx, row, cr, nil)
	if err != nil {
		return false, err
	}
	return !v.IsNull() && v.Bool(), nil
}

func (f *RowPredicateFn) Eval(ctx context.Context, record values.Object) (bool, error) {
	f.inRecord.Set(f.recordName, record)
	v, err := f.preparedFn.Eval(ctx, f.inRecord)
	if err != nil {
		return false, err
	}
	return !v.IsNull() && v.Bool(), nil
}

type RowMapFn struct {
	rowFn
}

func NewRowMapFn(fn *semantic.FunctionExpression, scope compiler.Scope) (*RowMapFn, error) {
	r, err := newRowFn(fn, scope)
	if err != nil {
		return nil, err
	}
	return &RowMapFn{
		rowFn: r,
	}, nil
}

func (f *RowMapFn) Prepare(cols []flux.ColMeta) error {
	err := f.dynamicFn.prepare(cols, nil)
	if err != nil {
		return err
	}
	k := f.preparedFn.Type().Nature()
	if k != semantic.Object {
		return errors.Newf(codes.Invalid, "map function must return an object, got %s", k.String())
	}
	return nil
}

func (f *RowMapFn) Type() semantic.MonoType {
	return f.preparedFn.Type()
}

func (f *RowMapFn) Eval(ctx context.Context, row int, cr flux.ColReader) (values.Object, error) {
	v, err := f.rowFn.eval(ctx, row, cr, nil)
	if err != nil {
		return nil, err
	}
	return v.Object(), nil
}

type RowReduceFn struct {
	rowFn
	isWrap  bool
	wrapObj *Record
}

func NewRowReduceFn(fn *semantic.FunctionExpression, scope compiler.Scope) (*RowReduceFn, error) {
	r, err := newRowFn(fn, scope)
	if err != nil {
		return nil, err
	}
	return &RowReduceFn{
		rowFn: r,
	}, nil
}

func (f *RowReduceFn) Prepare(cols []flux.ColMeta, reducerType map[string]semantic.MonoType) error {
	err := f.rowFn.prepare(cols, reducerType)
	if err != nil {
		return err
	}
	k := f.preparedFn.Type().Nature()
	f.isWrap = k != semantic.Object
	if f.isWrap {
		f.wrapObj = NewRecord(semantic.NewObjectType(
		//TODO (algow): create correct type
		//map[string]semantic.MonoType{
		//	DefaultValueColLabel: f.preparedFn.Type(),
		//},
		))
	}
	return nil
}

func (f *RowReduceFn) Type() semantic.MonoType {
	if f.isWrap {
		return f.wrapObj.Type()
	}
	return f.preparedFn.Type()
}

func (f *RowReduceFn) Eval(ctx context.Context, row int, cr flux.ColReader, extraParams map[string]values.Value) (values.Object, error) {
	v, err := f.rowFn.eval(ctx, row, cr, extraParams)
	if err != nil {
		return nil, err
	}
	if f.isWrap {
		f.wrapObj.Set(DefaultValueColLabel, v)
		return f.wrapObj, nil
	}
	return v.Object(), nil
}

func findColReferences(fn *semantic.FunctionExpression) []string {
	v := &colReferenceVisitor{
		recordName: fn.Block.Parameters.List[0].Key.Name,
	}
	semantic.Walk(v, fn)
	return v.refs
}

type colReferenceVisitor struct {
	recordName string
	refs       []string
}

func (c *colReferenceVisitor) Visit(node semantic.Node) semantic.Visitor {
	if me, ok := node.(*semantic.MemberExpression); ok {
		if obj, ok := me.Object.(*semantic.IdentifierExpression); ok && obj.Name == c.recordName {
			c.refs = append(c.refs, me.Property)
		}
	}
	return c
}

func (c *colReferenceVisitor) Done(semantic.Node) {}

type Record struct {
	t      semantic.MonoType
	values map[string]values.Value
}

func NewRecord(t semantic.MonoType) *Record {
	return &Record{
		t:      t,
		values: make(map[string]values.Value),
	}
}

func (r *Record) Type() semantic.MonoType {
	return r.t
}

func (r *Record) IsNull() bool {
	return false
}
func (r *Record) Str() string {
	panic(values.UnexpectedKind(semantic.Object, semantic.String))
}
func (r *Record) Bytes() []byte {
	panic(values.UnexpectedKind(semantic.Object, semantic.Bytes))
}
func (r *Record) Int() int64 {
	panic(values.UnexpectedKind(semantic.Object, semantic.Int))
}
func (r *Record) UInt() uint64 {
	panic(values.UnexpectedKind(semantic.Object, semantic.UInt))
}
func (r *Record) Float() float64 {
	panic(values.UnexpectedKind(semantic.Object, semantic.Float))
}
func (r *Record) Bool() bool {
	panic(values.UnexpectedKind(semantic.Object, semantic.Bool))
}
func (r *Record) Time() values.Time {
	panic(values.UnexpectedKind(semantic.Object, semantic.Time))
}
func (r *Record) Duration() values.Duration {
	panic(values.UnexpectedKind(semantic.Object, semantic.Duration))
}
func (r *Record) Regexp() *regexp.Regexp {
	panic(values.UnexpectedKind(semantic.Object, semantic.Regexp))
}
func (r *Record) Array() values.Array {
	panic(values.UnexpectedKind(semantic.Object, semantic.Array))
}
func (r *Record) Object() values.Object {
	return r
}
func (r *Record) Function() values.Function {
	panic(values.UnexpectedKind(semantic.Object, semantic.Function))
}
func (r *Record) Equal(rhs values.Value) bool {
	if r.Type() != rhs.Type() {
		return false
	}
	obj := rhs.Object()
	if r.Len() != obj.Len() {
		return false
	}
	for k, v := range r.values {
		val, ok := obj.Get(k)
		if !ok || !v.Equal(val) {
			return false
		}
	}
	return true
}

func (r *Record) Set(name string, v values.Value) {
	r.values[name] = v
}
func (r *Record) Get(name string) (values.Value, bool) {
	v, ok := r.values[name]
	if !ok {
		return values.Null, false
	}
	return v, true
}
func (r *Record) Len() int {
	return len(r.values)
}

func (r *Record) Range(f func(name string, v values.Value)) {
	for k, v := range r.values {
		f(k, v)
	}
}
