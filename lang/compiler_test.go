package lang_test

import (
	"bytes"
	"context"
	"math"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/ast"
	_ "github.com/influxdata/flux/builtin"
	fcsv "github.com/influxdata/flux/csv"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/lang"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/mock"
	"github.com/influxdata/flux/parser"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/plan/plantest"
	"github.com/influxdata/flux/stdlib/csv"
	"github.com/influxdata/flux/stdlib/influxdata/influxdb"
	"github.com/influxdata/flux/stdlib/universe"
	"github.com/influxdata/flux/values"
)

func init() {
	execute.RegisterSource(influxdb.FromKind, mock.CreateMockFromSource)
}

func TestFluxCompiler(t *testing.T) {
	ctx := context.Background()

	for _, tc := range []struct {
		q  string
		ok bool
	}{
		{q: `from(bucket: "foo")`, ok: true},
		{q: `t=""t.t`},
		{q: `t=0t.s`},
		{q: `x = from(bucket: "foo")`},
		{q: `x = from(bucket: "foo") |> yield()`, ok: true},
		{q: `from(bucket: "foo")`, ok: true},
	} {
		c := lang.FluxCompiler{
			Query: tc.q,
		}
		program, err := c.Compile(ctx)
		if err != nil {
			t.Fatalf("failed to compile AST: %v", err)
		}
		// we need to start the program to get compile errors derived from AST evaluation
		if _, err = program.Start(context.Background(), &memory.Allocator{}); tc.ok && err != nil {
			t.Errorf("expected query %q to compile successfully but got error %v", tc.q, err)
		} else if !tc.ok && err == nil {
			t.Errorf("expected query %q to compile with error but got no error", tc.q)
		}
	}
}

func TestASTCompiler(t *testing.T) {
	testcases := []struct {
		name   string
		now    time.Time
		file   *ast.File
		script string
		want   plantest.PlanSpec
	}{
		{
			name: "override now time using now option",
			now:  time.Unix(1, 1),
			script: `
import "csv"
option now = () => 2017-10-10T00:01:00Z
csv.from(csv: "foo,bar") |> range(start: 2017-10-10T00:00:00Z)
`,
			want: plantest.PlanSpec{
				Nodes: []plan.Node{
					&plan.PhysicalPlanNode{Spec: &csv.FromCSVProcedureSpec{}},
					&plan.PhysicalPlanNode{Spec: &universe.RangeProcedureSpec{}},
					&plan.PhysicalPlanNode{Spec: &universe.YieldProcedureSpec{}},
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
				},
				Resources: flux.ResourceManagement{ConcurrencyQuota: 1, MemoryBytesQuota: math.MaxInt64},
				Now:       parser.MustParseTime("2017-10-10T00:01:00Z").Value,
			},
		},
		{
			name: "get now time from compiler",
			now:  parser.MustParseTime("2018-10-10T00:00:00Z").Value,
			script: `
import "csv"
csv.from(csv: "foo,bar") |> range(start: 2017-10-10T00:00:00Z)
`,
			want: plantest.PlanSpec{
				Nodes: []plan.Node{
					&plan.PhysicalPlanNode{Spec: &csv.FromCSVProcedureSpec{}},
					&plan.PhysicalPlanNode{Spec: &universe.RangeProcedureSpec{}},
					&plan.PhysicalPlanNode{Spec: &universe.YieldProcedureSpec{}},
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
				},
				Resources: flux.ResourceManagement{ConcurrencyQuota: 1, MemoryBytesQuota: math.MaxInt64},
				Now:       parser.MustParseTime("2018-10-10T00:00:00Z").Value,
			},
		},
		{
			name: "prepend file",
			file: &ast.File{
				Body: []ast.Statement{
					&ast.OptionStatement{
						Assignment: &ast.VariableAssignment{
							ID: &ast.Identifier{Name: "now"},
							Init: &ast.FunctionExpression{
								Body: &ast.DateTimeLiteral{
									Value: parser.MustParseTime("2018-10-10T00:00:00Z").Value,
								},
							},
						},
					},
				},
			},
			script: `
import "csv"
csv.from(csv: "foo,bar") |> range(start: 2017-10-10T00:00:00Z)
`,
			want: plantest.PlanSpec{
				Nodes: []plan.Node{
					&plan.PhysicalPlanNode{Spec: &csv.FromCSVProcedureSpec{}},
					&plan.PhysicalPlanNode{Spec: &universe.RangeProcedureSpec{}},
					&plan.PhysicalPlanNode{Spec: &universe.YieldProcedureSpec{}},
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
				},
				Resources: flux.ResourceManagement{ConcurrencyQuota: 1, MemoryBytesQuota: math.MaxInt64},
				Now:       parser.MustParseTime("2018-10-10T00:00:00Z").Value,
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			astPkg, err := flux.Parse(tc.script)
			if err != nil {
				t.Fatalf("failed to parse script: %v", err)
			}

			c := lang.ASTCompiler{
				AST: astPkg,
				Now: tc.now,
			}

			if tc.file != nil {
				c.PrependFile(tc.file)
			}

			program, err := c.Compile(context.Background())
			if err != nil {
				t.Fatalf("failed to compile AST: %v", err)
			}
			// we need to start the program to get compile errors derived from AST evaluation
			if _, err := program.Start(context.Background(), &memory.Allocator{}); err != nil {
				t.Fatalf("failed to start program: %v", err)
			}

			got := program.(*lang.AstProgram).PlanSpec
			want := plantest.CreatePlanSpec(&tc.want)
			if err := plantest.ComparePlansShallow(want, got); err != nil {
				t.Error(err)
			}
		})
	}
}

// TestTableObjectCompiler evaluates a simple `from |> range |> filter` script on csv data, and
// extracts the TableObjects obtained from evaluation. It eventually compiles TableObjects and
// compares obtained results with expected ones (obtained from decoding the raw csv data).
func TestTableObjectCompiler(t *testing.T) {
	dataRaw := `#datatype,string,long,dateTime:RFC3339,long,string,string,string,string
#group,false,false,false,false,false,false,true,true
#default,_result,,,,,,,
,result,table,_time,_value,_field,_measurement,host,name
,,0,2018-05-22T19:53:26Z,15204688,io_time,diskio,host.local,disk0
,,0,2018-05-22T19:53:36Z,15204894,io_time,diskio,host.local,disk0
,,0,2018-05-22T19:53:46Z,15205102,io_time,diskio,host.local,disk0
,,0,2018-05-22T19:53:56Z,15205226,io_time,diskio,host.local,disk0
,,0,2018-05-22T19:54:06Z,15205499,io_time,diskio,host.local,disk0
,,0,2018-05-22T19:54:16Z,15205755,io_time,diskio,host.local,disk0
,,1,2018-05-22T19:53:26Z,648,io_time,diskio,host.local,disk2
,,1,2018-05-22T19:53:36Z,648,io_time,diskio,host.local,disk2
,,1,2018-05-22T19:53:46Z,648,io_time,diskio,host.local,disk2
,,1,2018-05-22T19:53:56Z,648,io_time,diskio,host.local,disk2
,,1,2018-05-22T19:54:06Z,648,io_time,diskio,host.local,disk2
,,1,2018-05-22T19:54:16Z,648,io_time,diskio,host.local,disk2
`

	rangedDataRaw := `#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#group,false,false,true,true,false,false,false,false,true,true
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,name
,,0,2017-10-10T00:00:00Z,2018-05-22T19:54:00Z,2018-05-22T19:53:26Z,15204688,io_time,diskio,host.local,disk0
,,0,2017-10-10T00:00:00Z,2018-05-22T19:54:00Z,2018-05-22T19:53:36Z,15204894,io_time,diskio,host.local,disk0
,,0,2017-10-10T00:00:00Z,2018-05-22T19:54:00Z,2018-05-22T19:53:46Z,15205102,io_time,diskio,host.local,disk0
,,0,2017-10-10T00:00:00Z,2018-05-22T19:54:00Z,2018-05-22T19:53:56Z,15205226,io_time,diskio,host.local,disk0
,,1,2017-10-10T00:00:00Z,2018-05-22T19:54:00Z,2018-05-22T19:53:26Z,648,io_time,diskio,host.local,disk2
,,1,2017-10-10T00:00:00Z,2018-05-22T19:54:00Z,2018-05-22T19:53:36Z,648,io_time,diskio,host.local,disk2
,,1,2017-10-10T00:00:00Z,2018-05-22T19:54:00Z,2018-05-22T19:53:46Z,648,io_time,diskio,host.local,disk2
,,1,2017-10-10T00:00:00Z,2018-05-22T19:54:00Z,2018-05-22T19:53:56Z,648,io_time,diskio,host.local,disk2
`

	filteredDataRaw := `#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#group,false,false,true,true,false,false,false,false,true,true
#default,_result,0,2017-10-10T00:00:00Z,2018-05-22T19:54:00Z,,,,,host.local,disk0
,result,table,_start,_stop,_time,_value,_field,_measurement,host,name

#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#group,false,false,true,true,false,false,false,false,true,true
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,name
,,1,2017-10-10T00:00:00Z,2018-05-22T19:54:00Z,2018-05-22T19:53:26Z,648,io_time,diskio,host.local,disk2
,,1,2017-10-10T00:00:00Z,2018-05-22T19:54:00Z,2018-05-22T19:53:36Z,648,io_time,diskio,host.local,disk2
,,1,2017-10-10T00:00:00Z,2018-05-22T19:54:00Z,2018-05-22T19:53:46Z,648,io_time,diskio,host.local,disk2
,,1,2017-10-10T00:00:00Z,2018-05-22T19:54:00Z,2018-05-22T19:53:56Z,648,io_time,diskio,host.local,disk2
`

	script := `import "csv"
data = "` + dataRaw + `"
csv.from(csv: data)
	|> range(start: 2017-10-10T00:00:00Z, stop: 2018-05-22T19:54:00Z)
	|> filter(fn: (r) => r._value < 1000)`

	wantFrom := getTablesFromRawOrFail(t, dataRaw)
	wantRange := getTablesFromRawOrFail(t, rangedDataRaw)
	wantFilter := getTablesFromRawOrFail(t, filteredDataRaw)

	vs, _, err := flux.Eval(script)
	if err != nil {
		t.Fatal(err)
	}

	tos := getTableObjects(vs)
	if len(tos) != 3 {
		t.Fatalf("wrong number of table objects, got %d", len(tos))
	}

	fromCsvTO := tos[2]
	if fromCsvTO.Kind != csv.FromCSVKind {
		t.Fatalf("unexpected kind for fromCSV: %s", fromCsvTO.Kind)
	}
	rangeTO := tos[1]
	if rangeTO.Kind != universe.RangeKind {
		t.Fatalf("unexpected kind for range: %s", rangeTO.Kind)
	}
	filterTO := tos[0]
	if filterTO.Kind != universe.FilterKind {
		t.Fatalf("unexpected kind for filter: %s", filterTO.Kind)
	}

	compareTableObjectWithTables(t, fromCsvTO, wantFrom)
	compareTableObjectWithTables(t, rangeTO, wantRange)
	compareTableObjectWithTables(t, filterTO, wantFilter)
	// run it twice to ensure compilation is idempotent and there are no side-effects
	compareTableObjectWithTables(t, fromCsvTO, wantFrom)
	compareTableObjectWithTables(t, rangeTO, wantRange)
	compareTableObjectWithTables(t, filterTO, wantFilter)
}

func compareTableObjectWithTables(t *testing.T, to *flux.TableObject, want []*executetest.Table) {
	t.Helper()

	got := getTableObjectTablesOrFail(t, to)
	if !cmp.Equal(want, got) {
		t.Fatalf("unexpected result -want/+got\n\n%s\n\n", cmp.Diff(want, got))
	}
}

func getTablesFromRawOrFail(t *testing.T, rawData string) []*executetest.Table {
	t.Helper()

	b := bytes.NewReader([]byte(rawData))
	result, err := fcsv.NewResultDecoder(fcsv.ResultDecoderConfig{}).Decode(b)
	if err != nil {
		t.Fatal(err)
	}

	return getTablesFromResultOrFail(t, result)
}

func getTableObjectTablesOrFail(t *testing.T, to *flux.TableObject) []*executetest.Table {
	t.Helper()

	toc := lang.TableObjectCompiler{
		Tables: to,
	}

	program, err := toc.Compile(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	q, err := program.Start(context.Background(), &memory.Allocator{})
	if err != nil {
		t.Fatal(err)
	}
	result := <-q.Results()
	if _, ok := <-q.Results(); ok {
		t.Fatalf("got more then one result for %s", to.Kind)
	}
	tables := getTablesFromResultOrFail(t, result)
	q.Done()
	if err := q.Err(); err != nil {
		t.Fatal(err)
	}
	return tables
}

func getTablesFromResultOrFail(t *testing.T, result flux.Result) []*executetest.Table {
	t.Helper()

	tables := make([]*executetest.Table, 0)
	if err := result.Tables().Do(func(table flux.Table) error {
		converted, err := executetest.ConvertTable(table)
		if err != nil {
			return err
		}
		tables = append(tables, converted)
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	executetest.NormalizeTables(tables)
	return tables
}

func getTableObjects(vs []values.Value) []*flux.TableObject {
	tos := make([]*flux.TableObject, 0)
	for _, v := range vs {
		if to, ok := v.(*flux.TableObject); ok {
			tos = append(tos, to)
			tos = append(tos, getTableObjectsFromArray(to.Parents)...)
		}
	}
	return tos
}

func getTableObjectsFromArray(parents values.Array) []*flux.TableObject {
	tos := make([]*flux.TableObject, 0)
	parents.Range(func(i int, v values.Value) {
		if to, ok := v.(*flux.TableObject); ok {
			tos = append(tos, to)
			tos = append(tos, getTableObjectsFromArray(to.Parents)...)
		}
	})
	return tos
}
