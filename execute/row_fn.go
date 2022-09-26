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
	compiledFn *compiledFn
}

type compiledFn struct {
	fn         compiler.Func
	inType     semantic.MonoType
	recordType semantic.MonoType
	cols       []flux.ColMeta
	extraTypes map[string]semantic.MonoType
	vectorized bool
}

func (f *compiledFn) isCacheHit(cols []flux.ColMeta, extraTypes map[string]semantic.MonoType, vectorized bool) bool {
	if f.vectorized != vectorized {
		return false
	}
	if len(f.cols) != len(cols) {
		return false
	}
	for i := range f.cols {
		if f.cols[i] != cols[i] {
			return false
		}
	}
	if len(f.extraTypes) != len(extraTypes) {
		return false
	}
	for k, v := range f.extraTypes {
		if w, ok := extraTypes[k]; !ok || v != w {
			return false
		}
	}
	return true
}

func newDynamicFn(fn *semantic.FunctionExpression, scope compiler.Scope) dynamicFn {
	return dynamicFn{
		scope:      scope,
		fn:         fn,
		recordName: fn.Parameters.List[0].Key.Name.Name(),
	}
}

// typeof returns an object monotype that matches the current column data.
func (f *dynamicFn) typeof(cols []flux.ColMeta, vectorized bool) (semantic.MonoType, error) {
	properties := make([]semantic.PropertyType, len(cols))
	for i, c := range cols {
		vtype := flux.SemanticType(c.Type)
		if vtype.Kind() == semantic.Unknown {
			return semantic.MonoType{}, errors.Newf(codes.Internal, "unknown column type: %s", c.Type)
		}
		if vectorized {
			vtype = semantic.NewVectorType(vtype)
		}
		properties[i] = semantic.PropertyType{
			Key:   []byte(c.Label),
			Value: vtype,
		}
	}
	return semantic.NewObjectType(properties), nil
}

func (f *dynamicFn) compileFunction(ctx context.Context, cols []flux.ColMeta, extraTypes map[string]semantic.MonoType, vectorized bool) error {

	// If the types have not changed we do not need to recompile, just use the cached version
	if f.compiledFn == nil || !f.compiledFn.isCacheHit(cols, extraTypes, vectorized) {
		// Prepare the type of the record column.
		recordType, err := f.typeof(cols, vectorized)
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
		fn, err := compiler.Compile(ctx, f.scope, f.fn, inType)
		if err != nil {
			return err
		}
		f.compiledFn = &compiledFn{
			fn:         fn,
			inType:     inType,
			recordType: recordType,
			cols:       cols,
			extraTypes: extraTypes,
			vectorized: vectorized,
		}
	}
	return nil
}

func (f *dynamicFn) prepare(ctx context.Context, cols []flux.ColMeta, extraTypes map[string]semantic.MonoType, vectorized bool) (preparedFn, error) {
	err := f.compileFunction(ctx, cols, extraTypes, vectorized)
	if err != nil {
		return preparedFn{}, err
	}

	// Construct the arguments that will be used when evaluating the function.
	arg0 := values.NewObject(f.compiledFn.recordType)
	args := values.NewObject(f.compiledFn.inType)
	args.Set(f.recordName, arg0)
	return preparedFn{
		fn:         f.compiledFn.fn,
		recordName: f.recordName,
		arg0:       arg0,
		args:       args,
	}, nil
}

type preparedFn struct {
	fn         compiler.Func
	recordName string
	arg0       values.Object
	args       values.Object
}

// returnType will return the return type of the prepared function.
func (f *preparedFn) returnType() semantic.MonoType {
	return f.fn.Type()
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
	preparedFn
}

func (f *tableFn) eval(ctx context.Context, tbl flux.Table) (values.Value, error) {
	key := tbl.Key()
	for j, col := range key.Cols() {
		f.arg0.Set(col.Label, key.Value(j))
	}
	return f.fn.Eval(ctx, f.args)
}

type TablePredicateFn struct {
	dynamicFn
}

func NewTablePredicateFn(fn *semantic.FunctionExpression, scope compiler.Scope) *TablePredicateFn {
	return &TablePredicateFn{
		dynamicFn: newDynamicFn(fn, scope),
	}
}

func (f *TablePredicateFn) Prepare(ctx context.Context, tbl flux.Table) (*TablePredicatePreparedFn, error) {
	fn, err := f.prepare(ctx, tbl.Key().Cols(), nil, false)
	if err != nil {
		return nil, err
	} else if fn.returnType().Nature() != semantic.Bool {
		return nil, errors.New(codes.Invalid, "table predicate function does not evaluate to a boolean")
	}
	return &TablePredicatePreparedFn{
		tableFn: tableFn{preparedFn: fn},
	}, nil
}

type TablePredicatePreparedFn struct {
	tableFn
}

func (f *TablePredicatePreparedFn) Eval(ctx context.Context, tbl flux.Table) (bool, error) {
	v, err := f.eval(ctx, tbl)
	if err != nil {
		return false, err
	}
	return !v.IsNull() && v.Bool(), nil
}

type rowFn struct {
	preparedFn
}

func (f *rowFn) eval(ctx context.Context, row int, cr flux.ColReader, extraParams map[string]values.Value) (values.Value, error) {
	for j, col := range cr.Cols() {
		f.arg0.Set(col.Label, ValueForRow(cr, row, j))
	}
	for k, v := range extraParams {
		f.args.Set(k, v)
	}
	return f.fn.Eval(ctx, f.args)
}

type RowPredicateFn struct {
	dynamicFn
}

func NewRowPredicateFn(fn *semantic.FunctionExpression, scope compiler.Scope) *RowPredicateFn {
	r := newDynamicFn(fn, scope)
	return &RowPredicateFn{dynamicFn: r}
}

func (f *RowPredicateFn) Prepare(ctx context.Context, cols []flux.ColMeta) (*RowPredicatePreparedFn, error) {
	fn, err := f.prepare(ctx, cols, nil, false)
	if err != nil {
		return nil, err
	} else if fn.returnType().Nature() != semantic.Bool {
		return nil, errors.New(codes.Invalid, "row predicate function does not evaluate to a boolean")
	}
	return &RowPredicatePreparedFn{
		rowFn: rowFn{preparedFn: fn},
	}, nil
}

type RowPredicatePreparedFn struct {
	rowFn
}

// InferredInputType will return the inferred input type. This type may
// contain type variables and will only contain the properties that could be
// inferred from type inference.
func (f *RowPredicatePreparedFn) InferredInputType() semantic.MonoType {
	return f.arg0.Type()
}

// InputType will return the prepared input type.
// This will be a fully realized type that was created
// after preparing the function with Prepare.
func (f *RowPredicatePreparedFn) InputType() semantic.MonoType {
	return f.arg0.Type()
}

func (f *RowPredicatePreparedFn) EvalRow(ctx context.Context, row int, cr flux.ColReader) (bool, error) {
	v, err := f.eval(ctx, row, cr, nil)
	if err != nil {
		return false, err
	}
	return !v.IsNull() && v.Bool(), nil
}

func (f *RowPredicatePreparedFn) Eval(ctx context.Context, record values.Object) (bool, error) {
	f.args.Set(f.recordName, record)
	v, err := f.fn.Eval(ctx, f.args)
	if err != nil {
		return false, err
	}
	return !v.IsNull() && v.Bool(), nil
}

type RowMapFn struct {
	dynamicFn
}

func NewRowMapFn(fn *semantic.FunctionExpression, scope compiler.Scope) *RowMapFn {
	return &RowMapFn{
		dynamicFn: newDynamicFn(fn, scope),
	}
}

func (f *RowMapFn) Prepare(ctx context.Context, cols []flux.ColMeta) (*RowMapPreparedFn, error) {
	fn, err := f.prepare(ctx, cols, nil, false)
	if err != nil {
		return nil, err
	} else if k := fn.returnType().Nature(); k != semantic.Object {
		return nil, errors.Newf(codes.Invalid, "map function must return an object, got %s", k.String())
	}
	return &RowMapPreparedFn{
		rowFn: rowFn{preparedFn: fn},
	}, nil
}

type RowMapPreparedFn struct {
	rowFn
}

func (f *RowMapPreparedFn) Type() semantic.MonoType {
	return f.fn.Type()
}

func (f *RowMapPreparedFn) Eval(ctx context.Context, row int, cr flux.ColReader) (values.Object, error) {
	v, err := f.eval(ctx, row, cr, nil)
	if err != nil {
		return nil, err
	}
	return v.Object(), nil
}

type RowReduceFn struct {
	dynamicFn
}

func NewRowReduceFn(fn *semantic.FunctionExpression, scope compiler.Scope) *RowReduceFn {
	return &RowReduceFn{
		dynamicFn: newDynamicFn(fn, scope),
	}
}

func (f *RowReduceFn) Prepare(ctx context.Context, cols []flux.ColMeta, reducerType map[string]semantic.MonoType) (*RowReducePreparedFn, error) {
	fn, err := f.prepare(ctx, cols, reducerType, false)
	if err != nil {
		return nil, err
	}
	return &RowReducePreparedFn{
		rowFn: rowFn{preparedFn: fn},
	}, nil
}

type RowReducePreparedFn struct {
	rowFn
}

func (f *RowReducePreparedFn) Type() semantic.MonoType {
	return f.fn.Type()
}

func (f *RowReducePreparedFn) Eval(ctx context.Context, row int, cr flux.ColReader, extraParams map[string]values.Value) (values.Object, error) {
	v, err := f.eval(ctx, row, cr, extraParams)
	if err != nil {
		return nil, err
	}
	return v.Object(), nil
}

type RowJoinFn struct {
	dynamicFn
}

func NewRowJoinFn(fn *semantic.FunctionExpression, scope compiler.Scope) *RowJoinFn {
	return &RowJoinFn{
		dynamicFn: newDynamicFn(fn, scope),
	}
}

func (f *RowJoinFn) Prepare(ctx context.Context, cols []flux.ColMeta, rightType map[string]semantic.MonoType, vectorized bool) (*RowJoinPreparedFn, error) {
	fn, err := f.prepare(ctx, cols, rightType, vectorized)
	if err != nil {
		return nil, err
	}
	return &RowJoinPreparedFn{preparedFn: fn}, nil
}

func (f *RowJoinFn) Type() semantic.MonoType {
	return f.fn.TypeOf()
}

func (f *RowJoinFn) ReturnType() semantic.MonoType {
	return f.fn.Block.ReturnStatement().Argument.TypeOf()
}

type RowJoinPreparedFn struct {
	preparedFn
}

func (f *RowJoinPreparedFn) Eval(ctx context.Context, args values.Object) (values.Value, error) {
	return f.fn.Eval(ctx, args)
}
