package compiler_test

import (
	"context"
	"errors"
	"testing"

	"github.com/apache/arrow/go/arrow/bitutil"
	arrowmem "github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/compiler"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/table"
	"github.com/influxdata/flux/internal/arrowutil"
	valcompiler "github.com/influxdata/flux/internal/compiler"
	itable "github.com/influxdata/flux/internal/execute/table"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

var (
	n                  = 1000
	mem                = &memory.Allocator{}
	vectorizedArgsType semantic.MonoType
	valueArgsType      semantic.MonoType

	mapFnExpr               = mustGetFnExpr(`(r) => ({result: r.a + r.b})`)
	baselineCompiledMapFn   *execute.RowMapPreparedFn
	vectorizedCompiledMapFn compiler.Func

	filterFnExpr               = mustGetFnExpr(`(r) => r.a + r.b > 500.0`)
	valueFilterFnExpr          = mustGetFnExpr(`(a, b) => a + b > 500.0`)
	baselineCompiledFilterFn   *execute.RowPredicatePreparedFn
	vectorizedCompiledFilterFn compiler.Func
	valueCompiledFilterFn      valcompiler.Func
)

func init() {
	var err error

	inputCols := []flux.ColMeta{
		{Label: "a", Type: flux.TFloat},
		{Label: "b", Type: flux.TFloat},
	}

	mapFn := execute.NewRowMapFn(mapFnExpr, nil)
	baselineCompiledMapFn, err = mapFn.Prepare(inputCols)
	if err != nil {
		panic(err)
	}

	vrt := recordType(semantic.NewArrayType(semantic.BasicFloat))
	vectorizedArgsType = argsType(vrt)
	vectorizedCompiledMapFn, err = compiler.Compile(nil, mapFnExpr, vectorizedArgsType)
	if err != nil {
		panic(err)
	}

	predFn := execute.NewRowPredicateFn(filterFnExpr, nil)
	baselineCompiledFilterFn, err = predFn.Prepare(inputCols)
	if err != nil {
		panic(err)
	}

	vectorizedCompiledFilterFn, err = compiler.Compile(nil, filterFnExpr, vectorizedArgsType)
	if err != nil {
		panic(err)
	}

	// valueArgsType = semantic.NewObjectType(
	// 	[]semantic.PropertyType{
	// 		{Key: []byte("r"), Value: semantic.NewObjectType(
	// 			[]semantic.PropertyType{
	// 				{Key: []byte("a"), Value: semantic.BasicFloat},
	// 				{Key: []byte("b"), Value: semantic.BasicFloat},
	// 			},
	// 		)},
	// 	},
	// )
	valueArgsType = semantic.NewObjectType(
		[]semantic.PropertyType{
			{Key: []byte("a"), Value: semantic.BasicFloat},
			{Key: []byte("b"), Value: semantic.BasicFloat},
		},
	)
	valueCompiledFilterFn, err = valcompiler.Compile(valueFilterFnExpr, valueArgsType)
	if err != nil {
		panic(err)
	}
}

func TestVectorize(t *testing.T) {
	inputTable := getInputTable(n)
	outputTable := getOutputTable(n)
	t.Run("baselineMap", func(t *testing.T) {
		gt := baselineMap(t, inputTable.Copy())
		got := table.Iterator{gt}
		want := table.Iterator{outputTable.Copy()}
		if diff := table.Diff(want, got); diff != "" {
			t.Errorf("tables were different: %s", diff)
		}
	})
	t.Run("vectorizedMap", func(t *testing.T) {
		gt := vectorizedMap(t, inputTable.Copy())
		got := table.Iterator{gt}
		want := table.Iterator{outputTable.Copy()}
		if diff := table.Diff(want, got); diff != "" {
			t.Errorf("tables were different: %s", diff)
		}
	})
	t.Run("baselineFilter", func(t *testing.T) {
		bitsets := baselineFilter(t, inputTable.Copy())
		var numOnes int
		for _, bs := range bitsets {
			numOnes += bitutil.CountSetBits(bs.Buf(), 0, bs.Len())
		}
		if want, got := 749, numOnes; want != got {
			t.Errorf("Did not get expected result, -want/+got: -%v/+%v", want, got)
		}
	})
	t.Run("vectorizedfilter", func(t *testing.T) {
		arrs := vectorizedFilter(t, inputTable.Copy())
		var numOnes int
		for _, arr := range arrs {
			for i := 0; i < arr.Len(); i++ {
				if arr.Get(i).Bool() {
					numOnes++
				}
			}
		}
		if want, got := 749, numOnes; want != got {
			t.Errorf("Did not get expected result, -want/+got: -%v/+%v", want, got)
		}
	})
}

func BenchmarkVectorize(b *testing.B) {
	inputTable := getInputTable(n)
	b.Run("baseline", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			result := baselineMap(b, inputTable.Copy())
			if result == nil {
				b.Fatal("got nil result")
			}
		}
	})
	b.Run("vectorized", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			result := vectorizedMap(b, inputTable.Copy())
			if result == nil {
				b.Fatal("got nil result")
			}
		}
	})
	b.Run("baselineFilter", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			bitsets := baselineFilter(b, inputTable.Copy())
			if len(bitsets) < 1 {
				b.Fatal("got no bitsets")
			}
		}
	})
	b.Run("vectorizedFilter", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			arrs := vectorizedFilter(b, inputTable.Copy())
			if len(arrs) < 1 {
				b.Fatal("got no arrs")
			}
		}
	})
	b.Run("valueFilter", func(b *testing.B) {
		// args := valcompiler.NewObject(valueArgsType)
		//
		// var recordType semantic.MonoType
		// if prop, err := valueArgsType.RecordProperty(0); err != nil {
		// 	b.Fatal(err)
		// } else {
		// 	recordType, err = prop.TypeOf()
		// 	if err != nil {
		// 		b.Fatal(err)
		// 	}
		// }
		// record := valcompiler.NewObject(recordType)
		// args.Set("r", record)
		args := valcompiler.NewObject(valueArgsType)

		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			bitsets := valueFilter(b, inputTable.Copy(), args)
			if len(bitsets) < 1 {
				b.Fatal("got no bitsets")
			}
		}
	})
}

func mustGetFnExpr(src string) *semantic.FunctionExpression {
	sg, err := runtime.AnalyzeSource(src)
	if err != nil {
		panic(err)
	}
	// The function expression
	fn := sg.Files[0].Body[0].(*semantic.ExpressionStatement).Expression.(*semantic.FunctionExpression)
	return fn
}

func recordType(fieldType semantic.MonoType) semantic.MonoType {
	properties := make([]semantic.PropertyType, 2)
	properties[0] = semantic.PropertyType{
		Key:   []byte("a"),
		Value: fieldType,
	}
	properties[1] = semantic.PropertyType{
		Key:   []byte("b"),
		Value: fieldType,
	}
	return semantic.NewObjectType(properties)

}

func argsType(recordType semantic.MonoType) semantic.MonoType {
	props := []semantic.PropertyType{
		{Key: []byte("r"), Value: recordType},
	}
	return semantic.NewObjectType(props)
}

func getInputTable(n int) flux.BufferedTable {
	floatArrayBuilder := array.NewFloatBuilder(mem)

	for i := 0.0; i < float64(n); i++ {
		floatArrayBuilder.Append(i)
	}
	a := floatArrayBuilder.NewFloatArray()

	for i := 0.0; i < float64(n); i++ {
		floatArrayBuilder.Append(i)
	}
	b := floatArrayBuilder.NewFloatArray()

	cr := &arrow.TableBuffer{
		GroupKey: execute.NewGroupKey(nil, nil),
		Columns: []flux.ColMeta{
			{Label: "a", Type: flux.TFloat},
			{Label: "b", Type: flux.TFloat},
		},
		Values: []array.Interface{a, b},
	}

	bt, err := table.Copy(table.FromBuffer(cr))
	if err != nil {
		panic(err)
	}
	return bt
}

func getOutputTable(n int) flux.BufferedTable {
	floatArrayBuilder := array.NewFloatBuilder(mem)

	for i := 0.0; i < float64(n); i++ {
		floatArrayBuilder.Append(i * 2)
	}
	result := floatArrayBuilder.NewFloatArray()

	cr := &arrow.TableBuffer{
		GroupKey: execute.NewGroupKey(nil, nil),
		Columns: []flux.ColMeta{
			{Label: "result", Type: flux.TFloat},
		},
		Values: []array.Interface{result},
	}

	bt, err := table.Copy(table.FromBuffer(cr))
	if err != nil {
		panic(err)
	}
	return bt
}

func colReaderToValues(cr flux.ColReader) values.Value {
	a := arrowutil.NewFloatArrayValue(cr.Floats(0))
	b := arrowutil.NewFloatArrayValue(cr.Floats(1))
	rec := values.NewObject(recordType(semantic.NewArrayType(semantic.BasicFloat)))
	rec.Set("a", a)
	rec.Set("b", b)
	return rec
}

func getArgs(rec values.Value) values.Object {
	args := values.NewObject(vectorizedArgsType)
	args.Set("r", rec)
	return args
}

func vectorizedMap(tb testing.TB, input flux.Table) flux.Table {

	b := table.NewBufferedBuilder(execute.NewGroupKey(nil, nil), mem)

	if err := input.Do(func(cr flux.ColReader) error {
		args := getArgs(colReaderToValues(cr))
		v, err := vectorizedCompiledMapFn.Eval(context.Background(), args)
		if err != nil {
			return err
		}
		o := v.Object()
		arrayValue, ok := o.Get("result")
		if !ok {
			return errors.New("no result column")
		}
		arr := arrayValue.(arrowutil.FloatArrayValue)
		floats := arr.GetArrowArray()
		tb := &arrow.TableBuffer{
			GroupKey: execute.NewGroupKey(nil, nil),
			Columns: []flux.ColMeta{
				{Label: "result", Type: flux.TFloat},
			},
			Values: []array.Interface{floats},
		}
		if err := b.AppendBuffer(tb); err != nil {
			return err
		}
		return nil
	}); err != nil {
		tb.Fatal(err)
	}

	t, err := b.Table()
	if err != nil {
		tb.Fatal(err)
	}
	return t
}

func baselineMap(tb testing.TB, input flux.Table) flux.Table {
	b := itable.NewArrowBuilder(execute.NewGroupKey(nil, nil), mem)
	b.Init([]flux.ColMeta{{Label: "result", Type: flux.TFloat}})
	b.Builders[0].Reserve(n)

	if err := input.Do(func(cr flux.ColReader) error {
		l := cr.Len()
		for i := 0; i < l; i++ {
			m, err := baselineCompiledMapFn.Eval(context.Background(), i, cr)
			if err != nil {
				return err
			}

			for j, c := range b.Cols() {
				v, ok := m.Get(c.Label)
				if !ok {
					return errors.New("no result column")
				}
				b.Builders[j].(*array.FloatBuilder).Append(v.Float())
			}
		}
		return nil
	}); err != nil {
		tb.Fatal(err)
	}

	t, err := b.Table()
	if err != nil {
		tb.Fatal(err)
	}
	return t

}

func baselineFilter(tb testing.TB, input flux.Table) []*arrowmem.Buffer {
	record := values.NewObject(baselineCompiledFilterFn.InputType())

	var bitsets []*arrowmem.Buffer
	if err := input.Do(func(cr flux.ColReader) error {
		l := cr.Len()
		bitset := arrowmem.NewResizableBuffer(mem)
		bitset.Resize(l)
		for i := 0; i < l; i++ {
			record.Set("a", execute.ValueForRow(cr, i, 0))
			record.Set("b", execute.ValueForRow(cr, i, 1))

			val, err := baselineCompiledFilterFn.Eval(context.Background(), record)
			if err != nil {
				bitset.Release()
				return err
			}
			bitutil.SetBitTo(bitset.Buf(), i, val)
		}
		bitsets = append(bitsets, bitset)
		return nil
	}); err != nil {
		tb.Fatal(err)
	}
	return bitsets

}

func vectorizedFilter(tb testing.TB, input flux.Table) []values.Array {
	var arrs []values.Array
	if err := input.Do(func(cr flux.ColReader) error {
		args := getArgs(colReaderToValues(cr))
		v, err := vectorizedCompiledFilterFn.Eval(context.Background(), args)
		if err != nil {
			return err
		}
		arrs = append(arrs, v.Array())
		return nil
	}); err != nil {
		tb.Fatal(err)
	}
	return arrs
}

type Array interface {
	Value(i int) valcompiler.Value
}

type floatArray struct {
	arr *array.Float
}

func (a floatArray) Value(i int) valcompiler.Value {
	return valcompiler.NewFloat(a.arr.Value(i))
}

func valueFilter(tb testing.TB, input flux.Table, args valcompiler.Value) []*arrowmem.Buffer {
	var bitsets []*arrowmem.Buffer
	if err := input.Do(func(cr flux.ColReader) error {
		l := cr.Len()
		bitset := arrowmem.NewResizableBuffer(mem)
		bitset.Resize(l)
		var (
			arr0 Array = floatArray{arr: cr.Floats(0)}
			arr1 Array = floatArray{arr: cr.Floats(1)}
		)
		for i := 0; i < l; i++ {
			args.Set(0, arr0.Value(i))
			args.Set(1, arr1.Value(i))
			// args.Set(0, valcompiler.ValueForRow(cr, i, 0))
			// args.Set(1, valcompiler.ValueForRow(cr, i, 1))

			val, err := valueCompiledFilterFn.Eval(context.Background(), args)
			if err != nil {
				bitset.Release()
				return err
			}
			bitutil.SetBitTo(bitset.Buf(), i, val.Bool())
		}
		bitsets = append(bitsets, bitset)
		return nil
	}); err != nil {
		tb.Fatal(err)
	}
	return bitsets

}
