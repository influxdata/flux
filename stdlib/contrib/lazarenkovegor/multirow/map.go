package multirow

import (
	"context"
	"fmt"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/compiler"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/table"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/internal/execute/dataset"
	"github.com/influxdata/flux/internal/fbsemantic"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

const (
	MapName = "map"
	MapKind = pkgpath + "." + MapName
)

func init() {
	runtime.RegisterPackageValue(pkgpath, MapName, flux.MustValue(flux.FunctionValue(
		MapName,
		createMapOpSpec,
		runtime.MustLookupBuiltinType(pkgpath, MapName),
	)))

	plan.RegisterProcedureSpec(MapKind, createMapProcSpec, MapKind)
	execute.RegisterTransformation(MapKind, createMapTransformation)
}

type WindowFilter struct {
	TimeColIndex int
	Do           func(filter *WindowFilter, colReader flux.ColReader, currentIndex, lastIndex int) (from int, to int)
}

type MapOpSpec struct {
	Fn               interpreter.ResolvedFunction
	FromArrayFn      values.Function
	WindowFilter     *WindowFilter
	HasRowParam      bool
	HasWindowParam   bool
	HasIndexParam    bool
	HasCountParam    bool
	HasPreviousParam bool
	Column           string
	InitValue        values.Object
	VirtualColumns   []string
}

type MapPlan struct {
	plan.DefaultCost
	MapOpSpec
}

func (s *MapPlan) Kind() plan.ProcedureKind {
	return plan.ProcedureKind(s.MapOpSpec.Kind())
}

func (s *MapPlan) Copy() plan.ProcedureSpec {
	return &MapPlan{MapOpSpec: s.MapOpSpec}
}

func (s *MapOpSpec) Kind() flux.OperationKind {
	return MapKind
}

func makeWindowFilter(left values.Value, right values.Value) (*WindowFilter, error) {
	getFrom := func(filter *WindowFilter, reader flux.ColReader, curIndex int, lastIndex int) int { return curIndex }
	getTo := getFrom
	res := &WindowFilter{TimeColIndex: -2}
	if left != values.Null {
		if bt, err := left.Type().Basic(); err != nil {
			return nil, errors.Newf(codes.Invalid, "invalid left param: %s", err.Error())
		} else {
			switch bt {
			case fbsemantic.TypeInt:
				val := int(left.Int())
				getFrom = func(filter *WindowFilter, reader flux.ColReader, curIndex int, lastIndex int) int {
					res := curIndex - val
					if res < 0 {
						res = 0
					}
					return res
				}
			case fbsemantic.TypeDuration:
				val := left.Duration()
				res.TimeColIndex = -1
				getFrom = func(filter *WindowFilter, reader flux.ColReader, curIndex int, lastIndex int) int {
					from := curIndex
					minTime := int64(values.Time(reader.Times(filter.TimeColIndex).Value(curIndex)).Add(val.Mul(-1)))
					for from > 0 {
						if reader.Times(filter.TimeColIndex).Value(from-1) < minTime {
							break
						}
						from--
					}
					return from
				}
			default:
				return nil, errors.Newf(codes.Invalid, "invalid left param type %T", bt)
			}
		}
	}

	if right != values.Null {
		if bt, err := right.Type().Basic(); err != nil {
			return nil, errors.Newf(codes.Invalid, "invalid right param: %s", err.Error())
		} else {
			switch bt {
			case fbsemantic.TypeInt:
				val := int(right.Int())
				getTo = func(filter *WindowFilter, reader flux.ColReader, curIndex int, lastIndex int) int {
					res := curIndex + val
					if res > lastIndex {
						res = lastIndex
					}
					return res
				}
			case fbsemantic.TypeDuration:
				val := right.Duration()
				res.TimeColIndex = -1
				getTo = func(filter *WindowFilter, reader flux.ColReader, curIndex int, lastIndex int) int {
					from := curIndex
					maxTime := int64(values.Time(reader.Times(filter.TimeColIndex).Value(curIndex)).Add(val))
					for from < lastIndex {
						if reader.Times(filter.TimeColIndex).Value(from+1) > maxTime {
							break
						}
						from++
					}
					return from
				}
			default:
				return nil, errors.Newf(codes.Invalid, "invalid left param type %T", bt)
			}
		}
	}
	res.Do = func(filter *WindowFilter, colReader flux.ColReader, currentIndex, lastIndex int) (from int, to int) {
		return getFrom(res, colReader, currentIndex, lastIndex), getTo(res, colReader, currentIndex, lastIndex)
	}
	return res, nil
}

func createMapOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(MapOpSpec)

	if fn, err := args.GetRequiredFunction("fn"); err != nil {
		return nil, err
	} else if fn, err := interpreter.ResolveFunction(fn); err != nil {
		return nil, err
	} else {
		spec.Fn = fn
		if fn.Fn.Parameters != nil {
			for _, v := range fn.Fn.Parameters.List {
				switch v.Key.Name.Name() {
				case "row":
					spec.HasRowParam = true
				case "window":
					spec.HasWindowParam = true
				case "index":
					spec.HasIndexParam = true
				case "count":
					spec.HasCountParam = true
				case "previous":
					spec.HasPreviousParam = true
				}
			}
		}
	}
	if c, f, err := args.GetString("column"); err != nil {
		return nil, err
	} else if f {
		spec.Column = c
	} else {
		spec.Column = "_value"
	}

	if o, f, err := args.GetObject("init"); err != nil {
		return nil, err
	} else if f {
		spec.InitValue = o
	} else {
		spec.InitValue = values.NewObjectWithValues(nil)
	}

	pkg, err := runtime.StdLib().ImportPackageObject("array")
	if err != nil {
		return nil, err
	}
	if from, ok := pkg.Get("from"); ok {
		spec.FromArrayFn = from.Function()
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("can't resolve function array.from")
	}

	var left values.Value = values.Null
	var right values.Value = values.Null
	if val, found := args.Get("left"); found {
		left = val
	}
	if val, found := args.Get("right"); found {
		right = val
	}
	spec.WindowFilter, err = makeWindowFilter(left, right)

	if n, found, err := args.GetArrayAllowEmpty("virtual", semantic.String); err != nil {
		return nil, err
	} else if found {
		l := n.Len()
		spec.VirtualColumns = make([]string, l)
		for i := 0; i < l; i++ {
			spec.VirtualColumns[i] = n.Get(i).Str()
		}
	}

	return spec, nil

}

func createMapProcSpec(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*MapOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Invalid, "invalid spec type %T", qs)
	}

	return &MapPlan{MapOpSpec: *spec}, nil
}

type mapTransformation struct {
	execute.ExecutionNode
	ds    execute.Dataset
	cache table.BuilderCache
	spec  *MapPlan
	ctx   context.Context
	mem   *memory.Allocator
}

func (s *mapTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return s.ds.RetractTable(key)
}

func (s *mapTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	rowTp, err := ColsMonoType(tbl.Cols())
	if err != nil {
		return err
	}
	arrayOfRowTp := semantic.NewArrayType(rowTp)

	pr := make([]semantic.PropertyType, 0, 4)
	if s.spec.HasWindowParam {
		pr = append(pr, semantic.PropertyType{Key: []byte("window"), Value: arrayOfRowTp})
	}
	if s.spec.HasRowParam {
		pr = append(pr, semantic.PropertyType{Key: []byte("row"), Value: rowTp})
	}
	if s.spec.HasCountParam {
		pr = append(pr, semantic.PropertyType{Key: []byte("count"), Value: semantic.BasicInt})
	}
	if s.spec.HasIndexParam {
		pr = append(pr, semantic.PropertyType{Key: []byte("index"), Value: semantic.BasicInt})
	}
	if s.spec.HasPreviousParam {
		pr = append(pr, semantic.PropertyType{Key: []byte("previous"), Value: semantic.NewObjectType(nil)})
	}

	useFnArgsType := semantic.NewObjectType(pr)
	fromArrayArgsType := semantic.NewObjectType([]semantic.PropertyType{{Key: []byte("rows"), Value: arrayOfRowTp}})

	scope := compiler.ToScope(s.spec.Fn.Scope)
	userFunction, err := compiler.Compile(scope, s.spec.Fn.Fn, useFnArgsType)
	if err != nil {
		return err
	}

	tb := NewTableBuilder(tbl.Key(), s.mem, s.spec.Column, s.spec.VirtualColumns)
	defer tb.Release()

	if err := tbl.Do(func(reader flux.ColReader) error {
		l := reader.Len()
		if l == 0 {
			return nil
		}
		lastIndex := l - 1
		var previous values.Object
		if s.spec.WindowFilter.TimeColIndex == -1 {
			s.spec.WindowFilter.TimeColIndex = execute.ColIdx("_time", reader.Cols())
			if s.spec.WindowFilter.TimeColIndex == -1 {
				return errors.New(codes.Invalid, "column _time not fount, but needs for make window")
			}
		}

		for curRowId := 0; curRowId < l; curRowId++ {
			var row values.Object
			if s.spec.HasRowParam {
				row = MakeRowObject(nil, reader, curRowId)
			}

			var rows values.Array
			if s.spec.HasWindowParam {
				from, to := s.spec.WindowFilter.Do(s.spec.WindowFilter, reader, curRowId, lastIndex)
				rows = s.makeWindowRows(arrayOfRowTp, rowTp, reader, from, to)
			}

			if err := s.doUserFunction(useFnArgsType, row, rows, fromArrayArgsType, userFunction, tb, curRowId, l, &previous); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}

	t, err := tb.Table()
	if err != nil {
		return err
	}

	return t.Do(func(reader flux.ColReader) error {
		ab, _ := table.GetBufferedBuilder(reader.Key(), &s.cache)
		return ab.AppendBuffer(reader)
	})

}

func (s *mapTransformation) makeWindowRows(arrayOfRowTp semantic.MonoType, rowTp semantic.MonoType, reader flux.ColReader, from, to int) values.Array {
	rows := values.NewArray(arrayOfRowTp)

	for ; from <= to; from++ {
		rows.Append(MakeRowObject(&rowTp, reader, from))
	}

	return rows
}

func (s *mapTransformation) doUserFunction(useFnArgsType semantic.MonoType, row values.Object, rows values.Array, fromArrayArgsType semantic.MonoType,
	userFunction compiler.Func, out *TableBuilder, index int, count int, previous *values.Object) error {
	args := values.NewObject(useFnArgsType)

	if s.spec.HasWindowParam {
		args2 := values.NewObject(fromArrayArgsType)
		args2.Set("rows", rows)
		stream, err := s.spec.FromArrayFn.Call(s.ctx, args2)
		if err != nil {
			return err
		}
		args.Set("window", stream)
	}

	if row != nil {
		args.Set("row", row)
	}
	if s.spec.HasIndexParam {
		args.Set("index", values.NewInt(int64(index)))
	}
	if s.spec.HasPreviousParam {
		if *previous == nil {
			if s.spec.InitValue == nil {
				*previous = values.NewObjectWithValues(nil)
			} else {
				*previous = s.spec.InitValue
			}
		}

		args.Set("previous", *previous)
	}

	if s.spec.HasCountParam {
		args.Set("count", values.NewInt(int64(count)))
	}
	res, err := userFunction.Eval(s.ctx, args)
	if err != nil {
		return err
	}

	obj, err := out.AppendRows(s.ctx, res, s.spec.HasPreviousParam)
	if err != nil {
		return err
	}
	if s.spec.HasPreviousParam {
		*previous = obj
	}
	return nil
}

func (s *mapTransformation) UpdateWatermark(id execute.DatasetID, t execute.Time) error {
	return s.ds.UpdateWatermark(t)
}

func (s *mapTransformation) UpdateProcessingTime(id execute.DatasetID, t execute.Time) error {
	return s.ds.UpdateProcessingTime(t)
}

func (s *mapTransformation) Finish(id execute.DatasetID, err error) {
	s.ds.Finish(err)
}

func createMapTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*MapPlan)
	if !ok {
		return nil, nil, errors.Newf(codes.Invalid, "invalid spec type %T", spec)
	}
	w := &mapTransformation{
		ctx:  a.Context(),
		mem:  a.Allocator(),
		spec: s,
		cache: table.BuilderCache{
			New: func(key flux.GroupKey) table.Builder {
				return table.NewBufferedBuilder(key, a.Allocator())
			},
		},
	}
	w.ds = dataset.New(id, &w.cache)
	return w, w.ds, nil
}
