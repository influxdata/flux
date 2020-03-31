package execute

import (
	"context"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/compiler"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

type dynamicFn struct {
	// Configuration attributes. These are initialized once
	// on creation and used for each new compilation.
	scope      compiler.Scope
	fn         *semantic.FunctionExpression
	recordName string

	// These are initialized after the function is prepared.
	preparedFn compiler.Func
	arg0       values.Object
	args       values.Object
}

func newDynamicFn(fn *semantic.FunctionExpression, scope compiler.Scope) dynamicFn {
	return dynamicFn{
		scope:      scope,
		fn:         fn,
		recordName: fn.Parameters.List[0].Key.Name,
	}
}

// typeof returns an object monotype that matches the current column data.
func (f *dynamicFn) typeof(cols []flux.ColMeta) (semantic.MonoType, error) {
	properties := make([]semantic.PropertyType, len(cols))
	for i, c := range cols {
		vtype := flux.SemanticType(c.Type)
		if vtype.Kind() == semantic.Unknown {
			return semantic.MonoType{}, errors.Newf(codes.Internal, "unknown column type: %s", c.Type)
		}
		properties[i] = semantic.PropertyType{
			Key:   []byte(c.Label),
			Value: vtype,
		}
	}
	return semantic.NewObjectType(properties), nil
}

func (f *dynamicFn) prepare(cols []flux.ColMeta, extraTypes map[string]semantic.MonoType) error {
	// Prepare the type of the record column.
	recordType, err := f.typeof(cols)
	if err != nil {
		return err
	}

	// Prepare the arguments type.
	properties := []semantic.PropertyType{
		{Key: []byte(f.recordName), Value: recordType},
	}
	for name, typ := range extraTypes {
		properties = append(properties, semantic.PropertyType{
			Key:   []byte(name),
			Value: typ,
		})
	}

	inType := semantic.NewObjectType(properties)
	fn, err := compiler.Compile(f.scope, f.fn, inType)
	if err != nil {
		return err
	}
	f.preparedFn = fn

	// Construct the arguments that will be used when evaluating the function.
	f.arg0 = values.NewObject(recordType)
	f.args = values.NewObject(inType)
	f.args.Set(f.recordName, f.arg0)
	return nil
}

// returnType will return the return type of the prepared function.
func (f *dynamicFn) returnType() semantic.MonoType {
	return f.preparedFn.Type()
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
	key := tbl.Key()
	for j, col := range key.Cols() {
		f.arg0.Set(col.Label, key.Value(j))
	}
	return f.preparedFn.Eval(ctx, f.args)
}

type TablePredicateFn struct {
	tableFn
}

func NewTablePredicateFn(fn *semantic.FunctionExpression, scope compiler.Scope) (*TablePredicateFn, error) {
	t := newTableFn(fn, scope)
	return &TablePredicateFn{tableFn: t}, nil
}

func (f *TablePredicateFn) Prepare(tbl flux.Table) error {
	if err := f.prepare(tbl.Key().Cols(), nil); err != nil {
		return err
	} else if f.returnType().Nature() != semantic.Bool {
		return errors.New(codes.Invalid, "table predicate function does not evaluate to a boolean")
	}
	return nil
}

func (f *TablePredicateFn) Eval(ctx context.Context, tbl flux.Table) (bool, error) {
	v, err := f.eval(ctx, tbl)
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
	for j, col := range cr.Cols() {
		f.arg0.Set(col.Label, ValueForRow(cr, row, j))
	}
	for k, v := range extraParams {
		f.args.Set(k, v)
	}
	return f.preparedFn.Eval(ctx, f.args)
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
	if err := f.prepare(cols, nil); err != nil {
		return err
	} else if f.returnType().Nature() != semantic.Bool {
		return errors.New(codes.Invalid, "row predicate function does not evaluate to a boolean")
	}
	return nil
}

// InferredInputType will return the inferred input type. This type may
// contain type variables and will only contain the properties that could be
// inferred from type inference.
func (f *RowPredicateFn) InferredInputType() semantic.MonoType {
	return f.arg0.Type()
}

// InputType will return the prepared input type.
// This will be a fully realized type that was created
// after preparing the function with Prepare.
func (f *RowPredicateFn) InputType() semantic.MonoType {
	return f.arg0.Type()
}

func (f *RowPredicateFn) EvalRow(ctx context.Context, row int, cr flux.ColReader) (bool, error) {
	v, err := f.eval(ctx, row, cr, nil)
	if err != nil {
		return false, err
	}
	return !v.IsNull() && v.Bool(), nil
}

func (f *RowPredicateFn) Eval(ctx context.Context, record values.Object) (bool, error) {
	f.args.Set(f.recordName, record)
	v, err := f.preparedFn.Eval(ctx, f.args)
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
	err := f.prepare(cols, nil)
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
	v, err := f.eval(ctx, row, cr, nil)
	if err != nil {
		return nil, err
	}
	return v.Object(), nil
}

type RowReduceFn struct {
	rowFn
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
	return f.rowFn.prepare(cols, reducerType)
}

func (f *RowReduceFn) Type() semantic.MonoType {
	return f.preparedFn.Type()
}

func (f *RowReduceFn) Eval(ctx context.Context, row int, cr flux.ColReader, extraParams map[string]values.Value) (values.Object, error) {
	v, err := f.rowFn.eval(ctx, row, cr, extraParams)
	if err != nil {
		return nil, err
	}
	return v.Object(), nil
}
