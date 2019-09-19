package universe_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/dependencies/dependenciestest"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/lang/langtest"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/stdlib/universe"
	"github.com/influxdata/flux/values"
	"github.com/influxdata/flux/values/objects"
)

var (
	to     *flux.TableObject
	tables []flux.Table
)

func init() {
	script := `
import "csv"

data = "#datatype,string,long,dateTime:RFC3339,double,string,string
#group,false,false,false,false,true,true
#default,_result,,,,,
,result,table,_time,_value,_measurement,user
,,0,2018-05-22T19:53:26Z,0,CPU,user1
,,0,2018-05-22T19:53:36Z,1,CPU,user1
,,1,2018-05-22T19:53:26Z,4,CPU,user2
,,1,2018-05-22T19:53:36Z,20,CPU,user2
,,1,2018-05-22T19:53:46Z,7,CPU,user2
,,2,2018-05-22T19:53:26Z,1,RAM,user1
"

csv.from(csv: data)`
	ctx := dependenciestest.Default().Inject(context.Background())
	vs, _, err := flux.Eval(ctx, script)
	if err != nil {
		panic(fmt.Errorf("cannot compile simple script to prepare test: %s", err))
	}
	for _, v := range vs {
		if v, ok := v.Value.(*flux.TableObject); ok {
			to = v
			break
		}
	}
	if to == nil {
		panic(errors.New("cannot find TableObject as result of test script"))
	}

	// init tables
	tables = make([]flux.Table, 0, 4)
	t0 := &executetest.Table{
		KeyCols: []string{"_measurement", "user"},
		ColMeta: []flux.ColMeta{
			{Label: "_time", Type: flux.TTime},
			{Label: "_value", Type: flux.TFloat},
			{Label: "_measurement", Type: flux.TString},
			{Label: "user", Type: flux.TString},
		},
		Data: [][]interface{}{
			{mustParseTime("2018-05-22T19:53:26.000000000Z"), 0.0, "CPU", "user1"},
			{mustParseTime("2018-05-22T19:53:36.000000000Z"), 1.0, "CPU", "user1"},
		},
	}
	t1 := &executetest.Table{
		KeyCols: []string{"_measurement", "user"},
		ColMeta: []flux.ColMeta{
			{Label: "_time", Type: flux.TTime},
			{Label: "_value", Type: flux.TFloat},
			{Label: "_measurement", Type: flux.TString},
			{Label: "user", Type: flux.TString},
		},
		Data: [][]interface{}{
			{mustParseTime("2018-05-22T19:53:26.000000000Z"), 4.0, "CPU", "user2"},
			{mustParseTime("2018-05-22T19:53:36.000000000Z"), 20.0, "CPU", "user2"},
			{mustParseTime("2018-05-22T19:53:46.000000000Z"), 7.0, "CPU", "user2"},
		},
	}
	t2 := &executetest.Table{
		KeyCols: []string{"_measurement", "user"},
		ColMeta: []flux.ColMeta{
			{Label: "_time", Type: flux.TTime},
			{Label: "_value", Type: flux.TFloat},
			{Label: "_measurement", Type: flux.TString},
			{Label: "user", Type: flux.TString},
		},
		Data: [][]interface{}{
			{mustParseTime("2018-05-22T19:53:26.000000000Z"), 1.0, "RAM", "user1"},
		},
	}
	t0.Normalize()
	t1.Normalize()
	t2.Normalize()
	tables = append(tables, t0, t1, t2)
}

func mustParseTime(t string) values.Time {
	if t, err := values.ParseTime(t); err != nil {
		panic(err)
	} else {
		return t
	}
}

func mustLookup(s values.Scope, valueID string) values.Value {
	v, found := s.Lookup(valueID)
	if !found {
		panic(fmt.Errorf("&%s not found in scope", valueID))
	}
	return v
}

func evalOrFail(t *testing.T, script string, mutator flux.ScopeMutator) values.Scope {
	t.Helper()

	ctx := dependenciestest.Default().Inject(context.Background())
	ctx = langtest.DefaultExecutionDependencies().Inject(ctx)
	_, s, err := flux.Eval(ctx, script, func(s values.Scope) {
		if mutator != nil {
			mutator(s)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestTableFind_Call(t *testing.T) {
	testCases := []struct {
		name string
		want flux.Table
		fn   string
		// fn      func(key values.Object) (values.Value, error)
		wantErr error
	}{
		{
			name: "exactly one match 1", // first table
			want: tables[0],
			fn:   `f = (key) => key.user == "user1" and key._measurement == "CPU"`,
		},
		{
			name: "exactly one match 2", // second table
			want: tables[1],
			fn:   `f = (key) => key.user == "user2"`,
		},
		{
			name: "exactly one match 3", // third table
			want: tables[2],
			fn:   `f = (key) => key.user == "user1" and key._measurement == "RAM"`,
		},
		{
			name: "multiple match", // first and third
			want: tables[0],
			fn:   `f = (key) => key.user == "user1"`,
		},
		{
			name:    "no match",
			wantErr: fmt.Errorf("no table found"),
			fn:      `f = (key) => key.user == "no-user"`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := dependenciestest.Default().Inject(context.Background())
			ctx = langtest.DefaultExecutionDependencies().Inject(ctx)
			_, scope, err := flux.Eval(ctx, tc.fn)
			if err != nil {
				t.Fatalf("error compiling function: %v", err)
			}

			fn, ok := scope.Lookup("f")
			if !ok {
				t.Fatal("must define a function to the f variable")
			}

			f := universe.NewTableFindFunction()
			res, err := f.Function().Call(ctx,
				values.NewObjectWithValues(map[string]values.Value{
					"tables": to,
					"fn":     fn,
				}))
			if err != nil {
				if tc.wantErr != nil {
					if diff := cmp.Diff(tc.wantErr.Error(), err.Error()); diff != "" {
						t.Errorf("expected different error -want/+got:\n%s\n", diff)
					}
					return
				}
				t.Fatal(err)
			}

			got, err := executetest.ConvertTable(res.(*objects.Table).Table())
			if err != nil {
				t.Fatal(err)
			}
			got.Normalize()

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("unexpected result -want/+got:\n%s\n", diff)
			}
		})
	}
}

func TestGetColumn_Call(t *testing.T) {
	script := `
// 'inj' is injected in the scope before evaluation
t = inj |> tableFind(fn: (key) => key.user == "user1")`

	s := evalOrFail(t, script, func(s values.Scope) {
		s.Set("inj", to)
	})
	tbl := mustLookup(s, "t")

	f := universe.NewGetColumnFunction()
	ctx := dependenciestest.Default().Inject(context.Background())
	ctx = langtest.DefaultExecutionDependencies().Inject(ctx)
	res, err := f.Function().Call(ctx,
		values.NewObjectWithValues(map[string]values.Value{
			"table":  tbl.(*objects.Table),
			"column": values.New("user"),
		}))
	if err != nil {
		t.Fatal(err)
	}

	got := res.(values.Array)
	want := values.NewArrayWithBacking(semantic.String, []values.Value{values.New("user1"), values.New("user1")})

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected result -want/+got:\n%s\n", diff)
	}

	// test for error
	f = universe.NewGetColumnFunction()
	_, err = f.Function().Call(ctx,
		values.NewObjectWithValues(map[string]values.Value{
			"table":  tbl.(*objects.Table),
			"column": values.New("idk"),
		}))
	if err == nil {
		t.Error("expected error got none")
	}

	wantErr := "cannot find column idk"
	if diff := cmp.Diff(wantErr, err.Error()); diff != "" {
		t.Errorf("expected different error -want/+got:\n%s\n", diff)
	}
}

func TestGetRecord_Call(t *testing.T) {
	script := `
// 'inj' is injected in the scope before evaluation
t = inj |> tableFind(fn: (key) => key.user == "user1")`

	s := evalOrFail(t, script, func(s values.Scope) {
		s.Set("inj", to)
	})
	tbl := mustLookup(s, "t")

	f := universe.NewGetRecordFunction()
	ctx := dependenciestest.Default().Inject(context.Background())
	ctx = langtest.DefaultExecutionDependencies().Inject(ctx)
	res, err := f.Function().Call(ctx,
		values.NewObjectWithValues(map[string]values.Value{
			"table": tbl.(*objects.Table),
			"idx":   values.New(int64(1)),
		}))
	if err != nil {
		t.Fatal(err)
	}

	got := res.(values.Object)
	want := values.NewObjectWithValues(map[string]values.Value{
		"_time":        values.New(mustParseTime("2018-05-22T19:53:36.000000000Z")),
		"_value":       values.New(1.0),
		"_measurement": values.New("CPU"),
		"user":         values.New("user1"),
	})

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected result -want/+got:\n%s\n", diff)
	}

	// test for error
	f = universe.NewGetRecordFunction()
	_, err = f.Function().Call(ctx,
		values.NewObjectWithValues(map[string]values.Value{
			"table": tbl.(*objects.Table),
			"idx":   values.New(int64(42)),
		}))
	if err == nil {
		t.Error("expected error got none")
	}

	wantErr := "index out of bounds: 42"
	if diff := cmp.Diff(wantErr, err.Error()); diff != "" {
		t.Errorf("expected different error -want/+got:\n%s\n", diff)
	}
}

// We have to write this test as a non-standard e2e test, because
// our framework doesn't allow comparison between "simple" values, but only streams of tables.
func TestIndexFns_EndToEnd(t *testing.T) {
	// TODO(affo): uncomment schema-testing lines (in the `script` too) once we decide to expose the schema.
	script := `
// 'inj' is injected in the scope before evaluation
t = inj |> tableFind(fn: (key) => key._measurement == "RAM")
c = t |> getColumn(column: "_value")
r = t |> getRecord(idx: 0)

// schemaOK0 = t.schema[0].label == "_time" and not t.schema[0].grouped
// schemaOK1 = t.schema[1].label == "_value" and not t.schema[1].grouped
// schemaOK2 = t.schema[2].label == "_measurement" and t.schema[2].grouped
// schemaOK3 = t.schema[3].label == "user" and t.schema[3].grouped
// schemaOK = schemaOK0 and schemaOK1 and schemaOK2 and schemaOK3
columnOK = c[0] == 1.0
// cannot compare directly:
// >>> unsupported binary expression {_value: float,_measurement: string,user: string,_time: time,} == {_value: float,_measurement: string,user: string,_time: time,}
recordOK = r._time == 2018-05-22T19:53:26Z and r._value == 1.0 and r._measurement == "RAM" and r.user == "user1"`

	s := evalOrFail(t, script, func(s values.Scope) {
		s.Set("inj", to)
	})

	for _, id := range []string{
		// "schemaOK",
		"columnOK",
		"recordOK",
	} {
		if !mustLookup(s, id).Bool() {
			t.Errorf("%s was not OK indeed", id)
		}
	}
}
