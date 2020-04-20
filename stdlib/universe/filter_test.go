package universe_test

import (
	"context"
	"testing"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/dependencies/dependenciestest"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/internal/gen"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/plan/plantest"
	"github.com/influxdata/flux/querytest"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/stdlib/influxdata/influxdb"
	"github.com/influxdata/flux/stdlib/universe"
	"github.com/influxdata/flux/values"
	"github.com/influxdata/flux/values/valuestest"
)

func TestFilter_NewQuery(t *testing.T) {
	tests := []querytest.NewQueryTestCase{
		{
			Name: "from with database filter and range",
			Raw:  `from(bucket:"mybucket") |> filter(fn: (r) => r["t1"]=="val1" and r["t2"]=="val2") |> range(start:-4h, stop:-2h) |> count()`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: influxdb.NameOrID{Name: "mybucket"},
						},
					},
					{
						ID: "filter1",
						Spec: &universe.FilterOpSpec{
							Fn: interpreter.ResolvedFunction{
								Fn:    executetest.FunctionExpression(t, `(r) => r["t1"] == "val1" and r["t2"] == "val2"`),
								Scope: valuestest.Scope(),
							},
						},
					},
					{
						ID: "range2",
						Spec: &universe.RangeOpSpec{
							Start: flux.Time{
								Relative:   -4 * time.Hour,
								IsRelative: true,
							},
							Stop: flux.Time{
								Relative:   -2 * time.Hour,
								IsRelative: true,
							},
							TimeColumn:  "_time",
							StartColumn: "_start",
							StopColumn:  "_stop",
						},
					},
					{
						ID: "count3",
						Spec: &universe.CountOpSpec{
							AggregateConfig: execute.DefaultAggregateConfig,
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "filter1"},
					{Parent: "filter1", Child: "range2"},
					{Parent: "range2", Child: "count3"},
				},
			},
		},
		{
			Name: "from with database filter (and with or) and range",
			Raw: `from(bucket:"mybucket")
						|> filter(fn: (r) =>
								(
									(r["t1"]=="val1")
									and
									(r["t2"]=="val2")
								)
								or
								(r["t3"]=="val3")
							)
						|> range(start:-4h, stop:-2h)
						|> count()`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: influxdb.NameOrID{Name: "mybucket"},
						},
					},
					{
						ID: "filter1",
						Spec: &universe.FilterOpSpec{
							Fn: interpreter.ResolvedFunction{
								Fn:    executetest.FunctionExpression(t, `(r) => (r["t1"] == "val1" and r["t2"] == "val2") or r["t3"] == "val3"`),
								Scope: valuestest.Scope(),
							},
						},
					},
					{
						ID: "range2",
						Spec: &universe.RangeOpSpec{
							Start: flux.Time{
								Relative:   -4 * time.Hour,
								IsRelative: true,
							},
							Stop: flux.Time{
								Relative:   -2 * time.Hour,
								IsRelative: true,
							},
							TimeColumn:  "_time",
							StartColumn: "_start",
							StopColumn:  "_stop",
						},
					},
					{
						ID: "count3",
						Spec: &universe.CountOpSpec{
							AggregateConfig: execute.DefaultAggregateConfig,
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "filter1"},
					{Parent: "filter1", Child: "range2"},
					{Parent: "range2", Child: "count3"},
				},
			},
		},
		{
			Name: "from with database filter including fields",
			Raw: `from(bucket:"mybucket")
						|> filter(fn: (r) =>
							(r["t1"] =="val1")
							and
							(r["_field"] == 10)
						)
						|> range(start:-4h, stop:-2h)
						|> count()`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: influxdb.NameOrID{Name: "mybucket"},
						},
					},
					{
						ID: "filter1",
						Spec: &universe.FilterOpSpec{
							Fn: interpreter.ResolvedFunction{
								Fn:    executetest.FunctionExpression(t, `(r) => r["t1"] == "val1" and r["_field"] == 10`),
								Scope: valuestest.Scope(),
							},
						},
					},
					{
						ID: "range2",
						Spec: &universe.RangeOpSpec{
							Start: flux.Time{
								Relative:   -4 * time.Hour,
								IsRelative: true,
							},
							Stop: flux.Time{
								Relative:   -2 * time.Hour,
								IsRelative: true,
							},
							TimeColumn:  "_time",
							StartColumn: "_start",
							StopColumn:  "_stop",
						},
					},
					{
						ID: "count3",
						Spec: &universe.CountOpSpec{
							AggregateConfig: execute.DefaultAggregateConfig,
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "filter1"},
					{Parent: "filter1", Child: "range2"},
					{Parent: "range2", Child: "count3"},
				},
			},
		},
		{
			Name: "from with database filter with no parens including fields",
			Raw: `from(bucket:"mybucket")
						|> filter(fn: (r) =>
							r["t1"]=="val1"
							and
							r["_field"] == 10
						)
						|> range(start:-4h, stop:-2h)
						|> count()`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: influxdb.NameOrID{Name: "mybucket"},
						},
					},
					{
						ID: "filter1",
						Spec: &universe.FilterOpSpec{
							Fn: interpreter.ResolvedFunction{
								Fn:    executetest.FunctionExpression(t, `(r) => r["t1"] == "val1" and r["_field"] == 10`),
								Scope: valuestest.Scope(),
							},
						},
					},
					{
						ID: "range2",
						Spec: &universe.RangeOpSpec{
							Start: flux.Time{
								Relative:   -4 * time.Hour,
								IsRelative: true,
							},
							Stop: flux.Time{
								Relative:   -2 * time.Hour,
								IsRelative: true,
							},
							TimeColumn:  "_time",
							StartColumn: "_start",
							StopColumn:  "_stop",
						},
					},
					{
						ID: "count3",
						Spec: &universe.CountOpSpec{
							AggregateConfig: execute.DefaultAggregateConfig,
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "filter1"},
					{Parent: "filter1", Child: "range2"},
					{Parent: "range2", Child: "count3"},
				},
			},
		},
		{
			Name: "from with database filter with no parens including regex and field",
			Raw: `from(bucket:"mybucket")
						|> filter(fn: (r) =>
							r["t1"]=~/^val1/
							and
							r["_field"] == 10.5
						)
						|> range(start:-4h, stop:-2h)
						|> count()`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: influxdb.NameOrID{Name: "mybucket"},
						},
					},
					{
						ID: "filter1",
						Spec: &universe.FilterOpSpec{
							Fn: interpreter.ResolvedFunction{
								Fn:    executetest.FunctionExpression(t, `(r) => r["t1"] =~ /^val1/ and r["_field"] == 10.5`),
								Scope: valuestest.Scope(),
							},
						},
					},
					{
						ID: "range2",
						Spec: &universe.RangeOpSpec{
							Start: flux.Time{
								Relative:   -4 * time.Hour,
								IsRelative: true,
							},
							Stop: flux.Time{
								Relative:   -2 * time.Hour,
								IsRelative: true,
							},
							TimeColumn:  "_time",
							StartColumn: "_start",
							StopColumn:  "_stop",
						},
					},
					{
						ID: "count3",
						Spec: &universe.CountOpSpec{
							AggregateConfig: execute.DefaultAggregateConfig,
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "filter1"},
					{Parent: "filter1", Child: "range2"},
					{Parent: "range2", Child: "count3"},
				},
			},
		},
		{
			Name: "from with database regex with escape",
			Raw: `from(bucket:"mybucket")
						|> filter(fn: (r) =>
							r["t1"]=~/^va\/l1/
						)`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: influxdb.NameOrID{Name: "mybucket"},
						},
					},
					{
						ID: "filter1",
						Spec: &universe.FilterOpSpec{
							Fn: interpreter.ResolvedFunction{
								Fn:    executetest.FunctionExpression(t, `(r) => r["t1"] =~ /^va\/l1/`),
								Scope: valuestest.Scope(),
							},
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "filter1"},
				},
			},
		},
		{
			Name: "from with database with two regex",
			Raw: `from(bucket:"mybucket")
						|> filter(fn: (r) =>
							r["t1"]=~/^va\/l1/
							and
							r["t2"] !~ /^val2/
						)`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: influxdb.NameOrID{Name: "mybucket"},
						},
					},
					{
						ID: "filter1",
						Spec: &universe.FilterOpSpec{
							Fn: interpreter.ResolvedFunction{
								Fn:    executetest.FunctionExpression(t, `(r) => r["t1"] =~ /^va\/l1/ and r["t2"] !~ /^val2/`),
								Scope: valuestest.Scope(),
							},
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "filter1"},
				},
			},
		},
		{
			Name: "from with drop",
			Raw:  `from(bucket:"mybucket") |> filter(fn: (r) => r._value > 0.0, onEmpty: "drop")`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: influxdb.NameOrID{Name: "mybucket"},
						},
					},
					{
						ID: "filter1",
						Spec: &universe.FilterOpSpec{
							Fn: interpreter.ResolvedFunction{
								Fn:    executetest.FunctionExpression(t, `(r) => r._value > 0.0`),
								Scope: valuestest.Scope(),
							},
							OnEmpty: "drop",
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "filter1"},
				},
			},
		},
		{
			Name: "from with keep",
			Raw:  `from(bucket:"mybucket") |> filter(fn: (r) => r._value > 0.0, onEmpty: "keep")`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: influxdb.NameOrID{Name: "mybucket"},
						},
					},
					{
						ID: "filter1",
						Spec: &universe.FilterOpSpec{
							Fn: interpreter.ResolvedFunction{
								Fn:    executetest.FunctionExpression(t, `(r) => r._value > 0.0`),
								Scope: valuestest.Scope(),
							},
							OnEmpty: "keep",
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "filter1"},
				},
			},
		},
		{
			Name:    "from with invalid parameter",
			Raw:     `from(bucket:"mybucket") |> filter(fn: (r) => true, onEmpty: "invalid")`,
			WantErr: true,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			querytest.NewQueryTestHelper(t, tc)
		})
	}
}

func TestMergeFilterAnyRule(t *testing.T) {
	var (
		from        = &influxdb.FromProcedureSpec{}
		count       = &universe.CountProcedureSpec{}
		filterOther = &universe.FilterProcedureSpec{
			Fn: interpreter.ResolvedFunction{
				Fn: executetest.FunctionExpression(t, `() => "foo"`),
			},
		}
		filterTrue = &universe.FilterProcedureSpec{
			Fn: interpreter.ResolvedFunction{
				Fn: executetest.FunctionExpression(t, `() => true`),
			},
		}
		filterFalse = &universe.FilterProcedureSpec{
			Fn: interpreter.ResolvedFunction{
				Fn: executetest.FunctionExpression(t, `() => false`),
			},
		}
	)

	tests := []plantest.RuleTestCase{
		{
			Name: "filterOther",
			// from -> filter => from -> filter
			Rules: []plan.Rule{universe.RemoveTrivialFilterRule{}},
			Before: &plantest.PlanSpec{
				Nodes: []plan.Node{
					plan.CreatePhysicalNode("from", from),
					plan.CreatePhysicalNode("filter", filterOther),
				},
				Edges: [][2]int{{0, 1}},
			},
			NoChange: true,
		},
		{
			Name: "filterFalse",
			// from -> filter => from -> filter
			Rules: []plan.Rule{universe.RemoveTrivialFilterRule{}},
			Before: &plantest.PlanSpec{
				Nodes: []plan.Node{
					plan.CreatePhysicalNode("from", from),
					plan.CreatePhysicalNode("filter", filterFalse),
				},
				Edges: [][2]int{{0, 1}},
			},
			NoChange: true,
		},
		{
			Name: "filterTrue",
			// from -> filter => from
			Rules: []plan.Rule{universe.RemoveTrivialFilterRule{}},
			Before: &plantest.PlanSpec{
				Nodes: []plan.Node{
					plan.CreatePhysicalNode("from", from),
					plan.CreatePhysicalNode("filter", filterTrue),
				},
				Edges: [][2]int{{0, 1}},
			},
			After: &plantest.PlanSpec{
				Nodes: []plan.Node{
					plan.CreatePhysicalNode("from", from),
				},
			},
		},
		{
			Name: "count filterTrue",
			// count -> filter => count
			Rules: []plan.Rule{universe.RemoveTrivialFilterRule{}},
			Before: &plantest.PlanSpec{
				Nodes: []plan.Node{
					plan.CreatePhysicalNode("count", count),
					plan.CreatePhysicalNode("filter", filterTrue),
				},
				Edges: [][2]int{{0, 1}},
			},
			After: &plantest.PlanSpec{
				Nodes: []plan.Node{
					plan.CreatePhysicalNode("count", count),
				},
			},
		},
		{
			Name: "from filterTrue count",
			// from -> filter -> count => from -> count
			Rules: []plan.Rule{universe.RemoveTrivialFilterRule{}},
			Before: &plantest.PlanSpec{
				Nodes: []plan.Node{
					plan.CreatePhysicalNode("from", from),
					plan.CreatePhysicalNode("filter", filterTrue),
					plan.CreatePhysicalNode("count", count),
				},
				Edges: [][2]int{{0, 1}, {1, 2}},
			},
			After: &plantest.PlanSpec{
				Nodes: []plan.Node{
					plan.CreatePhysicalNode("from", from),
					plan.CreatePhysicalNode("count", count),
				},
				Edges: [][2]int{{0, 1}},
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			plantest.PhysicalRuleTestHelper(t, &tc)
		})
	}
}

func TestFilter_Process(t *testing.T) {
	testCases := []struct {
		name string
		spec *universe.FilterProcedureSpec
		data []flux.Table
		want []*executetest.Table
	}{
		{
			name: `_value>5`,
			spec: &universe.FilterProcedureSpec{
				Fn: interpreter.ResolvedFunction{
					Fn:    executetest.FunctionExpression(t, `(r) => r._value > 5.0`),
					Scope: valuestest.Scope(),
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 1.0},
					{execute.Time(2), 6.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(2), 6.0},
				},
			}},
		},
		{
			name: "_value>5 multiple blocks",
			spec: &universe.FilterProcedureSpec{
				Fn: interpreter.ResolvedFunction{
					Fn:    executetest.FunctionExpression(t, `(r) => r._value > 5.0`),
					Scope: valuestest.Scope(),
				},
			},
			data: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"t1"},
					ColMeta: []flux.ColMeta{
						{Label: "t1", Type: flux.TString},
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{"a", execute.Time(1), 3.0},
						{"a", execute.Time(2), 6.0},
						{"a", execute.Time(2), 1.0},
					},
				},
				&executetest.Table{
					KeyCols: []string{"t1"},
					ColMeta: []flux.ColMeta{
						{Label: "t1", Type: flux.TString},
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{"b", execute.Time(3), 3.0},
						{"b", execute.Time(3), 2.0},
						{"b", execute.Time(4), 8.0},
					},
				},
			},
			want: []*executetest.Table{
				{
					KeyCols: []string{"t1"},
					ColMeta: []flux.ColMeta{
						{Label: "t1", Type: flux.TString},
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{"a", execute.Time(2), 6.0},
					},
				},
				{
					KeyCols: []string{"t1"},
					ColMeta: []flux.ColMeta{
						{Label: "t1", Type: flux.TString},
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{"b", execute.Time(4), 8.0},
					},
				},
			},
		},
		{
			name: "_value>5 and t1 = a and t2 = y",
			spec: &universe.FilterProcedureSpec{
				Fn: interpreter.ResolvedFunction{
					Fn:    executetest.FunctionExpression(t, `(r) => r._value > 5.0 and r.t1 == "a" and r.t2 == "y"`),
					Scope: valuestest.Scope(),
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "t1", Type: flux.TString},
					{Label: "t2", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(1), 1.0, "a", "x"},
					{execute.Time(2), 6.0, "a", "x"},
					{execute.Time(3), 8.0, "a", "y"},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "t1", Type: flux.TString},
					{Label: "t2", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(3), 8.0, "a", "y"},
				},
			}},
		},
		{
			name: `_value>5 with unused nulls`,
			spec: &universe.FilterProcedureSpec{
				Fn: interpreter.ResolvedFunction{
					Fn:    executetest.FunctionExpression(t, `(r) => r._value > 5.0`),
					Scope: valuestest.Scope(),
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "host", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(1), 1.0, "server01"},
					{execute.Time(2), 1.0, nil},
					{execute.Time(3), 6.0, "server02"},
					{execute.Time(4), 6.0, nil},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "host", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(3), 6.0, "server02"},
					{execute.Time(4), 6.0, nil},
				},
			}},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			executetest.ProcessTestHelper2(
				t,
				tc.data,
				tc.want,
				nil,
				func(id execute.DatasetID, alloc *memory.Allocator) (execute.Transformation, execute.Dataset) {
					ctx := dependenciestest.Default().Inject(context.Background())
					tx, d, err := universe.NewFilterTransformation(ctx, tc.spec, id, alloc)
					if err != nil {
						t.Fatal(err)
					}
					return tx, d
				},
			)
		})
	}
}

func BenchmarkFilter_Values(b *testing.B) {
<<<<<<< HEAD
	b.Run("1000", func(b *testing.B) {
		fn := executetest.FunctionExpression(b, `(r) => r._value > 0.0`)
		benchmarkFilter(b, 1000, fn)
=======
	b.Run("500", func(b *testing.B) {
		benchmarkFilter(b, 500, &semantic.FunctionExpression{
			Block: &semantic.FunctionBlock{
				Parameters: &semantic.FunctionParameters{
					List: []*semantic.FunctionParameter{
						{Key: &semantic.Identifier{Name: "r"}},
					},
				},
				Body: &semantic.BinaryExpression{
					Operator: ast.GreaterThanEqualOperator,
					Left: &semantic.MemberExpression{
						Object:   &semantic.IdentifierExpression{Name: "r"},
						Property: "_value",
					},
					Right: &semantic.FloatLiteral{Value: 0.0},
				},
			},
		})
>>>>>>> master
	})
}

func benchmarkFilter(b *testing.B, n int, fn *semantic.FunctionExpression) {
	b.ReportAllocs()
	spec := &universe.FilterProcedureSpec{
		Fn: interpreter.ResolvedFunction{
			Fn:    fn,
			Scope: values.NewScope(),
		},
	}
	executetest.ProcessBenchmarkHelper(b,
		func(alloc *memory.Allocator) (flux.TableIterator, error) {
			schema := gen.Schema{
				NumPoints: n,
				Alloc:     alloc,
				Tags: []gen.Tag{
					{Name: "_measurement", Cardinality: 1},
					{Name: "_field", Cardinality: 6},
					{Name: "t0", Cardinality: 100},
					{Name: "t1", Cardinality: 50},
				},
			}
			return gen.Input(schema)
		},
		func(id execute.DatasetID, alloc *memory.Allocator) (execute.Transformation, execute.Dataset) {
			t, d, err := universe.NewFilterTransformation(context.Background(), spec, id, alloc)
			if err != nil {
				b.Fatal(err)
			}
			return t, d
		},
	)
}
