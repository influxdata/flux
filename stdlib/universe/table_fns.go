package universe

import (
	"context"
	"fmt"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/lang"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
	"github.com/influxdata/flux/values/objects"
	"github.com/pkg/errors"
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
	flux.RegisterPackageValue("universe", "tableFind", NewTableFindFunction())
	flux.RegisterPackageValue("universe", "getColumn", NewGetColumnFunction())
	flux.RegisterPackageValue("universe", "getRecord", NewGetRecordFunction())
}

func NewTableFindFunction() values.Value {
	return values.NewFunction("tableFind",
		semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
			Parameters: map[string]semantic.PolyType{
				tableFindStreamArg: flux.TableObjectType,
				tableFindFunctionArg: semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
					Parameters: map[string]semantic.PolyType{
						tableFindFunctionGroupKeyArg: semantic.Tvar(1),
					},
					Required: semantic.LabelSet{tableFindFunctionGroupKeyArg},
					Return:   semantic.Bool,
				}),
			},
			Required:     semantic.LabelSet{tableFindStreamArg, tableFindFunctionArg},
			PipeArgument: tableFindStreamArg,
			Return:       objects.TableType,
		}),
		tableFindCall,
		false)
}

func tableFindCall(args values.Object) (values.Value, error) {
	arguments := interpreter.NewArguments(args)
	var to *flux.TableObject
	if v, err := arguments.GetRequired(tableFindStreamArg); err != nil {
		return nil, err
	} else if v.Type() != flux.TableObjectMonoType {
		return nil, fmt.Errorf("unexpected type for %v: want %v, got %v", tableFindStreamArg, "table stream", v.Type())
	} else {
		to = v.(*flux.TableObject)
	}

	fn, err := arguments.GetRequiredFunction(tableFindFunctionArg)
	if err != nil {
		return nil, fmt.Errorf("missing argument: %s", tableFindFunctionArg)
	}

	c := lang.TableObjectCompiler{
		Tables: to,
		Now:    time.Now(),
	}

	p, err := c.Compile(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "error in table object compilation")
	}

	q, err := p.Start(context.Background(), &memory.Allocator{})
	if err != nil {
		return nil, errors.Wrap(err, "error in table object start")
	}

	var t *objects.Table
	var found bool
	for res := range q.Results() {
		if err := res.Tables().Do(func(tbl flux.Table) error {
			if found {
				// the result is filled, you can skip other tables
				return nil
			}
			gk := objectFromGroupKey(tbl.Key())
			pass, err := fn.Call(values.NewObjectWithValues(map[string]values.Value{tableFindFunctionGroupKeyArg: gk}))
			if err != nil {
				return errors.Wrap(err, "failed to evaluate group key predicate function")
			}
			found = pass.Bool()
			if found {
				// We need to copy the table in memory and increase its refCount in order to make
				// subsequent calls to getRecord/Column idempotent. If we don't do it, then it would be
				// consumed by calls to `Do`, and subsequent calls to getRecord/Column would find
				// an empty table.
				// TODO(aff): Note that, for now, it is not enough to `tbl.RefCount(1)`, because we cannot rely on its
				//  implementation. When a table comes from `csv.from()` it is a `csv.tableDecoder` that
				//  does nothing when `RefCount` is called.
				if tbl, err := execute.CopyTable(tbl, &memory.Allocator{}); err != nil {
					return err
				} else {
					t = objects.NewTable(tbl)
				}
			}
			return nil
		}); err != nil {
			return nil, err
		}
	}
	if !found {
		return nil, fmt.Errorf("no table found")
	}
	return t, nil
}

func objectFromGroupKey(gk flux.GroupKey) values.Object {
	vsMap := make(map[string]values.Value, len(gk.Cols()))
	for j, c := range gk.Cols() {
		vsMap[c.Label] = gk.Value(j)
	}
	return values.NewObjectWithValues(vsMap)
}

func NewGetColumnFunction() values.Value {
	return values.NewFunction("getColumn",
		semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
			Parameters: map[string]semantic.PolyType{
				getColumnTableArg:  objects.TableType,
				getColumnColumnArg: semantic.String,
			},
			Required:     semantic.LabelSet{getColumnTableArg, getColumnColumnArg},
			PipeArgument: getColumnTableArg,
			Return:       semantic.Array,
		}),
		getColumnCall,
		false)
}

func getColumnCall(args values.Object) (values.Value, error) {
	arguments := interpreter.NewArguments(args)
	var tbl flux.Table
	if v, err := arguments.GetRequired(getColumnTableArg); err != nil {
		return nil, err
	} else if v.Type() != objects.TableMonoType {
		return nil, fmt.Errorf("unexpected type for %s: want %v, got %v", getColumnTableArg, objects.TableMonoType, v.Type())
	} else {
		tbl = v.(*objects.Table).Table
	}

	col, err := arguments.GetRequiredString(getColumnColumnArg)
	if err != nil {
		return nil, err
	}

	idx := execute.ColIdx(col, tbl.Cols())
	if idx < 0 {
		return nil, fmt.Errorf("cannot find column %s", col)
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
				vsSlice = append(vsSlice, values.NewNull(semantic.String))
			}
		case flux.TInt:
			if vs := cr.Ints(idx); vs.IsValid(i) {
				vsSlice = append(vsSlice, values.New(vs.Value(i)))
			} else {
				vsSlice = append(vsSlice, values.NewNull(semantic.Int))
			}
		case flux.TUInt:
			if vs := cr.UInts(idx); vs.IsValid(i) {
				vsSlice = append(vsSlice, values.New(vs.Value(i)))
			} else {
				vsSlice = append(vsSlice, values.NewNull(semantic.UInt))
			}
		case flux.TFloat:
			if vs := cr.Floats(idx); vs.IsValid(i) {
				vsSlice = append(vsSlice, values.New(vs.Value(i)))
			} else {
				vsSlice = append(vsSlice, values.NewNull(semantic.Float))
			}
		case flux.TBool:
			if vs := cr.Bools(idx); vs.IsValid(i) {
				vsSlice = append(vsSlice, values.New(vs.Value(i)))
			} else {
				vsSlice = append(vsSlice, values.NewNull(semantic.Bool))
			}
		case flux.TTime:
			if vs := cr.Times(idx); vs.IsValid(i) {
				vsSlice = append(vsSlice, values.New(values.Time(vs.Value(i))))
			} else {
				vsSlice = append(vsSlice, values.NewNull(semantic.Time))
			}
		default:
			execute.PanicUnknownType(typ)
		}
	}
	return values.NewArrayWithBacking(flux.SemanticType(typ), vsSlice)
}

func NewGetRecordFunction() values.Value {
	return values.NewFunction("getRecord",
		semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
			Parameters: map[string]semantic.PolyType{
				getRecordTableArg: objects.TableType,
				getRecordIndexArg: semantic.Int,
			},
			Required:     semantic.LabelSet{getRecordTableArg, getRecordIndexArg},
			PipeArgument: getRecordTableArg,
			// TODO(affo): this return type should be parameterized by the input types. It cannot be TVar,
			//  because if there is a TVar on the right, there must be one on the left.
			Return: semantic.Object,
		}),
		getRecordCall,
		false)
}

func getRecordCall(args values.Object) (values.Value, error) {
	arguments := interpreter.NewArguments(args)
	var tbl flux.Table
	if v, err := arguments.GetRequired(getRecordTableArg); err != nil {
		return nil, err
	} else if v.Type() != objects.TableMonoType {
		return nil, fmt.Errorf("unexpected type for %s: want %v, got %v", getRecordTableArg, objects.TableMonoType, v.Type())
	} else {
		tbl = v.(*objects.Table).Table
	}

	rowIdx, err := arguments.GetRequiredInt(getRecordIndexArg)
	if err != nil {
		return nil, err
	}

	var r values.Object
	if err = tbl.Do(func(cr flux.ColReader) error {
		if rowIdx < 0 || int(rowIdx) >= cr.Len() {
			return fmt.Errorf("index out of bounds: %d", rowIdx)
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
				v = values.NewNull(semantic.String)
			}
		case flux.TInt:
			if vs := cr.Ints(j); vs.IsValid(idx) {
				v = values.New(vs.Value(idx))
			} else {
				v = values.NewNull(semantic.Int)
			}
		case flux.TUInt:
			if vs := cr.UInts(j); vs.IsValid(idx) {
				v = values.New(vs.Value(idx))
			} else {
				v = values.NewNull(semantic.UInt)
			}
		case flux.TFloat:
			if vs := cr.Floats(j); vs.IsValid(idx) {
				v = values.New(vs.Value(idx))
			} else {
				v = values.NewNull(semantic.Float)
			}
		case flux.TBool:
			if vs := cr.Bools(j); vs.IsValid(idx) {
				v = values.New(vs.Value(idx))
			} else {
				v = values.NewNull(semantic.Bool)
			}
		case flux.TTime:
			if vs := cr.Times(j); vs.IsValid(idx) {
				v = values.New(values.Time(vs.Value(idx)))
			} else {
				v = values.NewNull(semantic.Time)
			}
		default:
			execute.PanicUnknownType(c.Type)
		}
		vsMap[c.Label] = v
	}
	return values.NewObjectWithValues(vsMap)
}
