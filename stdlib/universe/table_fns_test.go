package universe_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/dependencies/dependenciestest"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/lang/langtest"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/stdlib/universe"
	"github.com/influxdata/flux/values"
	"github.com/influxdata/flux/values/objects"
)

var prelude = `
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

inj = csv.from(csv: data)

`

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

func evalOrFail(t *testing.T, script string) values.Scope {
	t.Helper()

	ctx := dependenciestest.Default().Inject(context.Background())
	ctx = langtest.DefaultExecutionDependencies().Inject(ctx)
	_, s, err := runtime.Eval(ctx, script)
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
		wantErr      error
		omitExecDeps bool
	}{
		{
			name: "exactly one match 1", // first table
			want: func() flux.Table {
				want := &executetest.Table{
					KeyCols: []string{"_measurement", "user"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "_measurement", Type: flux.TString},
						{Label: "user", Type: flux.TString},
					},
					Data: [][]interface{}{
						{mustParseTime(`2018-05-22T19:53:26.000000000Z`), 0.0, "CPU", "user1"},
						{mustParseTime(`2018-05-22T19:53:36.000000000Z`), 1.0, "CPU", "user1"},
					},
				}
				want.Normalize()
				return want
			}(),
			fn: `f = (key) => key.user == "user1" and key._measurement == "CPU"`,
		},
		{
			name: "exactly one match 2", // second table
			want: func() flux.Table {
				want := &executetest.Table{
					KeyCols: []string{"_measurement", "user"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "_measurement", Type: flux.TString},
						{Label: "user", Type: flux.TString},
					},
					Data: [][]interface{}{
						{mustParseTime(`2018-05-22T19:53:26.000000000Z`), 4.0, "CPU", "user2"},
						{mustParseTime(`2018-05-22T19:53:36.000000000Z`), 20.0, "CPU", "user2"},
						{mustParseTime(`2018-05-22T19:53:46.000000000Z`), 7.0, "CPU", "user2"},
					},
				}
				want.Normalize()
				return want
			}(),
			fn: `f = (key) => key.user == "user2"`,
		},
		{
			name: "exactly one match 3", // third table
			want: func() flux.Table {
				want := &executetest.Table{
					KeyCols: []string{"_measurement", "user"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "_measurement", Type: flux.TString},
						{Label: "user", Type: flux.TString},
					},
					Data: [][]interface{}{
						{mustParseTime(`2018-05-22T19:53:26.000000000Z`), 1.0, "RAM", "user1"},
					},
				}
				want.Normalize()
				return want
			}(),
			fn: `f = (key) => key.user == "user1" and key._measurement == "RAM"`,
		},
		{
			name: "multiple match", // first and third
			want: func() flux.Table {
				want := &executetest.Table{
					KeyCols: []string{"_measurement", "user"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "_measurement", Type: flux.TString},
						{Label: "user", Type: flux.TString},
					},
					Data: [][]interface{}{
						{mustParseTime(`2018-05-22T19:53:26.000000000Z`), 0.0, "CPU", "user1"},
						{mustParseTime(`2018-05-22T19:53:36.000000000Z`), 1.0, "CPU", "user1"},
					},
				}
				want.Normalize()
				return want
			}(),
			fn: `f = (key) => key.user == "user1"`,
		},
		{
			name:    "no match",
			wantErr: fmt.Errorf("no table found"),
			fn:      `f = (key) => key.user == "no-user"`,
		},
		{
			name:         "no execution context", // notifying the user of no-execution context
			wantErr:      fmt.Errorf("do not have an execution context for tableFind, if using the repl, try executing this code on the server using the InfluxDB API"),
			fn:           `f = (key) => key.user == "user1" and key._measurement == "CPU"`,
			omitExecDeps: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := dependenciestest.Default().Inject(context.Background())
			if !tc.omitExecDeps {
				ctx = langtest.DefaultExecutionDependencies().Inject(ctx)
			}
			_, scope, err := runtime.Eval(ctx, prelude)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			to, ok := scope.Lookup("inj")
			if !ok {
				t.Fatal("unable to find input in prelude script")
			}

			_, scope, err = runtime.Eval(ctx, tc.fn)
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
	script := prelude + `
t = inj |> tableFind(fn: (key) => key.user == "user1")`

	s := evalOrFail(t, script)
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
	want := values.NewArrayWithBacking(semantic.NewArrayType(semantic.BasicString), []values.Value{values.New("user1"), values.New("user1")})

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
		t.Fatal("expected error got none")
	}

	wantErr := "cannot find column idk"
	if diff := cmp.Diff(wantErr, err.Error()); diff != "" {
		t.Errorf("expected different error -want/+got:\n%s\n", diff)
	}
}

func TestGetRecord_Call(t *testing.T) {
	script := prelude + `
t = inj |> tableFind(fn: (key) => key.user == "user1")`

	s := evalOrFail(t, script)
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

var nilTableFindBase = `
import "csv"

data = "
#group,false,false,false,false,false,false,true
#datatype,string,long,double,string,string,string,string
#default,_result,,,,,,
,result,table,_value,_field,_measurement,cpu,host
,,0,1,usage_user,cpu,cpu-total,influx.local
,,0,2,usage_user,cpu,cpu0,influx.local
,,0,3,usage_user,cpu,cpu1,influx.local
,,0,4,usage_user,cpu,cpu10,influx.local
,,0,5,usage_user,cpu,cpu11,influx.local
"

base = csv.from( csv: data )

// Intentionally fail the table find
find_no = (key) => key.host == "influx.local-NOT-FOUND"
find_yes = (key) => key.host == "influx.local"
`

func TestFindColumn_NoTable(t *testing.T) {
	script := nilTableFindBase + `
		filtered_list =
			base
				|> limit(n: 3)
				|> findColumn(fn: find_no, column: "cpu")

		ok = length( arr: filtered_list ) == 0
	`

	s := evalOrFail(t, script)

	for _, id := range []string{"ok"} {
		if !mustLookup(s, id).Bool() {
			t.Errorf("%s was not OK indeed", id)
		}
	}
}

func TestFindColumn_BadColumn(t *testing.T) {
	script := nilTableFindBase + `
		filtered_list =
			base
				|> limit(n: 3)
				|> findColumn(fn: find_yes, column: "cpu-NOT-FOUND")

		ok = length( arr: filtered_list ) == 0
	`

	s := evalOrFail(t, script)

	for _, id := range []string{"ok"} {
		if !mustLookup(s, id).Bool() {
			t.Errorf("%s was not OK indeed", id)
		}
	}
}

func TestFindRecord_NoTable(t *testing.T) {
	script := nilTableFindBase + `
		filtered_object =
			base |> findRecord(fn: find_no, idx: 0)

		ok = not exists filtered_object.cpu
	`

	s := evalOrFail(t, script)

	for _, id := range []string{"ok"} {
		if !mustLookup(s, id).Bool() {
			t.Errorf("%s was not OK indeed", id)
		}
	}
}

func TestFindRecord_BadIdx(t *testing.T) {
	script := nilTableFindBase + `
		filtered_object =
			base |> findRecord(fn: find_yes, idx: 1000)

		ok = not exists filtered_object.cpu
	`

	s := evalOrFail(t, script)

	for _, id := range []string{"ok"} {
		if !mustLookup(s, id).Bool() {
			t.Errorf("%s was not OK indeed", id)
		}
	}
}

// We have to write this test as a non-standard e2e test, because
// our framework doesn't allow comparison between "simple" values, but only streams of tables.
func TestIndexFns_EndToEnd(t *testing.T) {
	// TODO(affo): uncomment schema-testing lines (in the `script` too) once we decide to expose the schema.
	script := prelude + `
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

	s := evalOrFail(t, script)

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
