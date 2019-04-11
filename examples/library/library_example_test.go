package library_test

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	_ "github.com/influxdata/flux/builtin"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/cmd/flux/cmd"
	"github.com/influxdata/flux/csv"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/lang"
	"github.com/influxdata/flux/parser"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

func Example_fromGenerator() {
	t := `import g "generate"
g.from(start: 1993-02-16T00:00:00Z, stop: 1993-02-16T00:03:00Z, count: 5, fn: (n) => 1)`

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	i := interpreter.NewInterpreter()
	scope := flux.Prelude()

	astPkg := parser.ParseSource(t)
	if ast.Check(astPkg) > 0 {
		panic(ast.GetError(astPkg))
	}

	semPkg, err := semantic.New(astPkg)
	if err != nil {
		panic(err)
	}

	if _, err := i.Eval(semPkg, scope, flux.StdLib()); err != nil {
		panic(err)
	}

	v := scope.Return()

	// Ignore statements that do not return a value
	if v == nil {
		return
	}

	// Check for yield and execute query
	if v.Type() == flux.TableObjectMonoType {
		t := v.(*flux.TableObject)
		now, ok := scope.Lookup("now")
		if !ok {
			panic(fmt.Errorf("now option not set"))
		}
		nowTime, err := now.Function().Call(nil)
		if err != nil {
			panic(err)
		}
		spec, err := flux.ToSpec([]values.Value{t}, nowTime.Time().Time())
		if err != nil {
			panic(err)
		}
		compiler := lang.SpecCompiler{
			Spec: spec,
		}

		querier := cmd.NewQuerier()

		results, err := querier.Query(ctx, compiler)
		if err != nil {
			panic(err)
		}
		defer results.Release()

		buf := bytes.NewBuffer(nil)
		encoder := csv.NewMultiResultEncoder(csv.DefaultEncoderConfig())

		if _, err := encoder.Encode(buf, results); err != nil {
			panic(err)
		}

		// This substitution is done because the testable example's Output
		// section cannot contain carriage return while the csv encoder emits them
		fmt.Println(strings.Replace(buf.String(), "\r\n", "\n", -1))

		// Output:
		// #datatype,string,long,dateTime:RFC3339,long
		// #group,false,false,false,false
		// #default,_result,,,
		// ,result,table,_time,_value
		// ,,0,1993-02-16T00:00:00Z,1
		// ,,0,1993-02-16T00:00:36Z,1
		// ,,0,1993-02-16T00:01:12Z,1
		// ,,0,1993-02-16T00:01:48Z,1
		// ,,0,1993-02-16T00:02:24Z,1
	}
}
