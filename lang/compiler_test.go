package lang_test

import (
	"bytes"
	"context"
	"encoding/json"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/ast"
	_ "github.com/influxdata/flux/builtin"
	fcsv "github.com/influxdata/flux/csv"
	"github.com/influxdata/flux/dependencies/dependenciestest"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/interpreter"
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
		name   string
		now    time.Time
		extern *ast.File
		q      string
		err    string
	}{
		{
			name: "simple",
			q:    `from(bucket: "foo")`,
		},
		{
			name: "syntax error",
			q:    `t={]`,
			err:  "expected RBRACE",
		},
		{
			name: "type error",
			q:    `t=0 t.s`,
			err:  "type error",
		},
		{
			name: "from with no streaming data",
			q:    `x = from(bucket: "foo")`,
			err:  "no streaming data",
		},
		{
			name: "from with yield",
			q:    `x = from(bucket: "foo") |> yield()`,
		},
		{
			name: "extern",
			extern: &ast.File{
				Body: []ast.Statement{
					&ast.OptionStatement{
						Assignment: &ast.VariableAssignment{
							ID:   &ast.Identifier{Name: "twentySix"},
							Init: &ast.IntegerLiteral{Value: 26},
						},
					},
				},
			},
			q: `twentySeven = twentySix + 1
				twentySeven
				from(bucket: "foo")`,
		},
		{
			name: "extern with error",
			extern: &ast.File{
				Body: []ast.Statement{
					&ast.OptionStatement{
						Assignment: &ast.VariableAssignment{
							ID:   &ast.Identifier{Name: "twentySix"},
							Init: &ast.IntegerLiteral{Value: 26},
						},
					},
				},
			},
			q: `twentySeven = twentyFive + 2
				twentySeven
				from(bucket: "foo")`,
			err: "undefined identifier",
		},
		{
			name: "with now",
			now:  time.Unix(1000, 0),
			q:    `from(bucket: "foo")`,
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			c := lang.FluxCompiler{
				Now:    tc.now,
				Extern: tc.extern,
				Query:  tc.q,
			}

			// serialize and deserialize and make sure they are equal
			bs, err := json.Marshal(c)
			if err != nil {
				t.Error(err)
			}
			cc := lang.FluxCompiler{}
			err = json.Unmarshal(bs, &cc)
			if err != nil {
				t.Error(err)
			}
			if diff := cmp.Diff(c, cc); diff != "" {
				t.Errorf("compiler serialized/deserialized does not match: -want/+got:\n%v", diff)
			}

			program, err := c.Compile(ctx)
			if err != nil {
				if tc.err != "" {
					if !strings.Contains(err.Error(), tc.err) {
						t.Fatalf(`expected query to error with "%v" but got "%v"`, tc.err, err)
					} else {
						return
					}
				}
				t.Fatalf("failed to compile AST: %v", err)
			}

			astProg := program.(*lang.AstProgram)
			if astProg.Now != tc.now {
				t.Errorf(`unexpected value for now, want "%v", got "%v"`, tc.now, astProg.Now)
			}

			if p, ok := program.(lang.DependenciesAwareProgram); ok {
				p.SetExecutorDependencies(executetest.NewTestExecuteDependencies())
			}

			// we need to start the program to get compile errors derived from AST evaluation
			if _, err = program.Start(context.Background(), &memory.Allocator{}); tc.err == "" && err != nil {
				t.Errorf("expected query %q to start successfully but got error %v", tc.q, err)
			} else if tc.err != "" && err == nil {
				t.Errorf("expected query %q to start with error but got no error", tc.q)
			} else if tc.err != "" && err != nil && !strings.Contains(err.Error(), tc.err) {
				t.Errorf(`expected query to error with "%v" but got "%v"`, tc.err, err)
			}
		})
	}
}

func TestCompilationError(t *testing.T) {
	program, err := lang.Compile(`illegal query`, time.Unix(0, 0))
	if err != nil {
		// This shouldn't happen, has the script should be evaluated at program Start.
		t.Fatal(err)
	}
	program.SetExecutorDependencies(executetest.NewTestExecuteDependencies())

	_, err = program.Start(context.Background(), &memory.Allocator{})
	if err == nil {
		t.Fatal("compilation error expected, got none")
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

			// serialize and deserialize and make sure they are equal
			bs, err := json.Marshal(c)
			if err != nil {
				t.Error(err)
			}
			cc := lang.ASTCompiler{}
			err = json.Unmarshal(bs, &cc)
			if err != nil {
				t.Error(err)
			}
			if diff := cmp.Diff(c, cc); diff != "" {
				t.Errorf("compiler serialized/deserialized does not match: -want/+got:\n%v", diff)
			}

			if tc.file != nil {
				c.PrependFile(tc.file)
			}

			program, err := c.Compile(context.Background())
			if err != nil {
				t.Fatalf("failed to compile AST: %v", err)
			}
			if p, ok := program.(lang.DependenciesAwareProgram); ok {
				p.SetExecutorDependencies(executetest.NewTestExecuteDependencies())
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

func TestCompileOptions(t *testing.T) {
	src := `import "csv"
			csv.from(csv: "foo,bar")
				|> range(start: 2017-10-10T00:00:00Z)
				|> count()`

	now := parser.MustParseTime("2018-10-10T00:00:00Z").Value

	opt := lang.WithLogPlanOpts(plan.OnlyLogicalRules(removeCount{}))

	program, err := lang.Compile(src, now, opt)
	if err != nil {
		t.Fatalf("failed to compile script: %v", err)
	}
	program.SetExecutorDependencies(executetest.NewTestExecuteDependencies())

	// start program in order to evaluate planner options
	if _, err := program.Start(context.Background(), &memory.Allocator{}); err != nil {
		t.Fatalf("failed to start program: %v", err)
	}

	want := plantest.CreatePlanSpec(&plantest.PlanSpec{
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
	})

	if err := plantest.ComparePlansShallow(want, program.PlanSpec); err != nil {
		t.Fatalf("unexpected plans: %v", err)
	}
}

type removeCount struct{}

func (rule removeCount) Name() string {
	return "removeCountRule"
}
func (rule removeCount) Pattern() plan.Pattern {
	return plan.Pat(universe.CountKind, plan.Any())
}
func (rule removeCount) Rewrite(node plan.Node) (plan.Node, bool, error) {
	return node.Predecessors()[0], true, nil
}

func TestCompileOptions_FromFluxOptions(t *testing.T) {
	nowFn := func() time.Time {
		return parser.MustParseTime("2018-10-10T00:00:00Z").Value
	}

	plan.RegisterPhysicalRules(&plantest.MergeFromRangePhysicalRule{})
	plan.RegisterLogicalRules(&removeCount{})

	tcs := []struct {
		name    string
		files   []string
		want    *plan.Spec
		wantErr string
	}{
		{
			name: "no planner option set",
			files: []string{`
import "planner"

from(bucket: "bkt") |> range(start: 0) |> filter(fn: (r) => r._value > 0) |> count()`},
			want: plantest.CreatePlanSpec(&plantest.PlanSpec{
				Nodes: []plan.Node{
					&plan.PhysicalPlanNode{Spec: &influxdb.FromProcedureSpec{}},
					&plan.PhysicalPlanNode{Spec: &universe.FilterProcedureSpec{}},
					&plan.PhysicalPlanNode{Spec: &universe.YieldProcedureSpec{}},
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
				},
				Resources: flux.ResourceManagement{ConcurrencyQuota: 1, MemoryBytesQuota: math.MaxInt64},
				Now:       nowFn(),
			}),
		},
		{
			name: "remove push down range",
			files: []string{`
import "planner"

option planner.disablePhysicalRules = ["fromRangeRule"]

from(bucket: "bkt") |> range(start: 0) |> filter(fn: (r) => r._value > 0) |> count()`},
			want: plantest.CreatePlanSpec(&plantest.PlanSpec{
				Nodes: []plan.Node{
					&plan.PhysicalPlanNode{Spec: &influxdb.FromProcedureSpec{}},
					&plan.PhysicalPlanNode{Spec: &universe.RangeProcedureSpec{}},
					&plan.PhysicalPlanNode{Spec: &universe.FilterProcedureSpec{}},
					&plan.PhysicalPlanNode{Spec: &universe.YieldProcedureSpec{}},
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
					{2, 3},
				},
				Resources: flux.ResourceManagement{ConcurrencyQuota: 1, MemoryBytesQuota: math.MaxInt64},
				Now:       nowFn(),
			}),
		},
		{
			name: "remove push down range and count",
			files: []string{`
import "planner"

option planner.disablePhysicalRules = ["fromRangeRule"]
option planner.disableLogicalRules = ["removeCountRule"]

from(bucket: "bkt") |> range(start: 0) |> filter(fn: (r) => r._value > 0) |> count()`},
			want: plantest.CreatePlanSpec(&plantest.PlanSpec{
				Nodes: []plan.Node{
					&plan.PhysicalPlanNode{Spec: &influxdb.FromProcedureSpec{}},
					&plan.PhysicalPlanNode{Spec: &universe.RangeProcedureSpec{}},
					&plan.PhysicalPlanNode{Spec: &universe.FilterProcedureSpec{}},
					&plan.PhysicalPlanNode{Spec: &universe.CountProcedureSpec{}},
					&plan.PhysicalPlanNode{Spec: &universe.YieldProcedureSpec{}},
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
					{2, 3},
					{3, 4},
				},
				Resources: flux.ResourceManagement{ConcurrencyQuota: 1, MemoryBytesQuota: math.MaxInt64},
				Now:       nowFn(),
			}),
		},
		{
			name: "remove push down range and count - with non existent rule",
			files: []string{`
import "planner"

option planner.disablePhysicalRules = ["fromRangeRule", "non_existent"]
option planner.disableLogicalRules = ["removeCountRule", "non_existent"]

from(bucket: "bkt") |> range(start: 0) |> filter(fn: (r) => r._value > 0) |> count()`},
			want: plantest.CreatePlanSpec(&plantest.PlanSpec{
				Nodes: []plan.Node{
					&plan.PhysicalPlanNode{Spec: &influxdb.FromProcedureSpec{}},
					&plan.PhysicalPlanNode{Spec: &universe.RangeProcedureSpec{}},
					&plan.PhysicalPlanNode{Spec: &universe.FilterProcedureSpec{}},
					&plan.PhysicalPlanNode{Spec: &universe.CountProcedureSpec{}},
					&plan.PhysicalPlanNode{Spec: &universe.YieldProcedureSpec{}},
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
					{2, 3},
					{3, 4},
				},
				Resources: flux.ResourceManagement{ConcurrencyQuota: 1, MemoryBytesQuota: math.MaxInt64},
				Now:       nowFn(),
			}),
		},
		{
			name: "remove non existent rules does not produce any effect",
			files: []string{`
import "planner"

option planner.disablePhysicalRules = ["foo", "bar", "mew", "buz", "foxtrot"]
option planner.disableLogicalRules = ["foo", "bar", "mew", "buz", "foxtrot"]

from(bucket: "bkt") |> range(start: 0) |> filter(fn: (r) => r._value > 0) |> count()`},
			want: plantest.CreatePlanSpec(&plantest.PlanSpec{
				Nodes: []plan.Node{
					&plan.PhysicalPlanNode{Spec: &influxdb.FromProcedureSpec{}},
					&plan.PhysicalPlanNode{Spec: &universe.FilterProcedureSpec{}},
					&plan.PhysicalPlanNode{Spec: &universe.YieldProcedureSpec{}},
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
				},
				Resources: flux.ResourceManagement{ConcurrencyQuota: 1, MemoryBytesQuota: math.MaxInt64},
				Now:       nowFn(),
			}),
		},
		{
			name: "empty planner option does not produce any effect",
			files: []string{`
import "planner"

option planner.disablePhysicalRules = [""]
option planner.disableLogicalRules = [""]

from(bucket: "bkt") |> range(start: 0) |> filter(fn: (r) => r._value > 0) |> count()`},
			want: plantest.CreatePlanSpec(&plantest.PlanSpec{
				Nodes: []plan.Node{
					&plan.PhysicalPlanNode{Spec: &influxdb.FromProcedureSpec{}},
					&plan.PhysicalPlanNode{Spec: &universe.FilterProcedureSpec{}},
					&plan.PhysicalPlanNode{Spec: &universe.YieldProcedureSpec{}},
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
				},
				Resources: flux.ResourceManagement{ConcurrencyQuota: 1, MemoryBytesQuota: math.MaxInt64},
				Now:       nowFn(),
			}),
		},
		{
			name: "logical planner option must be an array",
			files: []string{`
import "planner"

option planner.disableLogicalRules = "not an array"

// remember to return streaming data
from(bucket: "does_not_matter")`},
			wantErr: `'planner.disableLogicalRules' must be an array of string, got string`,
		},
		{
			name: "physical planner option must be an array",
			files: []string{`
import "planner"

option planner.disablePhysicalRules = "not an array"

// remember to return streaming data
from(bucket: "does_not_matter")`},
			wantErr: `'planner.disablePhysicalRules' must be an array of string, got string`,
		},
		{
			name: "logical planner option must be an array of strings",
			files: []string{`
import "planner"

option planner.disableLogicalRules = [1.0]

// remember to return streaming data
from(bucket: "does_not_matter")`},
			wantErr: `'planner.disableLogicalRules' must be an array of string, got an array of float`,
		},
		{
			name: "physical planner option must be an array of strings",
			files: []string{`
import "planner"

option planner.disablePhysicalRules = [1.0]

// remember to return streaming data
from(bucket: "does_not_matter")`},
			wantErr: `'planner.disablePhysicalRules' must be an array of string, got an array of float`,
		},
		{
			name: "planner is an object defined by the user",
			files: []string{`
planner = {
	disablePhysicalRules: ["fromRangeRule"],
	disableLogicalRules: ["removeCountRule"]
}

from(bucket: "bkt") |> range(start: 0) |> filter(fn: (r) => r._value > 0) |> count()`},
			// This shouldn't change the plan.
			want: plantest.CreatePlanSpec(&plantest.PlanSpec{
				Nodes: []plan.Node{
					&plan.PhysicalPlanNode{Spec: &influxdb.FromProcedureSpec{}},
					&plan.PhysicalPlanNode{Spec: &universe.FilterProcedureSpec{}},
					&plan.PhysicalPlanNode{Spec: &universe.YieldProcedureSpec{}},
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
				},
				Resources: flux.ResourceManagement{ConcurrencyQuota: 1, MemoryBytesQuota: math.MaxInt64},
				Now:       nowFn(),
			}),
		},
		{
			name: "use planner option with alias",
			files: []string{`
import pl "planner"

option pl.disablePhysicalRules = ["fromRangeRule"]
option pl.disableLogicalRules = ["removeCountRule"]

from(bucket: "bkt") |> range(start: 0) |> filter(fn: (r) => r._value > 0) |> count()`},
			want: plantest.CreatePlanSpec(&plantest.PlanSpec{
				Nodes: []plan.Node{
					&plan.PhysicalPlanNode{Spec: &influxdb.FromProcedureSpec{}},
					&plan.PhysicalPlanNode{Spec: &universe.RangeProcedureSpec{}},
					&plan.PhysicalPlanNode{Spec: &universe.FilterProcedureSpec{}},
					&plan.PhysicalPlanNode{Spec: &universe.CountProcedureSpec{}},
					&plan.PhysicalPlanNode{Spec: &universe.YieldProcedureSpec{}},
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
					{2, 3},
					{3, 4},
				},
				Resources: flux.ResourceManagement{ConcurrencyQuota: 1, MemoryBytesQuota: math.MaxInt64},
				Now:       nowFn(),
			}),
		},
		{
			name: "multiple files - splitting options setting",
			files: []string{
				`package main
import pl "planner"

option pl.disablePhysicalRules = ["fromRangeRule"]

from(bucket: "bkt") |> range(start: 0) |> filter(fn: (r) => r._value > 0) |> count()`,
				`package foo
import "planner"

option planner.disableLogicalRules = ["removeCountRule"]`},
			want: plantest.CreatePlanSpec(&plantest.PlanSpec{
				Nodes: []plan.Node{
					&plan.PhysicalPlanNode{Spec: &influxdb.FromProcedureSpec{}},
					&plan.PhysicalPlanNode{Spec: &universe.RangeProcedureSpec{}},
					&plan.PhysicalPlanNode{Spec: &universe.FilterProcedureSpec{}},
					&plan.PhysicalPlanNode{Spec: &universe.CountProcedureSpec{}},
					&plan.PhysicalPlanNode{Spec: &universe.YieldProcedureSpec{}},
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
					{2, 3},
					{3, 4},
				},
				Resources: flux.ResourceManagement{ConcurrencyQuota: 1, MemoryBytesQuota: math.MaxInt64},
				Now:       nowFn(),
			}),
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			if strings.HasPrefix(tc.name, "multiple files") {
				t.Skip("how should options behave with multiple files?")
			}
			if len(tc.files) == 0 {
				t.Fatal("the test should have at least one file")
			}
			astPkg, err := flux.Parse(tc.files[0])
			if err != nil {
				t.Fatal(err)
			}

			if len(tc.files) > 1 {
				for _, file := range tc.files[1:] {
					otherPkg, err := flux.Parse(file)
					if err != nil {
						t.Fatal(err)
					}
					astPkg.Files = append(astPkg.Files, otherPkg.Files...)
				}
			}

			program := lang.CompileAST(astPkg, nowFn())
			program.SetExecutorDependencies(executetest.NewTestExecuteDependencies())

			if _, err := program.Start(context.Background(), &memory.Allocator{}); err != nil {
				if tc.wantErr == "" {
					t.Fatalf("failed to start program: %v", err)
				} else if got := getRootErr(err); tc.wantErr != got.Error() {
					t.Fatalf("expected wrong error -want/+got:\n\t- %s\n\t+ %s", tc.wantErr, got)
				}
				return
			} else if tc.wantErr != "" {
				t.Fatalf("expected error, got none")
			}

			if err := plantest.ComparePlansShallow(tc.want, program.PlanSpec); err != nil {
				t.Errorf("unexpected plans: %v", err)
			}
		})
	}
}

func getRootErr(err error) error {
	if err == nil {
		return err
	}
	fe, ok := err.(*flux.Error)
	if !ok {
		return err
	}
	if fe == nil {
		return fe
	}
	if fe.Err == nil {
		return fe
	}
	return getRootErr(fe.Err)
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

	vs, _, err := flux.Eval(context.Background(), dependenciestest.Default(), script)
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

	if p, ok := program.(lang.DependenciesAwareProgram); ok {
		p.SetExecutorDependencies(executetest.NewTestExecuteDependencies())
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

func getTableObjects(vs []interpreter.SideEffect) []*flux.TableObject {
	tos := make([]*flux.TableObject, 0)
	for _, v := range vs {
		if to, ok := v.Value.(*flux.TableObject); ok {
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
