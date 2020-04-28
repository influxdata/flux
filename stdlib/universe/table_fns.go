package universe

import (
	"context"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/compiler"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/lang"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
	"github.com/influxdata/flux/values/objects"
)

const (
	tableFindStreamArg           = "tables"
	tableFindFunctionArg         = "fn"
	tableFindFunctionGroupKeyArg = "key"
	getColumnTableArg            = "table"
	getColumnColumnArg           = "column"
	getRecordTableArg            = "table"
	getRecordIndexArg            = "idx"
)

func init() {
	runtime.RegisterPackageValue("universe", "tableFind", NewTableFindFunction())
	runtime.RegisterPackageValue("universe", "getColumn", NewGetColumnFunction())
	runtime.RegisterPackageValue("universe", "getRecord", NewGetRecordFunction())
	runtime.RegisterPackageValue("universe", "findColumn", NewFindColumnFunction())
	runtime.RegisterPackageValue("universe", "findRecord", NewFindRecordFunction())
}

func NewTableFindFunction() values.Value {
	return values.NewFunction("tableFind",
		runtime.MustLookupBuiltinType("universe", "tableFind"),
		tableFindCall,
		false)
}

func tableFindCall(ctx context.Context, args values.Object) (values.Value, error) {
	arguments := interpreter.NewArguments(args)
	v, err := arguments.GetRequired(tableFindStreamArg)
	if err != nil {
		return nil, err
	}
	to, ok := v.(*flux.TableObject)
	if !ok {
		return nil, errors.Newf(codes.Invalid, "expected TableObject but instead got %T", v)
	}

	var fn *execute.TablePredicateFn
	if call, err := arguments.GetRequiredFunction(tableFindFunctionArg); err != nil {
		return nil, errors.Newf(codes.Invalid, "missing argument: %s", tableFindFunctionArg)
	} else {
		predicate, err := interpreter.ResolveFunction(call)
		if err != nil {
			return nil, err
		}

		fn, err = execute.NewTablePredicateFn(predicate.Fn, compiler.ToScope(predicate.Scope))
		if err != nil {
			return nil, err
		}
	}

	t, err := tableFind(ctx, to, fn)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, errors.New(codes.NotFound, "no table found")
	}
	return t, nil
}

// Returns an error in the second return value, or the found table in the first
// return value, or nil to indicate that no table was found.
func tableFind(ctx context.Context, to *flux.TableObject, fn *execute.TablePredicateFn) (*objects.Table, error) {
	c := lang.TableObjectCompiler{
		Tables: to,
		Now:    time.Now(),
	}

	p, err := c.Compile(ctx)
	if err != nil {
		return nil, errors.Wrap(err, codes.Inherit, "error in table object compilation")
	}

	if !lang.HaveExecutionDependencies(ctx) {
		return nil, errors.New(codes.Invalid, "do not have an execution context for tableFind, if using the repl, try executing this code on the server using the InfluxDB API")
	}

	deps := lang.GetExecutionDependencies(ctx)
	if p, ok := p.(lang.LoggingProgram); ok {
		p.SetLogger(deps.Logger)
	}
	q, err := p.Start(ctx, deps.Allocator)
	if err != nil {
		return nil, errors.Wrap(err, codes.Inherit, "error in table object start")
	}

	var t *objects.Table
	var found bool
	for res := range q.Results() {
		if err := res.Tables().Do(func(tbl flux.Table) error {
			defer tbl.Done()
			if found {
				// the result is filled, you can skip other tables
				return nil
			}

			if err := fn.Prepare(tbl); err != nil {
				return err
			}

			var err error
			found, err = fn.Eval(ctx, tbl)
			if err != nil {
				return errors.Wrap(err, codes.Inherit, "failed to evaluate group key predicate function")
			}

			if found {
				t, err = objects.NewTable(tbl)
				if err != nil {
					return err
				}
			} else {
				// TODO(jsternberg): Remove the Do call when Done
				// is implemented for all table types.
				_ = tbl.Do(func(flux.ColReader) error { return nil })
			}
			return nil
		}); err != nil {
			return nil, err
		}
	}
	if !found {
		return nil, nil
	}
	return t, err
}

func NewGetColumnFunction() values.Value {
	return values.NewFunction("getColumn",
		runtime.MustLookupBuiltinType("universe", "getColumn"),
		getColumnCall,
		false)
}

func getColumnCall(ctx context.Context, args values.Object) (values.Value, error) {
	arguments := interpreter.NewArguments(args)
	var tbl flux.Table
	if v, err := arguments.GetRequired(getColumnTableArg); err != nil {
		return nil, err
	} else if v.Type() != objects.TableMonoType {
		return nil, errors.Newf(codes.Invalid, "unexpected type for %s: want %v, got %v", getColumnTableArg, objects.TableMonoType, v.Type())
	} else {
		tbl = v.(*objects.Table).Table()
	}

	col, err := arguments.GetRequiredString(getColumnColumnArg)
	if err != nil {
		return nil, err
	}

	idx := execute.ColIdx(col, tbl.Cols())
	if idx < 0 {
		return nil, errors.Newf(codes.Invalid, "cannot find column %s", col)
	}
	var a values.Array
	if err = tbl.Do(func(cr flux.ColReader) error {
		a = arrayFromColumn(idx, cr)
		return nil
	}); err != nil {
		return nil, err
	}
	return a, nil
}

func arrayFromColumn(idx int, cr flux.ColReader) values.Array {
	typ := cr.Cols()[idx].Type
	vsSlice := make([]values.Value, 0, cr.Len())
	for i := 0; i < cr.Len(); i++ {
		switch typ {
		case flux.TString:
			if vs := cr.Strings(idx); vs.IsValid(i) {
				vsSlice = append(vsSlice, values.New(vs.ValueString(i)))
			} else {
				vsSlice = append(vsSlice, values.NewNull(semantic.BasicString))
			}
		case flux.TInt:
			if vs := cr.Ints(idx); vs.IsValid(i) {
				vsSlice = append(vsSlice, values.New(vs.Value(i)))
			} else {
				vsSlice = append(vsSlice, values.NewNull(semantic.BasicInt))
			}
		case flux.TUInt:
			if vs := cr.UInts(idx); vs.IsValid(i) {
				vsSlice = append(vsSlice, values.New(vs.Value(i)))
			} else {
				vsSlice = append(vsSlice, values.NewNull(semantic.BasicUint))
			}
		case flux.TFloat:
			if vs := cr.Floats(idx); vs.IsValid(i) {
				vsSlice = append(vsSlice, values.New(vs.Value(i)))
			} else {
				vsSlice = append(vsSlice, values.NewNull(semantic.BasicFloat))
			}
		case flux.TBool:
			if vs := cr.Bools(idx); vs.IsValid(i) {
				vsSlice = append(vsSlice, values.New(vs.Value(i)))
			} else {
				vsSlice = append(vsSlice, values.NewNull(semantic.BasicBool))
			}
		case flux.TTime:
			if vs := cr.Times(idx); vs.IsValid(i) {
				vsSlice = append(vsSlice, values.New(values.Time(vs.Value(i))))
			} else {
				vsSlice = append(vsSlice, values.NewNull(semantic.BasicTime))
			}
		default:
			execute.PanicUnknownType(typ)
		}
	}
	return values.NewArrayWithBacking(semantic.NewArrayType(flux.SemanticType(typ)), vsSlice)
}

func NewGetRecordFunction() values.Value {
	return values.NewFunction("getRecord",
		runtime.MustLookupBuiltinType("universe", "getRecord"),
		getRecordCall,
		false)
}

func getRecordCall(ctx context.Context, args values.Object) (values.Value, error) {
	arguments := interpreter.NewArguments(args)
	var tbl flux.Table
	if v, err := arguments.GetRequired(getRecordTableArg); err != nil {
		return nil, err
	} else if v.Type() != objects.TableMonoType {
		return nil, errors.Newf(codes.Invalid, "unexpected type for %s: want %v, got %v", getRecordTableArg, objects.TableMonoType, v.Type())
	} else {
		tbl = v.(*objects.Table).Table()
	}

	rowIdx, err := arguments.GetRequiredInt(getRecordIndexArg)
	if err != nil {
		return nil, err
	}

	var r values.Object
	if err = tbl.Do(func(cr flux.ColReader) error {
		if rowIdx < 0 || int(rowIdx) >= cr.Len() {
			return errors.Newf(codes.OutOfRange, "index out of bounds: %d", rowIdx)
		}
		r = objectFromRow(int(rowIdx), cr)
		return nil
	}); err != nil {
		return nil, err
	}
	return r, nil
}

func objectFromRow(idx int, cr flux.ColReader) values.Object {
	vsMap := make(map[string]values.Value, len(cr.Cols()))
	for j, c := range cr.Cols() {
		var v values.Value
		switch c.Type {
		case flux.TString:
			if vs := cr.Strings(j); vs.IsValid(idx) {
				v = values.New(vs.ValueString(idx))
			} else {
				v = values.NewNull(semantic.BasicString)
			}
		case flux.TInt:
			if vs := cr.Ints(j); vs.IsValid(idx) {
				v = values.New(vs.Value(idx))
			} else {
				v = values.NewNull(semantic.BasicInt)
			}
		case flux.TUInt:
			if vs := cr.UInts(j); vs.IsValid(idx) {
				v = values.New(vs.Value(idx))
			} else {
				v = values.NewNull(semantic.BasicUint)
			}
		case flux.TFloat:
			if vs := cr.Floats(j); vs.IsValid(idx) {
				v = values.New(vs.Value(idx))
			} else {
				v = values.NewNull(semantic.BasicFloat)
			}
		case flux.TBool:
			if vs := cr.Bools(j); vs.IsValid(idx) {
				v = values.New(vs.Value(idx))
			} else {
				v = values.NewNull(semantic.BasicBool)
			}
		case flux.TTime:
			if vs := cr.Times(j); vs.IsValid(idx) {
				v = values.New(values.Time(vs.Value(idx)))
			} else {
				v = values.NewNull(semantic.BasicTime)
			}
		default:
			execute.PanicUnknownType(c.Type)
		}
		vsMap[c.Label] = v
	}
	return values.NewObjectWithValues(vsMap)
}

func NewFindColumnFunction() values.Value {
	return values.NewFunction("findColumn",
		runtime.MustLookupBuiltinType("universe", "findColumn"),
		findColumnCall,
		false)
}

func findColumnCall(ctx context.Context, args values.Object) (values.Value, error) {
	arguments := interpreter.NewArguments(args)
	v, err := arguments.GetRequired(tableFindStreamArg)
	if err != nil {
		return nil, err
	}
	to, ok := v.(*flux.TableObject)
	if !ok {
		return nil, errors.Newf(codes.Invalid, "expected TableObject but instead got %T", v)
	}

	var fn *execute.TablePredicateFn
	if call, err := arguments.GetRequiredFunction(tableFindFunctionArg); err != nil {
		return nil, errors.Newf(codes.Invalid, "missing argument: %s", tableFindFunctionArg)
	} else {
		predicate, err := interpreter.ResolveFunction(call)
		if err != nil {
			return nil, err
		}

		fn, err = execute.NewTablePredicateFn(predicate.Fn, compiler.ToScope(predicate.Scope))
		if err != nil {
			return nil, err
		}
	}

	col, err := arguments.GetRequiredString(getColumnColumnArg)
	if err != nil {
		return nil, err
	}

	tv, err := tableFind(ctx, to, fn)
	if err != nil {
		return nil, err
	}

	if tv == nil {
		return emptyArray(), nil
	}

	tbl := tv.Table()

	idx := execute.ColIdx(col, tbl.Cols())
	if idx < 0 {
		return emptyArray(), nil
	}
	var a values.Array
	if err = tbl.Do(func(cr flux.ColReader) error {
		a = arrayFromColumn(idx, cr)
		return nil
	}); err != nil {
		return nil, err
	}
	return a, nil
}

func emptyArray() values.Array {
	vsSlice := make([]values.Value, 0)
	return values.NewArrayWithBacking(semantic.NewArrayType(semantic.BasicString), vsSlice)
}

func NewFindRecordFunction() values.Value {
	return values.NewFunction("findRecord",
		runtime.MustLookupBuiltinType("universe", "findRecord"),
		findRecordCall,
		false)
}

func findRecordCall(ctx context.Context, args values.Object) (values.Value, error) {
	arguments := interpreter.NewArguments(args)
	v, err := arguments.GetRequired(tableFindStreamArg)
	if err != nil {
		return nil, err
	}
	to, ok := v.(*flux.TableObject)
	if !ok {
		return nil, errors.Newf(codes.Invalid, "expected TableObject but instead got %T", v)
	}

	var fn *execute.TablePredicateFn
	if call, err := arguments.GetRequiredFunction(tableFindFunctionArg); err != nil {
		return nil, errors.Newf(codes.Invalid, "missing argument: %s", tableFindFunctionArg)
	} else {
		predicate, err := interpreter.ResolveFunction(call)
		if err != nil {
			return nil, err
		}

		fn, err = execute.NewTablePredicateFn(predicate.Fn, compiler.ToScope(predicate.Scope))
		if err != nil {
			return nil, err
		}
	}

	rowIdx, err := arguments.GetRequiredInt(getRecordIndexArg)
	if err != nil {
		return nil, err
	}

	tv, err := tableFind(ctx, to, fn)
	if err != nil {
		return nil, err
	}

	if tv == nil {
		return emptyObject(), nil
	}

	tbl := tv.Table()

	var r values.Object
	if err = tbl.Do(func(cr flux.ColReader) error {
		if rowIdx < 0 || int(rowIdx) >= cr.Len() {
			r = emptyObject()
			return nil
		}
		r = objectFromRow(int(rowIdx), cr)
		return nil
	}); err != nil {
		return nil, err
	}
	return r, nil
}

func emptyObject() values.Object {
	vsMap := make(map[string]values.Value)
	return values.NewObjectWithValues(vsMap)
}
