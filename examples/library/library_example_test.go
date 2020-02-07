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
	t := `import g "internal/gen"
g.tables(n: 6, seed: 0) |> keep(columns: ["_value"])
`

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

	ctx = executetest.NewTestExecuteDependencies().Inject(ctx)
	alloc := &memory.Allocator{}
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
	// #datatype,string,long,double
	// #group,false,false,false
	// #default,_result,,
	// ,result,table,_value
	// ,,0,-14.079293543218107
	// ,,0,28.54665479040335
	// ,,0,-84.60098163078523
	// ,,0,9.981145558465496
	// ,,0,95.9759964561731
	// ,,0,44.77419397459176
}
