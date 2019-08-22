package library_test

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/ast"
	_ "github.com/influxdata/flux/builtin"
	"github.com/influxdata/flux/csv"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/lang"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/parser"
)

func Example_fromGenerator() {
	t := `import g "generate"
g.from(start: 1993-02-16T00:00:00Z, stop: 1993-02-16T00:03:00Z, count: 5, fn: (n) => 1)`

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	astPkg := parser.ParseSource(t)
	if ast.Check(astPkg) > 0 {
		panic(ast.GetError(astPkg))
	}
	compiler := lang.ASTCompiler{
		AST: astPkg,
	}

	program, err := compiler.Compile(ctx)
	if err != nil {
		panic(err)
	}

	alloc := &memory.Allocator{}

	if p, ok := program.(lang.DependenciesAwareProgram); ok {
		p.SetExecutorDependencies(executetest.NewTestExecuteDependencies())
	}
	q, err := program.Start(ctx, alloc)
	if err != nil {
		panic(err)
	}

	results := flux.NewResultIteratorFromQuery(q)
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
