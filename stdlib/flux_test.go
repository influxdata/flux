package stdlib_test

import (
	"bufio"
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/querytest"
	"github.com/influxdata/flux/stdlib"
)

func init() {
	flux.FinalizeBuiltIns()
}

var querier = querytest.NewQuerier()

func TestFluxEndToEnd(t *testing.T) {
	runEndToEnd(t, querier, stdlib.FluxTestPackages)
}
func BenchmarkFluxEndToEnd(b *testing.B) {
	benchEndToEnd(b, querier, stdlib.FluxTestPackages)
}

func runEndToEnd(t *testing.T, querier *querytest.Querier, pkgs []*ast.Package) {
	for _, pkg := range pkgs {
		name := pkg.Files[0].Name
		c := compiler{pkg: pkg}
		t.Run(name, func(t *testing.T) {
			testFlux(t, querier, c)
		})
	}
}

func benchEndToEnd(b *testing.B, querier *querytest.Querier, pkgs []*ast.Package) {
	for _, pkg := range pkgs {
		name := pkg.Files[0].Name
		c := compiler{pkg: pkg}
		b.Run(name, func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				testFlux(b, querier, c)
			}
		})
	}
}

func testFlux(t testing.TB, querier *querytest.Querier, compiler flux.Compiler) {
	r, err := querier.C.Query(context.Background(), compiler)
	if err != nil {
		t.Fatalf("unexpected error while executing test: %v", err)
	}
	defer r.Done()
	result, ok := <-r.Ready()
	if !ok {
		t.Fatalf("TEST error retrieving query result: %s", r.Err())
	}

	var out bytes.Buffer
	defer func() {
		if t.Failed() {
			scanner := bufio.NewScanner(&out)
			for scanner.Scan() {
				t.Log(scanner.Text())
			}
		}
	}()

	for _, res := range result {
		if err := res.Tables().Do(func(tbl flux.Table) error {
			_, _ = execute.NewFormatter(tbl, nil).WriteTo(&out)
			return nil
		}); err != nil {
			t.Error(err)
		}
	}
}

type compiler struct {
	pkg *ast.Package
}

func (c compiler) Compile(ctx context.Context) (*flux.Spec, error) {
	return flux.CompileAST(ctx, c.pkg, time.Now())
}

func (c compiler) CompilerType() flux.CompilerType {
	return "test"
}
