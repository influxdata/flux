package compiler_test

import (
	"context"
	"errors"
	"testing"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/compiler"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/table"
	"github.com/influxdata/flux/internal/arrowutil"
	itable "github.com/influxdata/flux/internal/execute/table"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

var (
	n      = 1000
	fnExpr = mustGetFnExpr(`(r) => ({result: r.a + r.b})`)

	baselineCompiledFn   *execute.RowMapPreparedFn
	vectorizedArgsType   semantic.MonoType
	vectorizedCompiledFn compiler.Func

	mem = &memory.Allocator{}
)

func init() {
	var err error

	fn := execute.NewRowMapFn(fnExpr, nil)
	baselineCompiledFn, err = fn.Prepare([]flux.ColMeta{
		{Label: "a", Type: flux.TFloat},
		{Label: "b", Type: flux.TFloat},
	})
	if err != nil {
		panic(err)
	}

	vrt := recordType(semantic.NewArrayType(semantic.BasicFloat))
	vectorizedArgsType = argsType(vrt)
	vectorizedCompiledFn, err = compiler.Compile(nil, fnExpr, vectorizedArgsType)
	if err != nil {
		panic(err)
	}
}

func TestVectorize(t *testing.T) {
	inputTable := getInputTable(n)
	outputTable := getOutputTable(n)
	t.Run("baseline", func(t *testing.T) {
		gt := baselineMap(t, inputTable.Copy())
		got := table.Iterator{gt}
		want := table.Iterator{outputTable.Copy()}
		if diff := table.Diff(want, got); diff != "" {
			t.Errorf("tables were different: %s", diff)
		}
	})
	t.Run("vectorized", func(t *testing.T) {
		gt := vectorizedMap(t, inputTable.Copy())
		got := table.Iterator{gt}
		want := table.Iterator{outputTable.Copy()}
		if diff := table.Diff(want, got); diff != "" {
			t.Errorf("tables were different: %s", diff)
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
		v, err := vectorizedCompiledFn.Eval(context.Background(), args)
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
			m, err := baselineCompiledFn.Eval(context.Background(), i, cr)
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
