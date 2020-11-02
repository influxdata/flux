package lang_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	_ "github.com/influxdata/flux/builtin"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/lang"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/runtime"
)

func runQuery(script string) (flux.Query, error) {
	program, err := lang.Compile(script, runtime.Default, time.Unix(0, 0))
	if err != nil {
		return nil, err
	}
	ctx := executetest.NewTestExecuteDependencies().Inject(context.Background())
	q, err := program.Start(ctx, &memory.Allocator{})
	if err != nil {
		return nil, err
	}
	return q, nil
}

var validScript = `
import "csv"

data = "
#datatype,string,long,long,string
#group,false,false,false,true
#default,_result,,,
,result,table,value,tag
,,0,10,a
,,0,10,a
,,1,20,b
,,1,20,b
,,2,30,c
,,2,30,c
,,3,40,d
,,3,40,d
"

csv.from(csv: data) |> yield(name: "res")`

func TestQuery_Results(t *testing.T) {
	q, err := runQuery(validScript)
	if err != nil {
		t.Fatalf("unexpected error while creating query: %s", err)
	}

	// gather counts
	var resCount, tableCount, totRows int
	for res := range q.Results() {
		resCount++
		if err := res.Tables().Do(func(tbl flux.Table) error {
			tableCount++
			return tbl.Do(func(cr flux.ColReader) error {
				tags := cr.Strings(1)
				for i := 0; i < tags.Len(); i++ {
					totRows++
				}
				return nil
			})
		}); err != nil {
			t.Fatalf("unexpected error while iterating over tables: %s", err)
		}
	}

	// com,pare expected counts
	if resCount != 1 {
		t.Errorf("got %d results instead of %d", resCount, 1)
	}
	if tableCount != 4 {
		t.Errorf("got %d tables instead of %d", tableCount, 4)
	}
	if totRows != 8 {
		t.Errorf("got %d rows instead of %d", totRows, 8)
	}

	// release query resources
	q.Done()

	if q.Err() != nil {
		t.Fatalf("unexpected error from query execution: %s", q.Err())
	}
}

func TestQuery_Stats(t *testing.T) {
	t.Skip("stats are updated by the controller, running a standalone query won't update them")

	q, err := runQuery(validScript)
	if err != nil {
		t.Fatalf("unexpected error while creating query: %s", err)
	}

	// consume results
	for res := range q.Results() {
		if err := res.Tables().Do(func(tbl flux.Table) error {
			return tbl.Do(func(cr flux.ColReader) error {
				return nil
			})
		}); err != nil {
			t.Fatalf("unexpected error while iterating over tables: %s", err)
		}
	}

	stats := q.Statistics()
	if stats.TotalDuration <= 0 {
		t.Errorf("unexpected total duration: %v", stats.CompileDuration)
	}
	if stats.CompileDuration <= 0 {
		t.Errorf("unexpected compile duration: %v", stats.CompileDuration)
	}
	if stats.QueueDuration <= 0 {
		t.Errorf("unexpected queue duration: %v", stats.CompileDuration)
	}
	if stats.PlanDuration <= 0 {
		t.Errorf("unexpected plan duration: %v", stats.CompileDuration)
	}
	if stats.ExecuteDuration <= 0 {
		t.Errorf("unexpected execute duration: %v", stats.CompileDuration)
	}
	if stats.Concurrency <= 0 {
		t.Errorf("unexpected concurrency: %v", stats.CompileDuration)
	}
	if stats.MaxAllocated <= 0 {
		t.Errorf("unexpected max allocated: %v", stats.CompileDuration)
	}
}

func TestQuery_RuntimeError(t *testing.T) {
	var invalidScript = `
import "csv"

data = "
#datatype,string,long,long,string
#group,false,false,false,true
#default,_result,,,
,result,table,value,tag
,,0,10,a
"

csv.from(csv: data) |> map(fn: (r) => r.nonexistent)`

	q, err := runQuery(invalidScript)
	if err != nil {
		t.Fatalf("unexpected error while creating query: %s", err)
	}

	// consume and check for error
	for res := range q.Results() {
		if err := res.Tables().Do(func(tbl flux.Table) error {
			return tbl.Do(func(cr flux.ColReader) error {
				// does nothing
				return nil
			})
		}); err == nil {
			t.Fatal("expected error from accessing wrong property, got none")
		}
	}

	if q.Err() != nil {
		t.Fatalf("unexpected error from query execution: %s", q.Err())
	}
}

// This test verifies that when a query involves table functions or chain(), the plan nodes
// the main query generates does not reuse the node IDs that are already used by the table
// functions or chain()
func TestPlanNodeUniqueness(t *testing.T) {
	prelude := `
import "experimental/array"
import "experimental"

data = array.from(rows: [{
_measurement: "command",
_field: "id",
_time: 2018-12-19T22:13:30Z,
_value: 12,
}, {
_measurement: "command",
_field: "id",
_time: 2018-12-19T22:13:40Z,
_value: 23,
}, {
_measurement: "command",
_field: "id",
_time: 2018-12-19T22:13:50Z,
_value: 34,
}, {
_measurement: "command",
_field: "guild",
_time: 2018-12-19T22:13:30Z,
_value: 12,
}, {
_measurement: "command",
_field: "guild",
_time: 2018-12-19T22:13:40Z,
_value: 23,
}, {
_measurement: "command",
_field: "guild",
_time: 2018-12-19T22:13:50Z,
_value: 34,
}])
`
	tcs := []struct {
		name   string
		script string
		want   string
	}{
		{
			name: "chain",
			script: `
id = data
|> range(start: 2018-12-19T22:13:30Z, stop: 2018-12-19T22:14:21Z)
|> filter(fn: (r) => r["_field"] == "id")

guild = data
|> range(start: 2018-12-19T22:13:30Z, stop: 2018-12-19T22:14:21Z)
|> filter(fn: (r) => r["_field"] == "guild")

experimental.chain(first: id, second: guild)
`,
			want: `[digraph {
  experimental/array.from0
  range1
  filter2
  // r._field == "id"
  generated_yield

  experimental/array.from0 -> range1
  range1 -> filter2
  filter2 -> generated_yield
}
 digraph {
  experimental/array.from3
  range4
  filter5
  // r._field == "guild"
  generated_yield

  experimental/array.from3 -> range4
  range4 -> filter5
  filter5 -> generated_yield
}
]`,
		},
		{
			name: "tableFns",
			script: `
ids = data
|> range(start: 2018-12-19T22:13:30Z, stop: 2018-12-19T22:14:21Z)
|> filter(fn: (r) => r["_field"] == "id")
|> sort()
|> tableFind(fn: (key) => true)
|> getColumn(column: "_field")

id = ids[0]

data
|> range(start: 2018-12-19T22:13:30Z, stop: 2018-12-19T22:14:21Z)
|> filter(fn: (r) => r["_field"] == id)
`,
			want: `[digraph {
  experimental/array.from0
  range1
  filter2
  // r._field == "id"
  sort3
  generated_yield

  experimental/array.from0 -> range1
  range1 -> filter2
  filter2 -> sort3
  sort3 -> generated_yield
}
 digraph {
  experimental/array.from4
  range5
  filter6
  // r._field == "id"
  generated_yield

  experimental/array.from4 -> range5
  range5 -> filter6
  filter6 -> generated_yield
}
]`,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			if q, err := runQuery(prelude + tc.script); err != nil {
				t.Error(err)
			} else {
				got := fmt.Sprintf("%v", q.Statistics().Metadata["flux/query-plan"])
				if !cmp.Equal(tc.want, got) {
					t.Errorf("unexpected value -want/+got\n%s", cmp.Diff(tc.want, got))
				}
			}
		})
	}
}
